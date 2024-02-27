package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

var (
	K          = 0.2109375
	GoRoutines = 4
	Iterations = 100
)

const params = 2 * 6 * 64

func main() {
	var file string
	var lim int
	flag.IntVar(&GoRoutines, "c", 4, "Number of CPUs to use")
	flag.IntVar(&Iterations, "i", 100, "Number of iterations")
	flag.IntVar(&lim, "lim", 0, "Number of entries to load")
	flag.StringVar(&file, "f", "rand.txt", "File to load entries from")
	flag.Parse()

	entries, err := LoadEntries(file, 100000)
	if err != nil {
		fmt.Println(err)
		return
	}
	v := ToVector()
	optimize(entries, v)
}

type Evaluator struct {
	entries []entry
	engines []*eval.Engine
}

func NewEvaluator(entries []entry, c int) *Evaluator {
	engines := make([]*eval.Engine, c)
	for i := 0; i < c; i++ {
		engines[i] = eval.NewEvalEngine()
	}
	return &Evaluator{entries: entries, engines: engines}
}

func (e *Evaluator) E() float64 {
	type result struct {
		score  int16
		result float64
	}
	in := make(chan entry, GoRoutines)
	out := make(chan result, GoRoutines)
	var totalError float64

	go func() {
		for _, e := range e.entries {
			in <- e
		}
		close(in)
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < GoRoutines; i++ {
		wg.Add(1)
		go func(engine *eval.Engine) {
			defer wg.Done()
			for entry := range in {
				engine.Board = board.NewBoard(entry.fen)
				engine.Board.Phase = engine.Board.GetGamePhase()
				score := engine.Quiescence(context.Background(), -eval.Inf, eval.Inf, int16(engine.Board.Side))
				out <- result{score: score, result: entry.result}
			}
		}(e.engines[i])
	}

	wgOut := sync.WaitGroup{}
	wgOut.Add(1)
	go func() {
		defer wgOut.Done()
		for o := range out {
			diff := o.result - sigmoid(float64(o.score))
			totalError += diff * diff
		}
	}()
	wg.Wait()
	close(out)
	wgOut.Wait()
	return totalError / float64(len(e.entries))
}

type entry struct {
	fen    string
	result float64
}

func LoadEntries(file string, lim ...int) ([]entry, error) {
	limit := 0
	if len(lim) > 0 {
		limit = lim[0]
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	entries := make([]entry, 0)
	var scoreStr, fen string
	var result float64
	var found bool
	var count int
	for s.Scan() {
		scoreStr, fen, found = strings.Cut(s.Text(), " ")
		if !found {
			continue
		}

		result, err = strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			continue
		}

		entries = append(entries, entry{fen: fen, result: result})
		count++
		if count == limit {
			break
		}
	}

	return entries, nil
}

func sigmoid(x float64) float64 {
	return 1 / (1 + math.Pow(10, (-K*x)/400))
}

func E(entries []entry) float64 {
	type result struct {
		score  int16
		result float64
	}
	in := make(chan entry, GoRoutines)
	out := make(chan result, GoRoutines)
	var totalError float64

	go func() {
		for _, e := range entries {
			in <- e
		}
		close(in)
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < GoRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine := eval.NewEvalEngine()
			for e := range in {
				engine.Board = board.NewBoard(e.fen)
				engine.Board.Phase = engine.Board.GetGamePhase()
				score := engine.Quiescence(context.Background(), -eval.Inf, eval.Inf, int16(engine.Board.Side))
				out <- result{score: score, result: e.result}
			}
		}()
	}

	wgOut := sync.WaitGroup{}
	wgOut.Add(1)
	go func() {
		defer wgOut.Done()
		for o := range out {
			diff := o.result - sigmoid(float64(o.score))
			totalError += diff * diff
		}
	}()
	wg.Wait()
	close(out)
	wgOut.Wait()
	return totalError / float64(len(entries))
}

func optimize(entries []entry, vec [params]float64) {
	paramSpace := len(vec)
	var eUp, currE float64
	maxIter := 100
	step := 5.0
	grad := make([]float64, paramSpace)
	bestVec := vec
	evaluator := NewEvaluator(entries, GoRoutines)
	log.Println("Initial Error:", evaluator.E())
	for iter := 0; iter < maxIter; iter++ {
		now := time.Now()
		ApplyWeights(&bestVec)
		currE = evaluator.E()
		for i := 0; i < paramSpace; i++ {
			bestVec[i] += step
			ApplyWeights(&bestVec)
			eUp = evaluator.E()
			grad[i] = 100000 * step * (currE - eUp) / currE
			bestVec[i] -= step
		}

		for i := 0; i < paramSpace; i++ {
			bestVec[i] += grad[i]
		}
		ApplyWeights(&bestVec)
		log.Printf("%d/%d Error: %f (%v)\n", iter, maxIter, evaluator.E(), time.Since(now))
	}
	PrintWeights(bestVec)
}

func PSTHash() int {
	var hash int
	for stage := 0; stage < 2; stage++ {
		for color := 0; color < 2; color++ {
			for piece := 0; piece <= board.KINGS; piece++ {
				for sq := 0; sq < 64; sq++ {
					hash ^= stage * color * piece * sq * eval.PST[stage][color][piece][sq]
				}
			}
		}
	}

	return hash
}

func ToVector() [params]float64 {
	vec := [params]float64{}
	for piece := 0; piece <= board.KINGS; piece++ {
		for stage := 0; stage < 2; stage++ {
			for sq := 0; sq < 64; sq++ {
				vec[piece*128+stage*64+sq] = float64(eval.PST[stage][board.WHITE][piece][sq])
			}
		}
	}

	return vec
}

func ApplyWeights(vec *[params]float64) {
	stage := -1
	piece := -1
	for i, v := range vec {
		if i%64 == 0 {
			stage++
			stage %= 2
		}
		if i%128 == 0 {
			piece++
		}
		eval.PST[stage][board.WHITE][piece][i%64] = int(v)
	}
	eval.InitPSTs()
}

func PrintWeights(vec [params]float64) {
	var stageStr, pieceStr string
	stage := -1
	piece := -1
	weights := make([]int, 64)
	for i, v := range vec {
		if i%64 == 0 {
			stage++
			stage %= 2
		}
		if i%128 == 0 {
			piece++
		}
		weights[i%64] = int(v)
		if i%64 == 63 {
			if stage == 0 {
				stageStr = ""
			} else {
				stageStr = "EG"
			}
			switch piece {
			case 0:
				pieceStr = "pawn"
			case 1:
				pieceStr = "bishop"
			case 2:
				pieceStr = "knight"
			case 3:
				pieceStr = "rook"
			case 4:
				pieceStr = "queen"
			case 5:
				pieceStr = "king"
			}
			fmt.Printf("var %s%sPST = [64]int{\n", pieceStr, stageStr)
			for i, w := range weights {
				if i%8 == 0 {
					fmt.Print("\t")
				}
				fmt.Printf("%d, ", w)
				if i%8 == 7 {
					fmt.Println()
				}
			}
			fmt.Printf("}\n\n")
		}
	}
}

package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

var testPositions = []string{
	"startpos",
	"8/7k/8/3p4/8/1p2P3/1P2P3/7K w - - 0 1",
	"r1bq1rk1/bpp1nppp/3p1n2/p3p3/4P3/1BPP1N1P/PP1N1PP1/R1BQ1RK1 w - - 1 10",
	"r5k1/bbq2r2/1p1pR1pB/pBpP1pN1/P7/3Q2NP/5PP1/6K1 b - - 4 30",
}

func main() {
	capturePV := regexp.MustCompile(`pv (?P<pv>.*)`)
	for _, testPos := range testPositions {
		r, w, _ := os.Pipe()
		os.Stdout = w

		e := eval.NewEvalEngine()
		e.Board = board.NewBoard(testPos)
		position := e.Board.Copy()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		lr := bufio.NewScanner(r)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer w.Close()
			defer wg.Done()
			e.IDSearch(ctx, 50, false)
		}()

		for lr.Scan() {
			info := lr.Text()
			m := capturePV.FindStringSubmatch(info)
			pv := m[capturePV.SubexpIndex("pv")]
			tmpPos := position.Copy()
			pvMoves := strings.Fields(pv)

			for _, uciMove := range pvMoves {
				_, ok := tmpPos.MoveUCI(uciMove)
				if !ok {
					log.Printf("error parsing PV, position '%v' pv: %v on move: %v\n", position.ExportFEN(), info, uciMove)
					break
				}
			}
		}
		wg.Wait()
		cancel()
		r.Close()
	}

}

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/likeawizard/tofiks/pkg/uci"
	"github.com/pkg/profile"
	_ "go.uber.org/automaxprocs"
)

func main() {
	enableProfile := false
	enableMemProf := false
	flag.BoolVar(&enableProfile, "pgo", false, "Enable CPU profiling")
	flag.BoolVar(&enableMemProf, "memprof", false, "Enable memory profiling")
	flag.Parse()
	if enableProfile {
		f, err := os.Create("cmd/tofiks/default.pgo")
		if err != nil {
			log.Printf("Error creating profile file: %v", err)
			return
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Printf("Error starting profile: %v", err)
			return
		}
		defer pprof.StopCPUProfile()
	}

	if enableMemProf {
		defer profile.Start(profile.MemProfile, profile.ProfilePath("cmd/tofiks/")).Stop()
	}
	if len(os.Args) > 1 && os.Args[1] == "bench" {
		runBench()
		return
	}

	e := eval.NewEvalEngine()

	input := bufio.NewScanner(os.Stdin)
	for {
		input.Scan()
		cmd := uci.ParseUCI(input.Text())
		if cmd != nil {
			switch cmd.(type) {
			case *uci.Quit:
				return
			case *uci.Go, *uci.IsReady:
				go cmd.Exec(e)
			default:
				e.WG.Wait()
				e.WG.Add(1)
				go cmd.Exec(e)
			}
		}
	}
}

var benchPositions = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
	"r1bqkbnr/pppppppp/2n5/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 1 2",
}

func runBench() {
	const benchDepth = 10
	totalNodes := 0
	start := time.Now()

	for _, fen := range benchPositions {
		e := eval.NewEvalEngine()
		if err := e.Board.ImportFEN(fen); err != nil {
			log.Fatalf("bad bench FEN: %v", err)
		}
		ctx := context.Background()
		e.IDSearch(ctx, benchDepth, true)
		totalNodes += e.Stats.TotalNodes()
	}

	elapsed := time.Since(start)
	nps := int64(totalNodes)
	if elapsed.Milliseconds() > 0 {
		nps = (1000 * int64(totalNodes)) / elapsed.Milliseconds()
	}
	fmt.Printf("%d nodes %d nps\n", totalNodes, nps)
}

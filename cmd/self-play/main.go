package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/likeawizard/tofiks/internal/board"
	"github.com/likeawizard/tofiks/internal/config"
	eval "github.com/likeawizard/tofiks/internal/evaluation"
	_ "go.uber.org/automaxprocs"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load app config: %s\n", err)
	}
	b1 := &board.Board{}
	b1.Init(cfg)
	e, err := eval.NewEvalEngine(b1, cfg)
	if err != nil {
		fmt.Printf("Unable to load EvalEngine: %s\n", err)
		return
	}
	moves := make([]board.Move, 0)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10*1000)
			// ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500*1000)
			pv := []board.Move{}
			move, _ := e.GetMove(ctx, &pv, false)
			defer cancel()
			if move == 0 {
				fmt.Println("No legal moves.")
				return
			}
			b1.MakeMove(move)
			moves = append(moves, move)
			fmt.Println("playing:", move.String())

			b1.WritePGNToFile(b1.GeneratePGN(moves), "./dump.pgn")
		}

	}()
	wg.Wait()
}

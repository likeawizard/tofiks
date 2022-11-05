package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/likeawizard/tofiks/internal/board"
	"github.com/likeawizard/tofiks/internal/config"
	eval "github.com/likeawizard/tofiks/internal/evaluation"
	"github.com/pkg/profile"
	_ "go.uber.org/automaxprocs"
)

func main() {
	cfg, err := config.LoadConfig()
	defer profile.Start(profile.CPUProfile).Stop()

	if err != nil {
		fmt.Printf("Failed to load app config: %s\n", err)
	}
	fen := flag.String("fen", "", "FEN")
	flag.Parse()
	b := &board.Board{}
	b.ImportFEN(*fen)
	if b.ExportFEN() != *fen {
		fmt.Printf("Error importing FEN: %s, %s\n", b.ExportFEN(), *fen)
		return
	}
	e, err := eval.NewEvalEngine(b, cfg)
	if err != nil {
		fmt.Printf("Unable to load EvalEngine: %s\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500*1000)
	start := time.Now()
	pv := []board.Move{}
	move, ponder := e.GetMove(ctx, &pv, false)
	defer cancel()
	b.MakeMove(move)
	fmt.Println("bestmove", move, "ponder", ponder, time.Since(start))
}

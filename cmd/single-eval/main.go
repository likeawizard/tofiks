package main

import (
	"flag"
	"fmt"

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

	fmt.Println(e.GetEvaluation(b))
}

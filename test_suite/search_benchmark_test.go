package testsuite

import (
	"context"
	"os"
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

const searchBenchDepth = 6

var searchBenchPositions = perftResults[1:]

func BenchmarkPVS(b *testing.B) {
	for _, perft := range searchBenchPositions {
		e := eval.NewEvalEngine()
		e.Board = board.NewBoard(perft.fen)

		color := int16(1)
		if e.Board.Side != board.WHITE {
			color = -color
		}

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var line []board.Move
				e.PVS(context.Background(), nil, &line, searchBenchDepth, 0, -eval.Inf, eval.Inf, true, color)
			}
		})
	}
}

func BenchmarkQuiescence(b *testing.B) {
	for _, perft := range searchBenchPositions {
		e := eval.NewEvalEngine()
		e.Board = board.NewBoard(perft.fen)

		color := int16(1)
		if e.Board.Side != board.WHITE {
			color = -color
		}

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e.Quiescence(context.Background(), 0, -eval.Inf, eval.Inf, color)
			}
		})
	}
}

func BenchmarkIDSearch(b *testing.B) {
	old := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	defer func() { os.Stdout = old }()

	for _, perft := range searchBenchPositions {
		e := eval.NewEvalEngine()
		e.Board = board.NewBoard(perft.fen)

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e.IDSearch(context.Background(), searchBenchDepth, false)
			}
		})
	}
}

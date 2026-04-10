package testsuite

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/search"
)

var evalBenchPositions = perftResults[1:]

func BenchmarkGetEvaluation(b *testing.B) {
	for _, perft := range evalBenchPositions {
		e := search.NewEngine()
		e.Board = board.NewBoard(perft.fen)

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e.Eval.GetEvaluation(e.Board)
			}
		})
	}
}

func BenchmarkGetGamePhase(b *testing.B) {
	for _, perft := range evalBenchPositions {
		brd := board.NewBoard(perft.fen)

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				brd.GetGamePhase()
			}
		})
	}
}

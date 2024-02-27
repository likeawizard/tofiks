package testsuite

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

func BenchmarkMoveGen(b *testing.B) {
	for _, perft := range perftResults {
		brd := board.Board{}
		err := brd.ImportFEN(perft.fen)
		if err != nil {
			b.Fatalf("Failed to import FEN: %v", err)
		}
		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				brd.PseudoMoveGen()
			}
		})
	}
}

func BenchmarkMakeUnmake(b *testing.B) {
	for _, perft := range perftResults {
		brd := board.Board{}
		err := brd.ImportFEN(perft.fen)
		if err != nil {
			b.Fatalf("Failed to import FEN: %v", err)
		}
		moves := brd.PseudoMoveGen()
		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for i := range moves {
					umove := brd.MakeMove(moves[i])
					umove()
				}
			}
		})
	}
}

func BenchmarkGetEvaluation(b *testing.B) {
	for _, perft := range perftResults {
		e := eval.NewEvalEngine()
		e.Board = &board.Board{}
		err := e.Board.ImportFEN(perft.fen)
		if err != nil {
			b.Fatalf("Failed to import FEN: %v", err)
		}
		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e.GetEvaluation(e.Board)
			}
		})
	}
}

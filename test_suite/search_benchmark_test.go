package testsuite

import (
	"os"
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/search"
)

const searchBenchDepth = 6

var searchBenchPositions = perftResults[1:]

func BenchmarkPVS(b *testing.B) {
	for _, perft := range searchBenchPositions {
		e := search.NewEngine()
		e.Board = board.NewBoard(perft.fen)

		color := int16(1)
		if e.Board.Side != board.White {
			color = -color
		}

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var line []board.Move
				e.PVS(nil, &line, searchBenchDepth, 0, -search.Inf, search.Inf, true, color)
			}
		})
	}
}

func BenchmarkQuiescence(b *testing.B) {
	for _, perft := range searchBenchPositions {
		e := search.NewEngine()
		e.Board = board.NewBoard(perft.fen)

		color := int16(1)
		if e.Board.Side != board.White {
			color = -color
		}

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e.Quiescence(0, -search.Inf, search.Inf, color)
			}
		})
	}
}

func BenchmarkSEE(b *testing.B) {
	for _, perft := range searchBenchPositions {
		e := search.NewEngine()
		e.Board = board.NewBoard(perft.fen)
		captures := e.Board.PseudoCaptureAndQueenPromoGen()

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, m := range captures {
					if m.IsCapture() {
						e.SEE(m.From(), m.To())
					}
				}
			}
		})
	}
}

func BenchmarkIDSearch(b *testing.B) {
	old := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	defer func() { os.Stdout = old }()

	for _, perft := range searchBenchPositions {
		e := search.NewEngine()
		e.Board = board.NewBoard(perft.fen)

		b.Run(perft.position, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e.IDSearch(searchBenchDepth, false)
			}
		})
	}
}

package eval

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
)

var benchPositions = []string{
	board.StartingFEN,
	"r1bqkb1r/pppppppp/2n2n2/8/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3",
	"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
	"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
	"r3k2r/ppp2ppp/2n1bn2/2bpp1B1/4P3/2NP1N2/PPP2PPP/R2QKB1R w KQkq - 0 7",
	"8/pp3ppp/2p5/4k3/4P3/2PP4/PP3PPP/4K3 w - - 0 30",
}

func BenchmarkEvaluatePawns(b *testing.B) {
	for _, fen := range benchPositions {
		bd := board.NewBoard(fen)
		b.Run(fen, func(b *testing.B) {
			for range b.N {
				evaluatePawns(bd)
			}
		})
	}
}

func BenchmarkPawnTableProbeHit(b *testing.B) {
	for _, fen := range benchPositions {
		bd := board.NewBoard(fen)
		pt := NewPawnTable()
		score := evaluatePawns(bd)
		pt.Store(bd.PawnHash, score)
		b.Run(fen, func(b *testing.B) {
			for range b.N {
				pt.Probe(bd.PawnHash)
			}
		})
	}
}

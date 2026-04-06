package eval

import (
	"math/rand/v2"
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
)

// generateCaptureInfos creates random CaptureInfo values for benchmarking.
func generateCaptureInfos(n int) []CaptureInfo {
	r := rand.New(rand.NewPCG(42, 99))
	infos := make([]CaptureInfo, n)
	for i := range infos {
		infos[i] = newCaptureInfo(
			int8(r.IntN(2)),
			r.IntN(6),
			board.Square(r.IntN(64)),
			r.IntN(7),
		)
	}
	return infos
}

func BenchmarkCaptureInfoCreate(b *testing.B) {
	e := NewEvalEngine()
	e.Board = board.NewBoard("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1")
	// Create a capture move: black pawn d7d5 isn't a capture, use a known capture position
	e.Board = board.NewBoard("rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 2")
	moves := e.Board.PseudoMoveGen()
	// Find a capture move
	var capMove board.Move
	for _, m := range moves {
		if m.IsCapture() {
			capMove = m
			break
		}
	}
	if capMove == 0 {
		b.Fatal("no capture move found")
	}

	b.ResetTimer()
	for range b.N {
		_ = e.captureInfo(capMove)
	}
}

func BenchmarkUpdateCaptureHistoryByInfo(b *testing.B) {
	e := NewEvalEngine()
	infos := generateCaptureInfos(1024)

	b.ResetTimer()
	for i := range b.N {
		e.updateCaptureHistoryByInfo(infos[i%1024], 25)
	}
}

func BenchmarkUpdateCaptureHistoryByInfoMalus(b *testing.B) {
	e := NewEvalEngine()
	infos := generateCaptureInfos(1024)

	b.ResetTimer()
	for i := range b.N {
		e.updateCaptureHistoryByInfo(infos[i%1024], -25)
	}
}

// BenchmarkCaptureHistoryBetaCutoff simulates the beta cutoff path:
// bonus for best capture + malus for N tried captures.
func BenchmarkCaptureHistoryBetaCutoff(b *testing.B) {
	e := NewEvalEngine()
	infos := generateCaptureInfos(1024)

	b.ResetTimer()
	for i := range b.N {
		best := infos[i%1024]
		e.updateCaptureHistoryByInfo(best, 25)
		// Simulate 4 tried captures getting malus
		for j := 1; j <= 4; j++ {
			e.updateCaptureHistoryByInfo(infos[(i+j)%1024], -25)
		}
	}
}

// BenchmarkCaptureInfoArrayStack measures creating and filling the tried captures array.
func BenchmarkCaptureInfoArrayStack(b *testing.B) {
	e := NewEvalEngine()
	e.Board = board.NewBoard("r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4")
	moves := e.Board.PseudoMoveGen()
	var capMoves []board.Move
	for _, m := range moves {
		if m.IsCapture() {
			capMoves = append(capMoves, m)
		}
	}
	if len(capMoves) == 0 {
		b.Fatal("no capture moves")
	}

	b.ResetTimer()
	for range b.N {
		var tried [32]CaptureInfo
		n := 0
		for _, m := range capMoves {
			if n < 32 {
				tried[n] = e.captureInfo(m)
				n++
			}
		}
		_ = tried
		_ = n
	}
}

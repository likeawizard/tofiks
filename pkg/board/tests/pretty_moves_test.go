package board

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
)

func TestSimplePawnMove(t *testing.T) {
	move := "e2e4"
	pretty := "e4"
	b := board.NewBoard("")

	t.Run("Simple pawn move", func(t *testing.T) {
		if b.UCIToAlgebraic(move) != pretty {
			t.Errorf("Got: %s Want: %s", move, pretty)
		}
	})
}

func TestPawnCapture(t *testing.T) {
	var b board.Board
	move := "e4d5"
	pretty := "exd5"
	b.ImportFEN("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2")

	t.Run("e4 pawn takes d5 pawn", func(t *testing.T) {
		if b.UCIToAlgebraic(move) != pretty {
			t.Errorf("Got: %s Want: %s", move, pretty)
		}
	})
}

func TestSimplePieceMoveAndRankConflict(t *testing.T) {
	var b board.Board
	move := "g3h5"
	pretty := "Nh5"
	b.ImportFEN("rnbqkbnr/pppppppp/8/8/8/2N3N1/PPPPPPPP/R1BQKB1R w KQkq - 0 1")

	t.Run("Knight moves to h5 no conflict", func(t *testing.T) {
		prettified := b.UCIToAlgebraic(move)
		if prettified != pretty {
			t.Errorf("Got: %s Want: %s", prettified, pretty)
		}
	})

	move = "c3e4"
	pretty = "Nce4"

	t.Run("Knight from c3 moves to no conflict", func(t *testing.T) {
		prettified := b.UCIToAlgebraic(move)
		if prettified != pretty {
			t.Errorf("Got: %s Want: %s", prettified, pretty)
		}
	})
}

func TestSimplePieceMoveAndFileConflict(t *testing.T) {
	var b board.Board
	move := "f2g2"
	pretty := "Rg2"
	b.ImportFEN("1k6/5R2/8/8/8/8/5R2/1K6 w - - 0 1")

	t.Run("Rook moves to g2 no conflict", func(t *testing.T) {
		prettified := b.UCIToAlgebraic(move)
		if prettified != pretty {
			t.Errorf("Got: %s Want: %s", prettified, pretty)
		}
	})

	move = "f2f4"
	pretty = "R2f4"

	t.Run("Need rank disambiguation", func(t *testing.T) {
		prettified := b.UCIToAlgebraic(move)
		if prettified != pretty {
			t.Errorf("Got: %s Want: %s", prettified, pretty)
		}
	})
}

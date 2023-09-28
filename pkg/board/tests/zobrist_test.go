package board

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
)

func TestZobristTransposition(t *testing.T) {
	moves1 := "e2e4 e7e5 g1f3 g8f6"
	moves2 := "g1f3 e7e5 e2e4 g8f6"
	b1 := board.NewBoard("")
	b2 := board.NewBoard("")

	b1.PlayMovesUCI(moves1)
	b2.PlayMovesUCI(moves2)

	t.Run("Check transposition hash", func(t *testing.T) {
		if b1.Hash != b2.Hash {
			t.Fatalf("Hashes not equal")
		}
	})
}

func TestZobristDiff(t *testing.T) {
	moves1 := "d2d4 e7e5 g1f3 g8f6"
	moves2 := "e2e4 e7e5 g1f3 g8f6"
	b1 := board.NewBoard("")
	b2 := board.NewBoard("")

	b1.PlayMovesUCI(moves1)
	b2.PlayMovesUCI(moves2)

	t.Run("Verify different hash", func(t *testing.T) {
		if b1.Hash == b2.Hash {
			t.Fatalf("Hashes are equal")
		}
	})
}

// Asymetrical transposition - same position, opposite side to move.
func TestTempoLoss(t *testing.T) {
	var b board.Board
	moves := "e8e7 e1f2 e7e8 f2e2 e8e7 e2e1 e7e8"

	b.ImportFEN("rnbqkbnr/pppp1ppp/4p3/8/8/4PP2/PPPP2PP/RNBQKBNR b - - 0 2")
	seed := b.Hash
	b.PlayMovesUCI(moves)

	t.Run("Verify different hash", func(t *testing.T) {
		if b.Hash == seed {
			t.Fatalf("Hashes are equal")
		}
	})
}

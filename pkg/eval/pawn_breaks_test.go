package eval

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
)

func TestPawnBreaksStartingPosition(t *testing.T) {
	b := board.NewBoard(board.StartingFEN)
	if got := evaluatePawnBreaks(b); got != 0 {
		t.Errorf("starting position: want 0, got %d", got)
	}
}

func TestPawnBreaksDoublePush(t *testing.T) {
	// Only White d2 vs Black c5. d2-d4 attacks c5 (break); c5-c4 attacks nothing.
	b := board.NewBoard("4k3/8/8/2p5/8/8/3P4/4K3 w - - 0 1")
	origBreak := PawnBreak
	PawnBreak = 1
	defer func() { PawnBreak = origBreak }()
	got := evaluatePawnBreaks(b)
	if got != 1 {
		t.Errorf("d2 vs c5: want +1 (d2-d4 break), got %d", got)
	}
}

func TestPawnBreaksPieceBlocker(t *testing.T) {
	// Same as above but a White knight sits on d4, blocking the double push.
	b := board.NewBoard("4k3/8/8/2p5/3N4/8/3P4/4K3 w - - 0 1")
	origBreak := PawnBreak
	PawnBreak = 1
	defer func() { PawnBreak = origBreak }()
	got := evaluatePawnBreaks(b)
	if got != 0 {
		t.Errorf("knight on d4 blocks d2-d4: want 0, got %d", got)
	}
}

func TestPawnBreaksSinglePushFromRank3(t *testing.T) {
	// White d3 pawn, Black c5 pawn. d3-d4 attacks c5 (break). c5-c4 attacks d3 (break). Net 0.
	b := board.NewBoard("4k3/8/8/2p5/8/3P4/8/4K3 w - - 0 1")
	origBreak := PawnBreak
	PawnBreak = 1
	defer func() { PawnBreak = origBreak }()
	got := evaluatePawnBreaks(b)
	if got != 0 {
		t.Errorf("symmetric d3/c5: want 0 (breaks cancel), got %d", got)
	}
}

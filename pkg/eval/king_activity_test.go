package eval

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
)

// netKingActivity is white's king activity minus black's.
func netKingActivity(fen string) int {
	b := board.NewBoard(fen)
	wk := b.Pieces[board.White][board.Kings].LS1B()
	bk := b.Pieces[board.Black][board.Kings].LS1B()
	return getKingActivity(b, wk, board.White) - getKingActivity(b, bk, board.Black)
}

func TestKingShoulderBreaksSymmetry(t *testing.T) {
	// Both kings equally central: centralization cancels, shouldering must too.
	if got := netKingActivity("8/8/2k5/8/8/2K5/8/8 w - - 0 1"); got != 0 {
		t.Errorf("equal centralization: want 0, got %d", got)
	}

	// White king central, black king pushed toward the edge. Per unit of
	// centralization gap, kings in contact must score more than kings far apart.
	closeGap := DistCenter(sq("b6")) - DistCenter(sq("d4"))
	farGap := DistCenter(sq("a8")) - DistCenter(sq("d4"))
	closeShoulder := netKingActivity("8/8/1k6/8/3K4/8/8/8 w - - 0 1") + KingActivityDistCenter*closeGap
	farShoulder := netKingActivity("k7/8/8/8/3K4/8/8/8 w - - 0 1") + KingActivityDistCenter*farGap
	if closeShoulder <= 0 {
		t.Errorf("shouldering should favor the central king, got %d", closeShoulder)
	}
	if closeShoulder*farGap <= farShoulder*closeGap {
		t.Errorf("close kings should weigh more per unit gap: close=%d/%d far=%d/%d", closeShoulder, closeGap, farShoulder, farGap)
	}

	// Color flip negates the score.
	if a, b := netKingActivity("8/8/1k6/8/3K4/8/8/8 w - - 0 1"), netKingActivity("8/8/8/3k4/8/1K6/8/8 b - - 0 1"); a != -b {
		t.Errorf("mirror: want %d, got %d", -a, b)
	}
}

func sq(name string) int {
	return int(name[0]-'a') + 8*int('8'-name[1])
}

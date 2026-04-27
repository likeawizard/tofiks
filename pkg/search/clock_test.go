package search

import (
	"testing"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
)

// TestGetMovetimeLowTimeNotInfinite guards against issue #106 — with only a
// few ms left on the clock, integer truncation in the per-move budget formula
// used to return 0, which NewTimeControl/GetContext both reinterpreted as
// "infinite search".
func TestGetMovetimeLowTimeNotInfinite(t *testing.T) {
	c := &Clock{Wtime: 10, Btime: 4879}
	// fmCounter high enough that movestogo is clamped to 10.
	got := c.GetMovetime(94, board.White)
	if got <= 0 {
		t.Fatalf("GetMovetime with 10ms wtime returned %v, want > 0", got)
	}
	if got > 10*time.Millisecond {
		t.Fatalf("GetMovetime returned %v, want <= remaining 10ms", got)
	}
}

func TestNewTimeControlLowTimeNotInfinite(t *testing.T) {
	c := &Clock{Wtime: 10, Btime: 4879}
	tc := c.NewTimeControl(94, board.White)
	if tc.budget <= 0 {
		t.Fatalf("budget = %v, want > 0 (issue #106)", tc.budget)
	}
	if tc.maxBudget <= 0 || tc.maxBudget > 10*time.Millisecond {
		t.Fatalf("maxBudget = %v, want in (0, 10ms]", tc.maxBudget)
	}
}

func TestNewTimeControlLowTimeHasHardLimit(t *testing.T) {
	c := &Clock{Wtime: 10, Btime: 4879}
	tc := c.NewTimeControl(94, board.White)
	if tc.hardLimit <= 0 {
		t.Fatal("NewTimeControl with low time should set a hardLimit (issue #106)")
	}
}

// TestNewTimeControlOverheadExceedsTime guards against the corner case where
// remaining time is less than the configured overhead. hardLimit must still
// be clamped to remainingTime so the deadline timer fires before we flag.
func TestNewTimeControlOverheadExceedsTime(t *testing.T) {
	c := &Clock{Wtime: 3, Btime: 4879, Overhead: 20}
	tc := c.NewTimeControl(50, board.White)
	if tc.hardLimit <= 0 || tc.hardLimit > 3*time.Millisecond {
		t.Fatalf("hardLimit = %v, want in (0, 3ms]", tc.hardLimit)
	}
}

// TestNewTimeControlMovetimeBelowOverhead guards against the case where
// movetime is set but ≤ overhead. Should panic-search, not fall through to
// the bare-"go"-is-infinite path.
func TestNewTimeControlMovetimeBelowOverhead(t *testing.T) {
	c := &Clock{Movetime: 10, Overhead: 20}
	tc := c.NewTimeControl(10, board.White)
	if tc.hardLimit <= 0 {
		t.Fatalf("hardLimit = %v, want > 0 (panic mode)", tc.hardLimit)
	}
}

// TestGetMovetimeGoNoArgsStillInfinite preserves the `go` (no parameters)
// behavior: without any time info, the engine should search infinitely until
// an external stop. Only c.Infinite / truly-zero clocks should trigger this.
func TestNewTimeControlGoNoArgs(t *testing.T) {
	c := &Clock{}
	tc := c.NewTimeControl(10, board.White)
	if tc.budget != 0 {
		t.Fatalf("NewTimeControl with no time info budget = %v, want 0", tc.budget)
	}
	if tc.hardLimit != 0 {
		t.Fatalf("NewTimeControl with no time info hardLimit = %v, want 0", tc.hardLimit)
	}
}

func TestNewTimeControlInfiniteFlag(t *testing.T) {
	c := &Clock{Wtime: 10_000, Infinite: true}
	tc := c.NewTimeControl(10, board.White)
	if tc.hardLimit != 0 {
		t.Fatalf("NewTimeControl with Infinite=true hardLimit = %v, want 0", tc.hardLimit)
	}
}

// TestGetMovetimeMovetimeCapIgnoresBudgetFix ensures the `go movetime` path is
// unchanged by the low-time fix.
func TestGetMovetimeMovetime(t *testing.T) {
	c := &Clock{Movetime: 500}
	got := c.GetMovetime(1, board.White)
	if got != 500*time.Millisecond {
		t.Fatalf("GetMovetime with movetime=500 returned %v, want 500ms", got)
	}
}

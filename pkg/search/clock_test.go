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
	c := &Clock{Wtime: 10_000, Btime: 600_000, HasClock: true, Infinite: true}
	tc := c.NewTimeControl(10, board.White)
	defer tc.Stop()
	if !tc.infinite || tc.timer != nil || tc.ShouldStop() {
		t.Fatal("go infinite/ponder must run unlimited until stop or ponderhit")
	}
	if tc.budget <= 0 || tc.hardLimit <= 0 {
		t.Fatalf("budget=%v hardLimit=%v: the go command's limits must be kept for ponderhit", tc.budget, tc.hardLimit)
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

// TestNewTimeControlZeroClockHasDeadline checks that a clock at zero still
// gets a hard deadline.
func TestNewTimeControlZeroClockHasDeadline(t *testing.T) {
	c := &Clock{Wtime: 0, Btime: 0, HasClock: true}
	tc := c.NewTimeControl(60, board.White)
	if tc.hardLimit <= 0 || tc.hardLimit > 2*time.Millisecond {
		t.Fatalf("hardLimit = %v, want in (0, 2ms]", tc.hardLimit)
	}
}

// TestGetMovetimeHonoursMovestogo checks that a movestogo from the GUI is used
// instead of the built-in estimate.
func TestGetMovetimeHonoursMovestogo(t *testing.T) {
	paced := (&Clock{Wtime: 60000, Btime: 60000, HasClock: true}).GetMovetime(1, board.White)
	last := (&Clock{Wtime: 60000, Btime: 60000, Movestogo: 1, HasClock: true}).GetMovetime(1, board.White)
	if last <= 5*paced {
		t.Fatalf("movestogo 1 allocated %v, paced default %v; want movestogo honoured", last, paced)
	}
}

// TestPonderHitBudgetFromOwnClock guards against the lichess time forfeits:
// after a ponderhit the remaining think time must derive from the pondering
// side's own clock (20s here), never from the opponent's (600s).
func TestPonderHitBudgetFromOwnClock(t *testing.T) {
	c := &Clock{Wtime: 20_000, Btime: 600_000, HasClock: true, Infinite: true}
	tc := c.NewTimeControl(45, board.White)
	defer tc.Stop()
	tc.PonderHit()
	if tc.infinite || !tc.IsPonderHit() || tc.timer == nil {
		t.Fatal("PonderHit must arm the search")
	}
	// Own budget is 20s/11 ≈ 1.8s: soft stop ≈ 0.6s, hard stop ≈ 1.2s.
	// The opponent's clock would give 18s and 36s.
	if tc.budget <= 0 || tc.budget > 2*time.Second {
		t.Fatalf("budget = %v, want within (0, 2s]", tc.budget)
	}
	if tc.hardLimit <= 0 || tc.hardLimit > 4*time.Second {
		t.Fatalf("hardLimit = %v, want within (0, 4s]", tc.hardLimit)
	}
}

// TestPonderHitWithoutLimitsKeepsSearching: a go ponder without clocks has no
// time limits to apply, so after ponderhit the search runs on as a bare go
// would, until its depth limit or stop.
func TestPonderHitWithoutLimitsKeepsSearching(t *testing.T) {
	tc := (&Clock{Infinite: true}).NewTimeControl(10, board.White)
	tc.PonderHit()
	if tc.infinite || tc.ShouldStop() || tc.timer != nil || tc.ShouldAbort() {
		t.Fatal("ponderhit without limits must neither stop nor abort the search")
	}
}

// TestPonderHitIgnoresTimedSearch: a stale ponderhit, arriving after the
// ponder search ended and a normal search started, must not touch the latter.
func TestPonderHitIgnoresTimedSearch(t *testing.T) {
	c := &Clock{Wtime: 20_000, Btime: 20_000, HasClock: true}
	tc := c.NewTimeControl(45, board.White)
	defer tc.Stop()
	budget := tc.budget
	tc.PonderHit()
	if tc.budget != budget || tc.IsPonderHit() {
		t.Fatalf("budget %v -> %v: ponderhit altered a timed search", budget, tc.budget)
	}
}

// TestPonderHitHardDeadlineAborts checks that the deadline armed by PonderHit
// flips the abort flag.
func TestPonderHitHardDeadlineAborts(t *testing.T) {
	// 60ms on our clock: budget 5ms, so ponderhit allows 1ms soft / 2ms hard.
	c := &Clock{Wtime: 60, Btime: 600_000, HasClock: true, Infinite: true}
	tc := c.NewTimeControl(45, board.White)
	defer tc.Stop()
	tc.PonderHit()
	deadline := time.Now().Add(time.Second)
	for !tc.ShouldAbort() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !tc.ShouldAbort() {
		t.Fatal("hard deadline after ponderhit did not abort the search")
	}
}

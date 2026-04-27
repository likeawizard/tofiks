package search

import (
	"sync/atomic"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
)

// Time control constants. Conservative starting values; tune via SPSA later.
const (
	// Eval drop (in centipawns) that triggers a time extension.
	tcEvalDropThreshold = 30
	// Fraction of budget to extend on significant eval drop (1/4 = 25%).
	tcEvalDropExtendDiv = 4
)

type Clock struct {
	Wtime     int
	Btime     int
	Winc      int
	Binc      int
	Overhead  int
	Movetime  int
	Movestogo int
	Infinite  bool
}

// remainingTime returns the actual clock time left for the given side,
// minus overhead, as a safety cap for time control.
func (c *Clock) remainingTime(side int8) time.Duration {
	t := c.Wtime
	if side == board.Black {
		t = c.Btime
	}
	return time.Millisecond * time.Duration(max(t-c.Overhead, 0))
}

func (c *Clock) GetMovetime(fmCounter int, side int8) time.Duration {
	c.Movestogo = max(40-fmCounter, 10)
	switch {
	case c.Movetime > 0:
		return time.Millisecond * time.Duration(c.Movetime-c.Overhead)
	default:
		movestogo := c.Movestogo
		t := c.Wtime
		inc := c.Winc
		if side == board.Black {
			t = c.Btime
			inc = c.Binc
		}
		base := (t+(inc-c.Overhead)*movestogo)/(movestogo+1) - c.Overhead
		// Safety check for no time allocated.
		if base <= 0 && t > 0 {
			base = max((t-c.Overhead)/2, 1)
		}
		return time.Millisecond * time.Duration(base)
	}
}

// TimeControl manages time for a single search. Owns the soft per-iteration
// budget for prediction-based stops and a hard wall-clock deadline that
// aborts the search via the aborted flag.
//
// budget=0 disables iteration prediction (used for movetime / UCI infinite /
// no-clock fallback). hardLimit=0 disables the deadline timer.
type TimeControl struct {
	start            time.Time
	lastIterStart    time.Time
	timer            *time.Timer
	budget           time.Duration
	maxBudget        time.Duration
	hardLimit        time.Duration
	lastIterDuration time.Duration
	bestMoveChanges  int
	iterations       int
	prevBestMove     board.Move
	aborted          atomic.Bool
	prevEval         int16
}

// NewTimeControl creates a TimeControl for the current search.
func (c *Clock) NewTimeControl(fmCounter int, side int8) *TimeControl {
	now := time.Now()
	tc := &TimeControl{start: now, lastIterStart: now}

	if c.Infinite {
		return tc
	}
	base := c.GetMovetime(fmCounter, side)
	if base <= 0 {
		return tc
	}
	hardLimit := base * 2
	if r := c.remainingTime(side); r > 0 && r < hardLimit {
		hardLimit = r
	}
	tc.hardLimit = hardLimit
	tc.armDeadline()
	if c.Movetime > 0 {
		// Movetime mode: hard deadline only, no iteration prediction.
		return tc
	}
	tc.budget = base
	tc.maxBudget = hardLimit
	return tc
}

// armDeadline starts a timer that flips the aborted flag when hardLimit
// elapses. Keeps the per-node abort check to a single atomic load instead
// of a time.Now() syscall.
func (tc *TimeControl) armDeadline() {
	if tc.hardLimit <= 0 {
		return
	}
	tc.timer = time.AfterFunc(tc.hardLimit, func() { tc.aborted.Store(true) })
}

// Stop cancels the deadline timer when the search finishes naturally.
func (tc *TimeControl) Stop() {
	if tc.timer != nil {
		tc.timer.Stop()
	}
}

// ShouldAbort reports whether the search must stop. Hot-path: single atomic load.
func (tc *TimeControl) ShouldAbort() bool {
	return tc.aborted.Load()
}

// Abort signals the running search to stop at its next abort check.
func (tc *TimeControl) Abort() {
	tc.aborted.Store(true)
}

// ShouldStop returns true if the next iteration is predicted to not complete
// within the remaining budget. Uses the last iteration's duration to
// estimate the next (assuming ~4x branching factor).
// budget==0 disables prediction (movetime / infinite / no-clock).
func (tc *TimeControl) ShouldStop() bool {
	if tc.budget == 0 {
		return false
	}
	elapsed := time.Since(tc.start)
	predicted := tc.lastIterDuration * 4
	return elapsed+predicted > tc.budget
}

// RecordIteration updates TC tracking after a completed iteration.
// Extends the budget when eval drops significantly (position is harder than
// expected and worth investing more time).
func (tc *TimeControl) RecordIteration(best board.Move, eval int16) {
	if tc.iterations > 0 {
		if best != tc.prevBestMove {
			tc.bestMoveChanges++
		}
		if tc.budget > 0 {
			drop := int(tc.prevEval) - int(eval)
			if drop > tcEvalDropThreshold {
				tc.extend(tc.budget / tcEvalDropExtendDiv)
			}
		}
	}
	tc.prevBestMove = best
	tc.prevEval = eval
	tc.iterations++
}

// AspirationFailed extends the budget on aspiration window failure — the
// position is volatile and worth investing more time.
func (tc *TimeControl) AspirationFailed() {
	if tc.budget == 0 {
		return
	}
	tc.extend(tc.budget / 2)
}

// IterationStarted records the start of a new depth iteration.
func (tc *TimeControl) IterationStarted() {
	tc.lastIterStart = time.Now()
}

// IterationFinished records the end of a depth iteration.
func (tc *TimeControl) IterationFinished() {
	tc.lastIterDuration = time.Since(tc.lastIterStart)
}

func (tc *TimeControl) extend(d time.Duration) {
	tc.budget += d
	if tc.budget > tc.maxBudget {
		tc.budget = tc.maxBudget
	}
}

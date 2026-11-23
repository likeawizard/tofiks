package search

import (
	"sync"
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
	// Fraction of the normal move budget to keep searching after a ponderhit:
	// the position has already been searched for the opponent's whole think time.
	tcPonderHitDiv = 3
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
	HasClock  bool
}

// remainingTime returns the actual clock time left for the given side,
// minus overhead, as a safety cap for time control.
func (c *Clock) remainingTime(side int8) time.Duration {
	t := c.Wtime
	if side == board.Black {
		t = c.Btime
	}
	return time.Millisecond * time.Duration(max(t-c.Overhead, 1))
}

func (c *Clock) GetMovetime(fmCounter int, side int8) time.Duration {
	switch {
	case c.Movetime > 0:
		return time.Millisecond * time.Duration(c.Movetime-c.Overhead)
	default:
		movestogo := c.Movestogo
		if movestogo <= 0 {
			movestogo = max(40-fmCounter, 10)
		}
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

// hardLimitFor caps a soft budget: twice the budget, never more than the
// side's remaining clock, never below a millisecond.
func (c *Clock) hardLimitFor(side int8, soft time.Duration) time.Duration {
	return max(min(2*soft, c.remainingTime(side)), time.Millisecond)
}

// TimeControl manages time for a single search. Owns the soft per-iteration
// budget for prediction-based stops and a hard wall-clock deadline that
// aborts the search via the aborted flag.
//
// budget=0 disables iteration prediction (movetime / no-clock fallback).
// hardLimit=0 disables the deadline timer. An infinite search (go infinite,
// go ponder) carries the limits its go command described but leaves them
// unarmed until PonderHit.
//
// The search goroutine consults the limits once per iteration; the UCI reader
// may call Abort, PonderHit and Stop at any time. ShouldAbort stays a single
// atomic load on the per-node hot path.
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
	mu               sync.Mutex
	prevBestMove     board.Move
	aborted          atomic.Bool
	infinite         bool
	ponderHit        bool
	prevEval         int16
}

// NewTimeControl creates a TimeControl for the current search.
func (c *Clock) NewTimeControl(fmCounter int, side int8) *TimeControl {
	now := time.Now()
	tc := &TimeControl{start: now, lastIterStart: now, infinite: c.Infinite}

	base := c.GetMovetime(fmCounter, side)
	if base <= 0 {
		if !c.HasClock && c.Movetime <= 0 {
			// Bare go, go infinite or go depth: no time limits at all.
			return tc
		}
		// Add bare minimum of time when nothing is allocated.
		base = time.Millisecond
	}
	if c.Movetime > 0 {
		// Movetime mode: hard deadline only, no iteration prediction,
		// independent of wtime/btime.
		tc.hardLimit = 2 * base
	} else {
		tc.budget = base
		tc.maxBudget = c.hardLimitFor(side, base)
		tc.hardLimit = tc.maxBudget
	}
	if !tc.infinite {
		tc.mu.Lock()
		tc.armDeadline()
		tc.mu.Unlock()
	}
	return tc
}

// armDeadline starts the timer that flips the aborted flag when hardLimit
// elapses, replacing any earlier timer. Keeps the per-node abort check to a
// single atomic load instead of a time.Now() syscall. mu must be held.
func (tc *TimeControl) armDeadline() {
	if tc.timer != nil {
		tc.timer.Stop()
		tc.timer = nil
	}
	if tc.hardLimit <= 0 {
		return
	}
	tc.timer = time.AfterFunc(tc.hardLimit, func() { tc.aborted.Store(true) })
}

// Stop cancels the deadline timer when the search finishes naturally.
func (tc *TimeControl) Stop() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
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

// PonderHit turns a pondering search into the timed search its go command
// described, measured from now. The position has already been searched for
// the opponent's whole think time, so only a fraction of the budget is kept.
// A no-op unless the search is infinite.
func (tc *TimeControl) PonderHit() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if !tc.infinite {
		return
	}
	tc.infinite = false
	tc.ponderHit = true
	tc.start = time.Now()
	if tc.budget > 0 {
		tc.budget = max(tc.budget/tcPonderHitDiv, time.Millisecond)
		tc.maxBudget = min(2*tc.budget, tc.maxBudget)
		tc.hardLimit = tc.maxBudget
	}
	tc.armDeadline()
}

// IsPonderHit reports whether a ponder search has been converted into a timed
// search by PonderHit.
func (tc *TimeControl) IsPonderHit() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.ponderHit
}

// ShouldStop returns true if the next iteration is predicted to not complete
// within the remaining budget. Uses the last iteration's duration to
// estimate the next (assuming ~4x branching factor).
// Never stops an infinite search or one without a budget.
func (tc *TimeControl) ShouldStop() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.infinite || tc.budget == 0 {
		return false
	}
	return time.Since(tc.start)+tc.lastIterDuration*4 > tc.budget
}

// RecordIteration updates TC tracking after a completed iteration.
// Extends the budget when eval drops significantly (position is harder than
// expected and worth investing more time).
func (tc *TimeControl) RecordIteration(best board.Move, eval int16) {
	if tc.iterations > 0 {
		if best != tc.prevBestMove {
			tc.bestMoveChanges++
		}
		if int(tc.prevEval)-int(eval) > tcEvalDropThreshold {
			tc.extend(tcEvalDropExtendDiv)
		}
	}
	tc.prevBestMove = best
	tc.prevEval = eval
	tc.iterations++
}

// AspirationFailed extends the budget on aspiration window failure — the
// position is volatile and worth investing more time.
func (tc *TimeControl) AspirationFailed() {
	tc.extend(2)
}

// IterationStarted records the start of a new depth iteration.
func (tc *TimeControl) IterationStarted() {
	tc.lastIterStart = time.Now()
}

// IterationFinished records the end of a depth iteration.
func (tc *TimeControl) IterationFinished() {
	tc.lastIterDuration = time.Since(tc.lastIterStart)
}

// extend adds budget/div to the budget, capped at maxBudget. A no-op while
// pondering or without a budget.
func (tc *TimeControl) extend(div int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.infinite || tc.budget == 0 {
		return
	}
	tc.budget = min(tc.budget+tc.budget/time.Duration(div), tc.maxBudget)
}

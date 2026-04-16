package search

import (
	"context"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
)

// Time control constants. Conservative starting values; tune via SPSA later.
const (
	// Eval drop (in centipawns) that triggers a time extension.
	tcEvalDropThreshold = 30
	// Fraction of budget to extend on significant eval drop (1/4 = 25%).
	tcEvalDropExtendDiv = 4
	// Node-fraction thresholds: if the best move consumed this fraction of root
	// nodes, the position is "clear" and we can stop earlier; below the low
	// threshold, the position is contested and we extend.
	tcNodeFracHigh = 90 // % — clear position, shrink budget
	tcNodeFracLow  = 40 // % — contested position, extend budget
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

func (c *Clock) GetContext(fmCounter int, side int8) (context.Context, context.CancelFunc) {
	clock := c.GetMovetime(fmCounter, side)
	switch {
	case c.Infinite || clock == time.Duration(0):
		return context.WithCancel(context.Background())
	default:
		// Context deadline allows extensions but never exceeds remaining time.
		hardLimit := min(clock*2, c.remainingTime(side))
		return context.WithTimeout(context.Background(), hardLimit)
	}
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

// TimeControl manages time for a single search. It tracks iteration durations
// and predicts whether the next iteration will complete within the budget,
// avoiding wasted time on doomed iterations that the context would abort.
//
// budget is the soft limit (initially = base). maxBudget is the hard cap
// (2x base, matching the context deadline). Extensions on volatile positions
// push budget toward maxBudget.
type TimeControl struct {
	start            time.Time
	lastIterStart    time.Time
	budget           time.Duration
	maxBudget        time.Duration
	lastIterDuration time.Duration
	bestMoveChanges  int
	iterations       int
	extensions       uint // number of TC extensions granted; used for diminishing amounts
	bestMoveNodePct  int  // % of root nodes on best move (0-100), last completed iter
	prevBestMove     board.Move
	prevEval         int16
	infinite         bool
}

// NewTimeControl creates a TimeControl for the current search.
// Fixed movetime and infinite searches bypass time prediction.
func (c *Clock) NewTimeControl(fmCounter int, side int8) TimeControl {
	if c.Infinite || c.Movetime > 0 {
		return TimeControl{start: time.Now(), infinite: true}
	}
	base := c.GetMovetime(fmCounter, side)
	if base <= 0 {
		return TimeControl{start: time.Now(), infinite: true}
	}
	now := time.Now()
	remaining := c.remainingTime(side)
	return TimeControl{
		start:         now,
		budget:        base,
		maxBudget:     min(base*2, remaining),
		lastIterStart: now,
	}
}

// ShouldStop returns true if the next iteration is predicted to not complete
// within the remaining budget. Uses the last iteration's duration to
// estimate the next (assuming ~4x branching factor).
//
// When the best move consumed most root nodes (>90%), the position is clear
// and we compare against a tighter effective budget (stop sooner). Extensions
// for contested positions (low node fraction) happen in RecordIteration.
// A zero-value TimeControl never stops (safe for direct IDSearch calls in tests).
func (tc *TimeControl) ShouldStop() bool {
	if tc.infinite || tc.budget == 0 {
		return false
	}
	elapsed := time.Since(tc.start)
	predicted := tc.lastIterDuration * 4

	effectiveBudget := tc.budget
	if tc.iterations >= 5 && tc.bestMoveNodePct >= tcNodeFracHigh {
		effectiveBudget = effectiveBudget * 3 / 4
	}

	return elapsed+predicted > effectiveBudget
}

// RecordIteration updates TC tracking after a completed iteration.
// bestMoveNodePct is the percentage (0-100) of root nodes spent on the best
// move this iteration. Extends budget on eval drops and contested positions
// (low node fraction). Records node fraction for ShouldStop to use.
func (tc *TimeControl) RecordIteration(best board.Move, eval int16, bestMoveNodePct int) {
	if tc.iterations > 0 {
		if best != tc.prevBestMove {
			tc.bestMoveChanges++
		}
		if !tc.infinite && tc.budget > 0 {
			// Extend on significant eval drop.
			drop := int(tc.prevEval) - int(eval)
			if drop > tcEvalDropThreshold {
				tc.extendDiminishing(tc.budget / tcEvalDropExtendDiv)
			}
			// Extend on contested position (low node fraction on best move).
			if bestMoveNodePct > 0 && bestMoveNodePct <= tcNodeFracLow {
				tc.extendDiminishing(tc.budget / tcEvalDropExtendDiv)
			}
		}
	}
	tc.bestMoveNodePct = bestMoveNodePct
	tc.prevBestMove = best
	tc.prevEval = eval
	tc.iterations++
}

// AspirationFailed extends the budget on aspiration window failure — the
// position is volatile and worth investing more time.
func (tc *TimeControl) AspirationFailed() {
	if tc.infinite || tc.budget == 0 {
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

// extendDiminishing grants an extension that halves with each prior call.
// First = d, second = d/2, third = d/4, etc. Converges so total extensions
// can't exceed ~2d regardless of how many iterations fire.
func (tc *TimeControl) extendDiminishing(d time.Duration) {
	d >>= tc.extensions
	if d <= 0 {
		return
	}
	tc.extensions++
	tc.extend(d)
}

package search

import (
	"context"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
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
		return time.Millisecond * time.Duration((t+((inc-c.Overhead)*movestogo))/(movestogo+1)-c.Overhead)
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
// A zero-value TimeControl never stops (safe for direct IDSearch calls in tests).
func (tc *TimeControl) ShouldStop() bool {
	if tc.infinite || tc.budget == 0 {
		return false
	}
	elapsed := time.Since(tc.start)
	predicted := tc.lastIterDuration * 4
	return elapsed+predicted > tc.budget
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

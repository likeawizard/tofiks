package lichess

import (
	"context"
	"strings"
	"time"
)

type TimeManagment struct {
	Speed        string
	MoveTarget   int
	Move         int
	Time         int
	Inc          int
	LagStopWatch time.Time
	Lag          int
	isWhite      bool
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (tm *TimeManagment) UpdateClock(state GameState) {
	tm.Move = len(strings.Fields(state.Moves)) / 2
	if tm.isWhite {
		tm.Time, tm.Inc = state.Wtime, state.Winc
	} else {
		tm.Time, tm.Inc = state.Btime, state.Binc
	}
}

func (tm *TimeManagment) AllotTime() int {
	if tm.Speed == "correspondence" {
		return 60 * 1000
	}

	movesLeft := max(tm.MoveTarget-tm.Move, 10)
	expectedTimeLeft := tm.Time + (tm.Inc-tm.Lag)*movesLeft
	return max(expectedTimeLeft/(movesLeft+2), 100)
}

func (tm *TimeManagment) GetTimeoutContext() (context.Context, context.CancelFunc) {
	t := tm.AllotTime()
	return context.WithTimeout(context.Background(), time.Millisecond*time.Duration(t))
}

func (tm *TimeManagment) GetPonderContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))

}

func NewTimeManagement(gs GameState, isWhite bool) *TimeManagment {
	tm := TimeManagment{}
	tm.Speed = gs.Speed
	tm.isWhite = isWhite
	tm.Move = len(strings.Fields(gs.State.Moves)) / 2
	switch tm.Speed {
	case "correspondence", "classical":
		tm.MoveTarget = 40
	case "blitz", "rapid":
		tm.MoveTarget = 30
	case "bullet", "ultraBullet":
		tm.MoveTarget = 25
	default:
		tm.MoveTarget = 25
	}
	if tm.isWhite {
		tm.Time, tm.Inc = gs.State.Wtime, gs.State.Winc
	} else {
		tm.Time, tm.Inc = gs.State.Btime, gs.State.Binc
	}

	return &tm
}

func (state *State) GetTime(isWhite bool) (int, int) {
	if isWhite {
		return state.Wtime, state.Winc
	} else {
		return state.Btime, state.Winc
	}
}

func (tm *TimeManagment) StartStopWatch() {
	tm.LagStopWatch = time.Now()
}

func (tm *TimeManagment) MeasureLag() {
	// Limit lag to 500 in case of unexpected lag spike
	tm.Lag = int(time.Since(tm.LagStopWatch).Milliseconds())
}

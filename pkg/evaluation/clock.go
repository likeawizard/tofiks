package eval

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

func (c *Clock) GetContext(fmCounter, side int) (context.Context, context.CancelFunc) {
	clock := c.GetMovetime(fmCounter, side)
	switch {
	case c.Infinite || clock == time.Duration(0):
		return context.WithCancel(context.Background())
	default:
		return context.WithTimeout(context.Background(), clock)
	}
}

func (c *Clock) GetMovetime(fmCounter, side int) time.Duration {
	c.Movestogo = max(40-fmCounter, 10)
	switch {
	case c.Movetime > 0:
		return time.Millisecond * time.Duration(c.Movetime-c.Overhead)
	default:
		movestogo := c.Movestogo
		t := c.Wtime
		inc := c.Winc
		if side == board.BLACK {
			t = c.Btime
			inc = c.Binc
		}
		return time.Millisecond * time.Duration((t+((inc-c.Overhead)*movestogo))/(movestogo+1)-c.Overhead)
	}
}

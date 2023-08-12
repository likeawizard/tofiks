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
	var movetime = time.Millisecond * 100
	switch {
	case c.Infinite:
		return context.WithCancel(context.Background())
	default:
		movetime = c.GetMovetime(fmCounter, side)
	}

	return context.WithTimeout(context.Background(), movetime)

}

func (c *Clock) GetMovetime(fmCounter, side int) time.Duration {
	var movetime = time.Millisecond * 100
	c.Movestogo = Max(40-fmCounter, 10)
	switch {
	case c.Movetime > 0:
		movetime = time.Millisecond * time.Duration(c.Movetime-c.Overhead)
	default:
		movestogo := c.Movestogo
		t := c.Wtime
		inc := c.Winc
		if side == board.BLACK {
			t = c.Btime
			inc = c.Binc
		}
		movetime = time.Millisecond * time.Duration(Max((t+((inc-c.Overhead)*movestogo))/(movestogo+1)-c.Overhead, 10))
	}
	return movetime
}

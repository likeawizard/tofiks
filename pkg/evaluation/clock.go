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
	switch {
	case c.Movetime > 0:
		movetime = time.Millisecond * time.Duration(c.Movetime-c.Overhead)
	case c.Movestogo > 0:
		movestogo := c.Movestogo
		t := c.Wtime
		inc := c.Binc
		if side == board.BLACK {
			t = c.Btime
			inc = c.Binc
		}
		movetime = time.Millisecond * time.Duration((t+((inc-c.Overhead)*movestogo))/(movestogo+1))
	case c.Movestogo == 0:
		movestogo := Max(40-int(fmCounter), 10)
		t := c.Wtime
		inc := c.Binc
		if side == board.BLACK {
			t = c.Btime
			inc = c.Binc
		}
		movetime = time.Millisecond * time.Duration((t+((inc-c.Overhead)*movestogo))/(movestogo+1))
	}
	return movetime
}

package uci

import (
	"context"
	"fmt"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

const (
	STATUS_UNAVIALABELE int = iota
	STATUS_IDLE
	SATUS_BUSY
)

type UCI_CMD string

const (
	C_UCI = "uci"
	// C_DEBUG      = "debug" // on | off
	C_IS_READY   = "isready"
	C_SET_OPTION = "setoption" // name [value]
	C_POSITION   = "position"  // [ fen | startpos] moves ...
	C_GO         = "go"        // many opts
	C_STOP       = "stop"
	C_QUIT       = "quit"
	//TODO: commands that are defined by the uci protocol but not implemented
	//ponderhi, ucinewgame, register
)

type UCIOpts struct {
	Hash    int
	Ponder  bool
	OwnBook bool
}

type UCICmd interface {
	Exec(*eval.EvalEngine) bool
}

type UCI struct {
}

func (c *UCI) Exec(e *eval.EvalEngine) bool {
	fmt.Println("id name Tofiks 0.0.1")
	fmt.Println("id author Aturs Priede")
	//TODO: add available opts
	fmt.Println("uciok")
	return true
}

type IsReady struct {
}

func (c *IsReady) Exec(e *eval.EvalEngine) bool {
	e.MU.Lock()
	defer e.MU.Unlock()
	fmt.Println("readyok")
	return true
}

type Position struct {
	pos   string
	moves string
}

func (c *Position) Exec(e *eval.EvalEngine) bool {
	e.MU.Lock()
	defer e.MU.Unlock()
	e.Board = board.NewBoard(c.pos)
	return e.Board.PlayMovesUCI(c.moves)
}

type Quit struct {
}

func (c *Quit) Exec(e *eval.EvalEngine) bool {
	e.Quit = true
	return true
}

type Stop struct {
}

func (c Stop) Exec(e *eval.EvalEngine) bool {
	if e.Stop != nil {
		e.Stop()
	}
	return true
}

type Go struct {
	wtime     int
	btime     int
	binc      int
	winc      int
	depth     int
	movetime  int
	movestogo int
	infinite  bool
}

func (c Go) Exec(e *eval.EvalEngine) bool {
	var ctx context.Context
	var cancel context.CancelFunc
	switch {
	case c.infinite:
		ctx, cancel = context.WithCancel(context.Background())
	case c.movetime > 0:
		movetime := time.Millisecond * time.Duration(c.movetime)
		ctx, cancel = context.WithTimeout(context.Background(), movetime)
	case c.movestogo > 0:
		t := c.wtime
		inc := c.binc
		if e.Board.Side == board.BLACK {
			t = c.btime
			inc = c.binc
		}
		movetime := time.Millisecond * time.Duration((t+inc*c.movestogo)/c.movestogo)
		ctx, cancel = context.WithTimeout(context.Background(), movetime)
	default:
		movetime := time.Millisecond * time.Duration(c.movetime)
		ctx, cancel = context.WithTimeout(context.Background(), movetime)
	}
	var depth = c.depth
	if depth == 0 {
		depth = 50
	}
	e.Stop = cancel
	defer cancel()
	move, ponder := e.GetMove(ctx, depth)
	fmt.Printf("bestmove %s ponder %s\n", move, ponder)
	return true
}

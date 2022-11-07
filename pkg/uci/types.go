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
}

func (c Go) Exec(e *eval.EvalEngine) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	e.Stop = cancel
	defer cancel()
	pv := []board.Move{}
	move, ponder := e.GetMove(ctx, &pv, false)
	fmt.Printf("bestmove %s ponder %s\n", move, ponder)
	return true
}

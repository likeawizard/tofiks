package uci

import (
	"fmt"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/book"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

func (c *Go) Exec(e *eval.EvalEngine) bool {
	e.Clock.Wtime = c.wtime
	e.Clock.Winc = c.winc
	e.Clock.Btime = c.btime
	e.Clock.Binc = c.binc
	e.Clock.Movestogo = c.movestogo
	e.Clock.Movetime = c.movetime
	e.Clock.Infinite = c.infinite
	ctx, cancel := e.Clock.GetContext(int(e.Board.FullMoveCounter), e.Board.Side)
	var depth = c.depth
	if depth == 0 {
		depth = 50
	}
	e.Stop = cancel
	defer cancel()
	move, ponder := e.GetMove(ctx, depth)
	if ponder != 0 {
		fmt.Printf("bestmove %s\n", move)
	} else {
		fmt.Printf("bestmove %s\n", move)
	}

	return true
}

func (c *Stop) Exec(e *eval.EvalEngine) bool {
	if e.Stop != nil {
		e.Stop()
	}
	return true
}

func (c *Quit) Exec(e *eval.EvalEngine) bool {
	e.Quit = true
	return true
}

func (c *Position) Exec(e *eval.EvalEngine) bool {
	e.MU.Lock()
	defer e.MU.Unlock()
	e.Board = board.NewBoard(c.pos)
	return e.Board.PlayMovesUCI(c.moves)
}

func (c MoveOverhead) Exec(e *eval.EvalEngine) bool {
	e.MU.Lock()
	defer e.MU.Unlock()
	return true
}

func (c *IsReady) Exec(e *eval.EvalEngine) bool {
	e.MU.Lock()
	defer e.MU.Unlock()
	fmt.Println("readyok")
	return true
}

func (c *UCI) Exec(e *eval.EvalEngine) bool {
	availOpts := []UCIOpt{&Ponder{}, &Hash{}, &Clear{}, &MoveOverhead{}, &OwnBook{}}
	fmt.Println("id name Tofiks 0.0.1")
	fmt.Println("id author Aturs Priede")
	for _, opt := range availOpts {
		opt.Info()
	}
	fmt.Println("uciok")
	return true
}

func (c *SetOption) Exec(e *eval.EvalEngine) bool {
	e.MU.Lock()
	defer e.MU.Unlock()
	c.option.Set(e)
	return true
}

func (c *NewGame) Exec(e *eval.EvalEngine) bool {
	e.MU.Lock()
	defer e.MU.Unlock()
	e.TTable.Clear()
	e.KillerMoves = [100][2]board.Move{}
	e.GameHistory = [512]uint64{}
	return true
}

func (o *Hash) Set(e *eval.EvalEngine) {
	e.TTable = eval.NewTTable(o.size)
	fmt.Println("debug tt size", o.size)
}

func (o *Hash) Info() {
	fmt.Println("option name Hash type spin default 64 min 1 max 256")
}

func (o *OwnBook) Set(e *eval.EvalEngine) {
	e.OwnBook = o.enable
}

func (o *OwnBook) Info() {
	if book.LoadBook("book.bin") > 0 {
		fmt.Println("option name OwnBook type check default false")
	}
}

func (o *Ponder) Set(e *eval.EvalEngine) {
	//TODO - implement ponder and ponderhit functionality
}

func (o *Ponder) Info() {
	fmt.Println("option name Ponder type check default false")
}

func (o *Clear) Set(e *eval.EvalEngine) {
	e.TTable.Clear()
}

func (o *Clear) Info() {
	fmt.Println("option name Clear Hash type button")
}

func (o *MoveOverhead) Set(e *eval.EvalEngine) {
	fmt.Println("yeeer")
	e.Clock.Overhead = o.delay
}

func (o *MoveOverhead) Info() {
	fmt.Println("option name Move Overhead type spin default 0 min 0 max 1000")
}

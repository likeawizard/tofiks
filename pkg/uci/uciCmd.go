package uci

import (
	"fmt"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/book"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

func (c *Go) Exec(e *eval.Engine) bool {
	// Check if internal state is ready - should be done by gui
	e.WG.Wait()
	if c.isPerft {
		e.Board.PerftDebug(c.depth)
		return true
	}

	e.Clock.Wtime = c.wtime
	e.Clock.Winc = c.winc
	e.Clock.Btime = c.btime
	e.Clock.Binc = c.binc
	e.Clock.Movestogo = c.movestogo
	e.Clock.Movetime = c.movetime
	e.Clock.Infinite = c.infinite
	ctx, cancel := e.Clock.GetContext(int(e.Board.FullMoveCounter), e.Board.Side)
	depth := c.depth
	if depth == 0 {
		depth = 50
	}
	e.Stop = cancel
	defer cancel()
	move, ponder := e.GetMove(ctx, depth, c.infinite)
	e.ReportMove(move, ponder, e.Ponder)

	return true
}

func (c *Stop) Exec(e *eval.Engine) bool {
	defer e.WG.Done()
	if e.Stop != nil {
		// If we scored a ponderhit think for 1/3rd of the normal time unless mate has been already found
		if c.ponderhit && !e.MateFound {
			time.Sleep(e.Clock.GetMovetime(int(e.Board.FullMoveCounter), e.Board.Side) / 3)
		}
		e.Stop()
	}
	return true
}

func (c *Quit) Exec(_ *eval.Engine) bool {
	return true
}

func (c *Position) Exec(e *eval.Engine) bool {
	defer e.WG.Done()
	e.Board = board.NewBoard(c.pos)
	return e.PlayMovesUCI(c.moves)
}

func (c *IsReady) Exec(e *eval.Engine) bool {
	e.WG.Wait()
	fmt.Println("readyok")
	return true
}

func (c *UCI) Exec(e *eval.Engine) bool {
	defer e.WG.Done()
	availOpts := []Opt{&Ponder{}, &Hash{}, &Clear{}, &MoveOverhead{}, &OwnBook{}}
	fmt.Println("id name Tofiks v1.3.0")
	fmt.Println("id author Aturs Priede")
	for _, opt := range availOpts {
		opt.Info()
	}
	fmt.Println("uciok")
	return true
}

func (c *SetOption) Exec(e *eval.Engine) bool {
	defer e.WG.Done()
	c.option.Set(e)
	return true
}

func (c *NewGame) Exec(e *eval.Engine) bool {
	defer e.WG.Done()
	e.TTable.Clear()
	e.KillerMoves = [100][2]board.Move{}
	e.Plys = [512]uint64{}
	e.History = eval.HistoryHeuristic{}
	return true
}

func (o *Hash) Set(e *eval.Engine) {
	e.TTable = eval.NewTTable(o.size)
}

func (o *Hash) Info() {
	fmt.Println("option name Hash type spin default 64 min 1 max 256")
}

func (o *OwnBook) Set(e *eval.Engine) {
	e.OwnBook = o.enable
}

func (o *OwnBook) Info() {
	if book.LoadBook("book.bin") > 0 {
		fmt.Println("option name OwnBook type check default false")
	}
}

func (o *Ponder) Set(e *eval.Engine) {
	e.Ponder = o.enable
}

func (o *Ponder) Info() {
	fmt.Println("option name Ponder type check default false")
}

func (o *Clear) Set(e *eval.Engine) {
	e.TTable.Clear()
}

func (o *Clear) Info() {
	fmt.Println("option name Clear Hash type button")
}

func (o *MoveOverhead) Set(e *eval.Engine) {
	e.Clock.Overhead = o.delay
}

func (o *MoveOverhead) Info() {
	fmt.Println("option name Move Overhead type spin default 0 min 0 max 1000")
}

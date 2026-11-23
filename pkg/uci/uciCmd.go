package uci

import (
	"fmt"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/book"
	"github.com/likeawizard/tofiks/pkg/search"
)

// Exec installs the clock and time control, then runs the search on its own
// goroutine. The time control exists before that goroutine does, so a stop or
// ponderhit arriving on the reader can only ever address this search.
func (c *Go) Exec(e *search.Engine) bool {
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
	e.Clock.HasClock = c.hasClock
	depth := c.depth
	if depth == 0 {
		depth = 50
	}
	e.TC = e.Clock.NewTimeControl(int(e.Board.FullMoveCounter), e.Board.Side)
	e.WG.Add(1)
	go func() {
		defer e.WG.Done()
		defer e.TC.Stop()
		move, ponder := e.GetMove(depth, c.infinite)
		e.ReportMove(move, ponder, e.Ponder)
	}()

	return true
}

// Exec handles stop and ponderhit. Both act on the running search's time
// control without waiting for it.
func (c *Stop) Exec(e *search.Engine) bool {
	if c.ponderhit {
		e.TC.PonderHit()
	} else {
		e.TC.Abort()
	}
	return true
}

func (c *Quit) Exec(_ *search.Engine) bool {
	return true
}

func (c *Position) Exec(e *search.Engine) bool {
	e.Board = board.NewBoard(c.pos)
	return e.PlayMovesUCI(c.moves)
}

func (c *IsReady) Exec(_ *search.Engine) bool {
	fmt.Println("readyok")
	return true
}

func (c *UCI) Exec(_ *search.Engine) bool {
	availOpts := []Opt{&Ponder{}, &Hash{}, &Clear{}, &MoveOverhead{}, &OwnBook{}}
	fmt.Println("id name Tofiks v1.6.0")
	fmt.Println("id author Arturs Priede")
	for _, opt := range availOpts {
		opt.Info()
	}
	fmt.Println("uciok")
	return true
}

func (c *SetOption) Exec(e *search.Engine) bool {
	c.option.Set(e)
	return true
}

func (c *NewGame) Exec(e *search.Engine) bool {
	e.TTable.Clear()
	e.Eval.PawnTable.Clear()
	e.KillerMoves = [100][2]board.Move{}
	e.Plys = [512]uint64{}
	e.History = search.HistoryHeuristic{}
	return true
}

func (o *Hash) Set(e *search.Engine) {
	e.TTable = search.NewTTable(o.size)
}

func (o *Hash) Info() {
	fmt.Println("option name Hash type spin default 64 min 1 max 256")
}

func (o *OwnBook) Set(e *search.Engine) {
	e.OwnBook = o.enable
}

func (o *OwnBook) Info() {
	if book.LoadBook("book.bin") > 0 {
		fmt.Println("option name OwnBook type check default false")
	}
}

func (o *Ponder) Set(e *search.Engine) {
	e.Ponder = o.enable
}

func (o *Ponder) Info() {
	fmt.Println("option name Ponder type check default false")
}

func (o *Clear) Set(e *search.Engine) {
	e.TTable.Clear()
	e.Eval.PawnTable.Clear()
}

func (o *Clear) Info() {
	fmt.Println("option name Clear Hash type button")
}

func (o *MoveOverhead) Set(e *search.Engine) {
	e.Clock.Overhead = o.delay
}

func (o *MoveOverhead) Info() {
	fmt.Println("option name Move Overhead type spin default 0 min 0 max 1000")
}

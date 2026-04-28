package uci

import (
	"github.com/likeawizard/tofiks/pkg/search"
)

const (
	CmdUci = "uci"
	// CmdDebug      = "debug" // on | off.
	CmdIsReady   = "isready"
	CmdSetOption = "setoption" // name [value]
	CmdPosition  = "position"  // [ fen | startpos] moves ...
	CmdGo        = "go"        // many opts
	CmdStop      = "stop"
	CmdPonderhit = "ponderhit"
	CmdQuit      = "quit"
	CmdNewGame   = "ucinewgame"
)

type Cmd interface {
	Exec(*search.Engine) bool
}

type UCI struct{}

type IsReady struct{}

type Position struct {
	pos   string
	moves string
}

type Quit struct{}

type Stop struct {
	ponderhit bool
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
	isPerft   bool
}

type SetOption struct {
	option Opt
}

type NewGame struct{}

type Opt interface {
	Info()
	Set(e *search.Engine)
}

type Hash struct {
	size int
}

type Ponder struct {
	enable bool
}

type OwnBook struct {
	enable bool
}

type Clear struct{}

type MoveOverhead struct {
	delay int
}

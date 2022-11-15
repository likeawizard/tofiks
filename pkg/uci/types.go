package uci

import (
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
	C_NEW_GAME   = "ucinewgame"
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

type UCI struct{}

type IsReady struct{}

type Position struct {
	pos   string
	moves string
}

type Quit struct{}

type Stop struct{}

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

type SetOption struct {
	option UCIOpt
}

type NewGame struct{}

type UCIOpt interface {
	Info()
	Set(e *eval.EvalEngine)
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

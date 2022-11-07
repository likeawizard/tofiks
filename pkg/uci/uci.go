package uci

import (
	"regexp"
)

//TODO: stub, implement the UCI interface

// Parse uci command and return executable UCICmd on successful parse or nil
func ParseUCI(uciCmd string) UCICmd {
	cmdRE := regexp.MustCompile(`(?P<cmd>^\w+)\s?(?P<args>.*)`)
	if !cmdRE.MatchString(uciCmd) {
		return nil
	}
	match := cmdRE.FindStringSubmatch(uciCmd)
	cmd := match[cmdRE.SubexpIndex("cmd")]
	args := match[cmdRE.SubexpIndex("args")]

	switch cmd {
	case C_UCI:
		return &UCI{}
	case C_IS_READY:
		return &IsReady{}
	case C_STOP:
		return &Stop{}
	case C_QUIT:
		return &Quit{}
	case C_POSITION:
		pos := Position{}
		posRE := regexp.MustCompile(`(startpos|(fen\s(?P<fen>.+?)))\s*(?:$|moves\s)(?P<moves>.*)?`)
		match = posRE.FindStringSubmatch(args)
		fen := match[posRE.SubexpIndex("fen")]
		moves := match[posRE.SubexpIndex("moves")]
		pos.pos = fen
		pos.moves = moves
		return &pos
	case C_SET_OPTION:
		return nil
	case C_GO:
		return &Go{}
	}

	return nil
}

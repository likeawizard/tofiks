package uci

import (
	"regexp"
	"strconv"
	"strings"
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
	case C_PONDERHIT:
		return &Stop{ponderhit: true}
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
		opt := SetOption{}
		optRE := regexp.MustCompile(`name\s(?P<name>[\w\s]+)\svalue\s(?P<value>\w+)`)
		match = optRE.FindStringSubmatch(args)
		name := match[optRE.SubexpIndex("name")]
		value := match[optRE.SubexpIndex("value")]
		switch name {
		case "Ponder":
			enable := false
			if value == "true" {
				enable = true
			}
			opt.option = &Ponder{enable: enable}
			return &opt
		case "Hash":
			size, _ := strconv.Atoi(value)
			opt.option = &Hash{size: size}
			return &opt
		case "OwnBook":
			enable := false
			if value == "true" {
				enable = true
			}
			opt.option = &OwnBook{enable: enable}
			return &opt
		case "Clear Hash":
			opt.option = &Clear{}
			return &opt
		case "Move Overhead":
			delay, _ := strconv.Atoi(value)
			opt.option = &MoveOverhead{delay: delay}
			return &opt
		}
		return nil
	case C_GO:
		goCmd := Go{}
		goParts := strings.Fields(args)
		for i, s := range goParts {
			switch s {
			case "wtime":
				goCmd.wtime, _ = strconv.Atoi(goParts[i+1])
			case "btime":
				goCmd.btime, _ = strconv.Atoi(goParts[i+1])
			case "winc":
				goCmd.winc, _ = strconv.Atoi(goParts[i+1])
			case "binc":
				goCmd.binc, _ = strconv.Atoi(goParts[i+1])
			case "movestogo":
				goCmd.movestogo, _ = strconv.Atoi(goParts[i+1])
			case "depth":
				goCmd.depth, _ = strconv.Atoi(goParts[i+1])
			case "movetime":
				goCmd.movetime, _ = strconv.Atoi(goParts[i+1])
			case "infinite", "ponder":
				goCmd.infinite = true
			case "perft": // non-uci command execute perft instead
				goCmd.isPerft = true
				goCmd.depth, _ = strconv.Atoi(goParts[i+1])
			}
		}
		return &goCmd
	case C_NEW_GAME:
		return &NewGame{}
	}

	return nil
}

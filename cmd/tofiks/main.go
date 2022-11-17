package main

import (
	"bufio"
	"fmt"
	"os"

	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/likeawizard/tofiks/pkg/uci"
)

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	e, err := eval.NewEvalEngine()
	if err != nil {
		fmt.Printf("Unable to load EvalEngine: %s\n", err)
		return
	}

	input := bufio.NewScanner(os.Stdin)
	for {
		input.Scan()
		cmd := uci.ParseUCI(input.Text())
		if cmd != nil {
			switch cmd.(type) {
			case *uci.Quit:
				return
			case *uci.Go, *uci.IsReady:
				go cmd.Exec(e)
			default:
				e.WG.Add(1)
				go cmd.Exec(e)
			}
		}
	}
}

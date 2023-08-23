package main

import (
	"bufio"
	"os"

	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/likeawizard/tofiks/pkg/uci"
)

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	e := eval.NewEvalEngine()

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
				e.WG.Wait()
				e.WG.Add(1)
				go cmd.Exec(e)
			}
		}
	}
}

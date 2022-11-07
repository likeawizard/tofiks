package main

import (
	"bufio"
	"fmt"
	"os"

	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/likeawizard/tofiks/pkg/uci"
	_ "go.uber.org/automaxprocs"
)

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	e, err := eval.NewEvalEngine()
	if err != nil {
		fmt.Printf("Unable to load EvalEngine: %s\n", err)
		return
	}

	input := bufio.NewScanner(os.Stdin)
	for !e.Quit {
		input.Scan()
		cmd := uci.ParseUCI(input.Text())
		if cmd != nil {
			go cmd.Exec(e)
		}
	}
}

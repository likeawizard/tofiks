package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"runtime/pprof"

	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/likeawizard/tofiks/pkg/uci"
	"github.com/pkg/profile"
	_ "go.uber.org/automaxprocs"
)

func main() {
	enableProfile := false
	enableMemProf := false
	flag.BoolVar(&enableProfile, "pgo", false, "Enable CPU profiling")
	flag.BoolVar(&enableMemProf, "memprof", false, "Enable memory profiling")
	flag.Parse()
	if enableProfile {
		f, err := os.Create("cmd/tofiks/default.pgo")
		if err != nil {
			log.Printf("Error creating profile file: %v", err)
			return
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Printf("Error starting profile: %v", err)
			return
		}
		defer pprof.StopCPUProfile()
	}

	if enableMemProf {
		defer profile.Start(profile.MemProfile, profile.ProfilePath("cmd/tofiks/")).Stop()
	}
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

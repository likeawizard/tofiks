package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

const K = 0.477

type testPosition struct {
	result float64
	fen    string
}

var tests []testPosition

// const int nParams = initialGuess.size();
//    double bestE = E(initialGuess);
//    vector<int> bestParValues = initialGuess;
//    bool improved = true;
//    while ( improved ) {
//       improved = false;
//       for (int pi = 0; pi < nParams; pi++) {
//          vector<int> newParValues = bestParValues;
//          newParValues[pi] += 1;
//          double newE = E(newParValues);
//          if (newE < bestE) {
//             bestE = newE;
//             bestParValues = newParValues;
//             improved = true;
//          } else {
//             newParValues[pi] -= 2;
//             newE = E(newParValues);
//             if (newE < bestE) {
//                bestE = newE;
//                bestParValues = newParValues;
//                improved = true;
//             }
//          }
//       }
//    }
//    return bestParValues;

func main() {
	var dataPath string
	flag.StringVar(&dataPath, "data", "", "Texel training data")
	flag.Parse()
	load(dataPath)

	bestE := E()
	startE := bestE
	fmt.Println("start", bestE)
	bestPST := eval.PST
	improved := true
	for improved {
		improved = false
		newPST := eval.PST
		for stage := 0; stage < 2; stage++ {
			for color := 0; color < 2; color++ {
				for piece := 0; piece < 6; piece++ {
					for sqare := 0; sqare < 64; sqare++ {
						newPST[stage][color][piece][sqare] += 1
						eval.PST = newPST
						newE := E()
						if newE < bestE {
							fmt.Printf("Improved: %e (curr: %f start:%f)\n", bestE-newE, newE, startE)
							bestE = newE
							bestPST = newPST
							improved = true
						} else {
							newPST[stage][color][piece][sqare] -= 2
							eval.PST = newPST
							newE = E()
							if newE < bestE {
								fmt.Printf("Improved: %e (curr: %f start:%f)\n", bestE-newE, newE, startE)
								bestE = newE
								bestPST = newPST
								improved = true
							} else {
								fmt.Printf("Declined: %e (curr: %f start:%f)\n", bestE-newE, newE, startE)
								eval.PST = bestPST
							}
						}
					}
				}
			}
		}
	}
	fmt.Println("end", bestE)

	fmt.Println(eval.PST)
}

func load(path string) {
	tests = make([]testPosition, 0)
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("failed to open data file")
	}
	defer f.Close()
	r := bufio.NewReader(f)

EOF:
	for err == nil {

		var lineBytes []byte
		var line string
		isPrefix := true
		for isPrefix {
			lineBytes, isPrefix, err = r.ReadLine()
			if err != nil {
				break EOF
			}
			line += string(lineBytes)
		}
		cont := strings.SplitN(line, " ", 2)
		result, _ := strconv.ParseFloat(cont[0], 64)
		fen := cont[1]
		tests = append(tests, testPosition{result: result, fen: fen})
	}
}

func sgim(eval int32) float64 {
	return 1 / (1 + math.Pow(10, -K*float64(eval)/400))
}

func E() float64 {
	e, _ := eval.NewEvalEngine()
	ctx := context.Background()
	E := 0.0
	gameCount := 0
	for i := range tests {
		e.Board = &board.Board{}
		impErr := e.Board.ImportFEN(tests[i].fen)
		if impErr != nil {
			fmt.Println(impErr)
			continue
		}
		color := int32(1)
		if e.Board.Side != board.WHITE {
			color = -color
		}
		q := color * e.Quiescence(ctx, -eval.Inf, eval.Inf, color)
		s := sgim(q)
		E += (tests[i].result - s) * (tests[i].result - s)
		gameCount++
	}
	return E / float64(gameCount)
}

// func E() float64 {
// 	ctx := context.Background()
// 	E := 0.0
// 	eChan := make(chan float64, 16)
// 	qChan := make(chan bool, 16)
// 	gameCount := 0

// 	var wgCount sync.WaitGroup
// 	wgCount.Add(1)
// 	go func() {
// 		for e := range eChan {
// 			E += e
// 			gameCount++
// 		}
// 		wgCount.Done()
// 	}()

// 	for i := range tests {
// 		qChan <- true
// 		go func(t testPosition) {
// 			e, _ := eval.NewEvalEngine()
// 			e.Board = &board.Board{}
// 			err := e.Board.ImportFEN(t.fen)
// 			fmt.Println(t.fen)
// 			if err != nil {
// 				<-qChan
// 				return
// 			}
// 			color := int32(1)
// 			if e.Board.Side != board.WHITE {
// 				color = -color
// 			}
// 			q := color * e.Quiescence(ctx, -eval.Inf, eval.Inf, color)
// 			s := sgim(q)
// 			eChan <- (t.result - s) * (t.result - s)
// 			<-qChan
// 		}(tests[i])
// 	}
// 	close(eChan)
// 	fmt.Println("closed")
// 	wgCount.Wait()
// 	fmt.Println("completed")
// 	return E / float64(gameCount)
// }

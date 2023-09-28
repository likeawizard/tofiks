package testsuite

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/stretchr/testify/assert"
)

// Test validity of returned PV - bad TT returns or incorrect parsing of the line can cause wrong/illegal moves to leak into the output.
func TestValidPV(t *testing.T) {
	testPositions := []string{
		"startpos",
		"8/7k/8/3p4/8/1p2P3/1P2P3/7K w - - 0 1",
		"r1bq1rk1/bpp1nppp/3p1n2/p3p3/4P3/1BPP1N1P/PP1N1PP1/R1BQ1RK1 w - - 1 10",
		"r5k1/bbq2r2/1p1pR1pB/pBpP1pN1/P7/3Q2NP/5PP1/6K1 b - - 4 30",
	}

	capturePV := regexp.MustCompile(`pv (?P<pv>.*)`)
	for i, testPos := range testPositions {
		t.Run(fmt.Sprintf("Test Position %d", i), func(t *testing.T) {
			r, w, _ := os.Pipe()
			os.Stdout = w

			e := eval.NewEvalEngine()
			e.Board = board.NewBoard(testPos)
			position := e.Board.Copy()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			lr := bufio.NewScanner(r)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer w.Close()
				defer wg.Done()
				e.IDSearch(ctx, 50, false)
			}()

			for lr.Scan() {
				info := lr.Text()
				m := capturePV.FindStringSubmatch(info)
				pv := m[capturePV.SubexpIndex("pv")]
				tmpPos := position.Copy()
				pvMoves := strings.Fields(pv)

				for _, uciMove := range pvMoves {
					_, ok := tmpPos.MoveUCI(uciMove)
					assert.True(t, ok, "error parsing PV, position '%v' pv: %v on move: %v\n", position.ExportFEN(), info, uciMove)
				}
			}
			wg.Wait()
			cancel()
			r.Close()
		})
	}
}

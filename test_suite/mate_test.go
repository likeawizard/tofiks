package testsuite

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/stretchr/testify/assert"
)

// Find mates in positions. Ensure correct mate depth is found and that between iterative deepening the number does not go up.
//
// Mate depth can go down as sacrificial lines without immediate pay-off could be pruned or reduced in depth.
// A longer but 'obvious' mating sequence can be found first and a 'less obvious' but shorter later. It should never go up!
func TestMate(t *testing.T) {
	capturePV := regexp.MustCompile(`mate (?P<mate>-?\d+)`)
	for i, testPos := range matePositions {
		t.Run(fmt.Sprintf("Position #%d Mate in %d", i, testPos.mateIn), func(t *testing.T) {
			r, w, _ := os.Pipe()
			os.Stdout = w

			e := eval.NewEvalEngine()
			e.Board = board.NewBoard(testPos.fen)
			e.Clock = eval.Clock{
				Movetime: 1000 * 60,
			}
			lr := bufio.NewScanner(r)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer w.Close()
				defer wg.Done()
				e.IDSearch(50, true)
			}()

			mateInMin := 100
			if testPos.mateIn < 0 {
				mateInMin = -100
			}
			for lr.Scan() {
				info := lr.Text()
				m := capturePV.FindStringSubmatch(info)
				if m == nil {
					continue
				}
				mateStr := m[capturePV.SubexpIndex("mate")]
				mateIn, err := strconv.Atoi(mateStr)
				assert.Nil(t, err, "failed to parse mate score to int: %s", err)
				if testPos.mateIn > 0 {
					assert.LessOrEqual(t, mateIn, mateInMin, "mate score increased with depth was %d now %d", mateInMin, mateIn)
					mateInMin = min(mateInMin, mateIn)
				} else {
					assert.GreaterOrEqual(t, mateIn, mateInMin, "mate score increased with depth was %d now %d", mateInMin, mateIn)
					mateInMin = max(mateInMin, mateIn)
				}
				if mateInMin == testPos.mateIn {
					e.Stop <- struct{}{}
				}
			}
			wg.Wait()
			e.Stop <- struct{}{}
			assert.Equal(t, testPos.mateIn, mateInMin)
			r.Close()
		})
	}
}

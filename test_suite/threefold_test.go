package testsuite

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/stretchr/testify/assert"
)

func TestThreefoldRepetition(t *testing.T) {
	for _, testPos := range drawByThreefoldPositions {
		t.Run(fmt.Sprintf("Position %d", testPos.number), func(t *testing.T) {
			e := eval.NewEvalEngine()
			e.Board = board.NewBoard(testPos.fen)
			ok := e.PlayMovesUCI(testPos.moves)
			assert.True(t, ok, "Failed to play out moves")
			assert.True(t, e.IsDrawByRepetition(), "Failed to detect draw by repetition")
		})
	}
}

func TestForceThreeFoldRepetition(t *testing.T) {
	captureScore := regexp.MustCompile(`cp (?P<cp>-?\d+)`)
	score := 1
	var err error
	for _, testPos := range forceThreeFoldPositions {
		t.Run(fmt.Sprintf("Position %d", testPos.number), func(t *testing.T) {
			r, w, _ := os.Pipe()
			os.Stdout = w

			e := eval.NewEvalEngine()
			e.Board = board.NewBoard(testPos.fen)
			e.PlayMovesUCI(testPos.moves)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			lr := bufio.NewScanner(r)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer w.Close()
				defer wg.Done()
				e.IDSearch(ctx, 50, true)
			}()

			for lr.Scan() {
				info := lr.Text()
				m := captureScore.FindStringSubmatch(info)
				if m == nil {
					continue
				}
				scoreStr := m[captureScore.SubexpIndex("cp")]
				score, err = strconv.Atoi(scoreStr)
				assert.Nil(t, err, "failed to parse cp score: %s", err)
				if score == 0 {
					cancel()
				}
			}
			wg.Wait()
			cancel()
			assert.Equal(t, 0, score)
			r.Close()
		})
	}
}

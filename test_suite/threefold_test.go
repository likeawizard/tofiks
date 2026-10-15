package testsuite

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/search"
	"github.com/stretchr/testify/assert"
)

func TestThreefoldRepetition(t *testing.T) {
	for _, testPos := range drawByThreefoldPositions {
		t.Run(fmt.Sprintf("Position %d", testPos.number), func(t *testing.T) {
			e := search.NewEngine()
			e.Board = board.NewBoard(testPos.fen)
			ok := e.PlayMovesUCI(testPos.moves)
			assert.True(t, ok, "Failed to play out moves")
			assert.True(t, e.IsDrawByRepetition(), "Failed to detect draw by repetition")
		})
	}
}

func TestForceThreeFoldRepetition(t *testing.T) {
	// Repetition draws are dithered by ±1cp in search.
	const drawBand = 1
	captureScore := regexp.MustCompile(`cp (?P<cp>-?\d+)`)
	// Sentinel well outside the draw band so a test that never parses a score fails.
	score := 1000
	var err error
	for _, testPos := range forceThreeFoldPositions {
		t.Run(fmt.Sprintf("Position %d", testPos.number), func(t *testing.T) {
			r, w, _ := os.Pipe()
			os.Stdout = w

			e := search.NewEngine()
			e.Board = board.NewBoard(testPos.fen)
			e.PlayMovesUCI(testPos.moves)
			e.Clock.Movetime = int(time.Minute / time.Millisecond)
			e.TC = e.Clock.NewTimeControl(int(e.Board.FullMoveCounter), e.Board.Side)
			lr := bufio.NewScanner(r)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer w.Close()
				defer wg.Done()
				e.IDSearch(50, true)
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
				if score >= -drawBand && score <= drawBand {
					e.TC.Abort()
				}
			}
			wg.Wait()
			e.TC.Abort()
			assert.InDelta(t, 0, score, drawBand, "expected forced threefold to score as a draw")
			r.Close()
		})
	}
}

package testsuite

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
	"github.com/stretchr/testify/assert"
)

func FuzzEntry(f *testing.F) {
	f.Fuzz(func(t *testing.T, move uint32, depth int8, eType uint8, age int8, score int16) {
		if eType > 2 || depth < 0 || age < 0 {
			return
		}
		entry := eval.NewEntry(board.Move(move), depth, eval.EntryType(eType), age, score)
		assert.Equal(t, board.Move(move), entry.Move(), "move mismatch")
		assert.Equal(t, depth, entry.Depth(), "depth mismatch")
		assert.Equal(t, eval.EntryType(eType), entry.Type(), "type mismatch")
		assert.Equal(t, age, entry.Age(), "age mismatch")
		assert.Equal(t, score, entry.Score(), "score mismatch")
	})
}

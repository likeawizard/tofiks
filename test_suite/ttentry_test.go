package testsuite

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/search"
	"github.com/stretchr/testify/assert"
)

func TestMateScoreAcrossPlies(t *testing.T) {
	tests := []struct {
		name         string
		storePly     int
		retrievePly  int
		mateDistance int  // plies from the position to the mate leaf
		positive     bool // true = side to move delivers mate, false = side to move gets mated
	}{
		{
			name:         "white mates, retrieved at deeper ply",
			storePly:     3,
			retrievePly:  7,
			mateDistance: 2,
			positive:     true,
		},
		{
			name:         "white mates, retrieved at shallower ply",
			storePly:     8,
			retrievePly:  2,
			mateDistance: 4,
			positive:     true,
		},
		{
			name:         "black mates (negative score), retrieved at deeper ply",
			storePly:     2,
			retrievePly:  6,
			mateDistance: 3,
			positive:     false,
		},
		{
			name:         "black mates (negative score), retrieved at shallower ply",
			storePly:     10,
			retrievePly:  4,
			mateDistance: 1,
			positive:     false,
		},
		{
			name:         "same ply retrieval should be unchanged",
			storePly:     5,
			retrievePly:  5,
			mateDistance: 2,
			positive:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := search.NewTTable(1)
			hash := uint64(0xDEADBEEF)

			// Build the score as PVS would produce it at storePly.
			// Positive mate: CheckmateScore - (storePly + mateDistance)
			// Negative mate: -CheckmateScore + (storePly + mateDistance)
			mateLeafPly := int16(tc.storePly) + int16(tc.mateDistance)
			var storedScore, expectedScore int16
			if tc.positive {
				storedScore = search.CheckmateScore - mateLeafPly
				// At retrievePly the correct score: CheckmateScore - (retrievePly + mateDistance)
				expectedScore = search.CheckmateScore - int16(tc.retrievePly) - int16(tc.mateDistance)
			} else {
				storedScore = -search.CheckmateScore + mateLeafPly
				// At retrievePly the correct score: -CheckmateScore + (retrievePly + mateDistance)
				expectedScore = -search.CheckmateScore + int16(tc.retrievePly) + int16(tc.mateDistance)
			}

			tt.Store(hash, search.TT_EXACT, storedScore, 10, tc.storePly, 0)

			entry, ok := tt.Probe(hash)
			assert.True(t, ok, "TT probe should hit")

			score, ok := entry.GetScore(10, tc.retrievePly, -search.Inf, search.Inf)
			assert.True(t, ok, "exact entry with sufficient depth should return a score")
			assert.Equal(t, expectedScore, score,
				"mate score should be adjusted to reflect the new ply distance; expected %d but got %d",
				expectedScore, score,
			)
		})
	}
}

func FuzzEntry(f *testing.F) {
	f.Fuzz(func(t *testing.T, move uint32, depth int8, eType uint8, age int8, score int16) {
		if eType > 2 || depth < 0 || age < 0 {
			return
		}
		entry := search.NewEntry(board.Move(move), int(depth), search.EntryType(eType), age, score)
		assert.Equal(t, board.Move(move), entry.Move(), "move mismatch")
		assert.Equal(t, int(depth), entry.Depth(), "depth mismatch")
		assert.Equal(t, search.EntryType(eType), entry.Type(), "type mismatch")
		assert.Equal(t, age, entry.Age(), "age mismatch")
		assert.Equal(t, score, entry.Score(), "score mismatch")
	})
}

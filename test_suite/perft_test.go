package testsuite

import (
	"fmt"
	"log"
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/stretchr/testify/assert"
)

// Perft test known positions and validate correct node count.
//
// Perft also does internal health / sanity checks by re-validating updated and fully computed hashes.
func TestPerft(t *testing.T) {
	t.Parallel()
	maxDepth := 5
	if testing.Short() {
		maxDepth = 4
	}

	for _, tt := range perftResults {
		results := tt.getResultAtDepth(maxDepth)
		testName := fmt.Sprintf("%s at depth %d", tt.position, results.depth)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			nodes, time := board.Perft(tt.fen, results.depth)
			assert.Equal(t, results.nodes, nodes, "perft failed. got %d want %d", nodes, results.nodes)
			log.Printf("%s finished in %v", testName, time)
		})
	}
}

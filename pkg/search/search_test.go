package search

import (
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
)

// TestNoTTStoreAfterAbort checks that an aborted search stores nothing in the
// TT for the root position (issue #181).
//
// The game history seeded below makes the root position part of Plys, so any
// in-search recurrence of it is scored as a repetition draw before reaching a
// TT store. The only frame that can store the root hash is the root frame
// itself — which must not, because the search is aborted before it completes.
func TestNoTTStoreAfterAbort(t *testing.T) {
	e := NewEngine()
	if !e.PlayMovesUCI("e2e4 e7e5 g1f3 b8c6") {
		t.Fatal("failed to play setup moves")
	}
	// A depth-30 search cannot finish before the movetime hard deadline; the abort fires mid-tree.
	tc := (&Clock{Movetime: 25}).NewTimeControl(0, 1)
	defer tc.Stop()
	e.TC = tc

	line := []board.Move{}
	e.PVS([]board.Move{}, &line, 30, 0, -Inf, Inf, true, 1)

	if !tc.ShouldAbort() {
		t.Fatal("search finished before abort; test setup no longer aborts mid-tree")
	}
	if _, ok := e.TTable.Probe(e.Board.Hash); ok {
		t.Fatal("aborted search stored a TT entry for the root position")
	}
}

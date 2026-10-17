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

// TestQuietHistoryMalus checks that a quiet cutoff move gains history while the
// quiet moves searched before it lose the same amount, and that negative history
// never leaks into the 10-bit move score.
func TestQuietHistoryMalus(t *testing.T) {
	e := NewEngine()
	moves := e.Board.PseudoMoveGen()
	if len(moves) < 3 {
		t.Fatal("startpos move generation broken")
	}
	best, tried := moves[0], []board.Move{moves[1], moves[2], moves[0]}

	e.UpdateQuietHistory(3, best, tried)

	if got := e.GetHistory(best); got != 9 {
		t.Fatalf("cutoff move history = %d, want 9", got)
	}
	for _, m := range tried[:2] {
		if got := e.GetHistory(m); got != -9 {
			t.Fatalf("tried move history = %d, want -9", got)
		}
	}

	e.ScoreMoves(0, moves, nil, 0)
	for _, m := range moves {
		if s := int(m >> 22); s > scoreHistMax { // score lives in the top 10 bits
			t.Fatalf("quiet move %s scored %d above scoreHistMax; negative history wrapped", m.ClearScore(), s)
		}
	}
}

// TestQuiescenceInCheck checks that quiescence does not stand pat while in
// check: a checkmated side gets a mate score, not its static eval.
func TestQuiescenceInCheck(t *testing.T) {
	e := NewEngine()
	// White to move, checkmated by the queen on f2.
	e.Board = board.NewBoard("r1b1k1nr/pppp1ppp/2n5/2b1p3/2B1P3/P7/1PPP1qPP/RNBQK1NR w KQkq - 0 5")
	tc := (&Clock{Infinite: true}).NewTimeControl(0, 0)
	defer tc.Stop()
	e.TC = tc

	if !e.Board.InCheck {
		t.Fatal("test position is not a check")
	}
	got := e.Quiescence(1, -Inf, Inf, 1)
	want := int16(1) - CheckmateScore
	if got != want {
		t.Fatalf("quiescence in checkmate returned %d, want mate score %d", got, want)
	}
}

//go:build debug

package search

import (
	"fmt"

	"github.com/likeawizard/tofiks/pkg/board"
)

// Stability tracks search stability metrics across iterative deepening iterations.
type Stability struct {
	prevBestMove     board.Move
	prevEval         int16
	pvChanges        uint64
	aspirationOk     uint64
	aspirationFail   uint64
	aspReSearchNodes uint64
	lmrSearches      uint64
	lmrResearches    uint64
	scoreDeltaSum    uint64
	iterations       uint64
}

func (s *Stability) recordIteration(bestMove board.Move, eval int16) {
	if s.iterations > 0 {
		if bestMove != s.prevBestMove {
			s.pvChanges++
		}
		delta := int(eval) - int(s.prevEval)
		if delta < 0 {
			delta = -delta
		}
		s.scoreDeltaSum += uint64(delta)
	}
	s.prevBestMove = bestMove
	s.prevEval = eval
	s.iterations++
}

func (s *Stability) recordAspiration(failed bool) {
	if failed {
		s.aspirationFail++
	} else {
		s.aspirationOk++
	}
}

// recordAspirationReSearch records the node cost of an aspiration re-search —
// how many nodes were consumed walking the tree again after a fail at the
// narrow window. This is the metric fail-soft TT probes are supposed to
// reduce, because warm TT entries should prove cutoffs at a wider re-search.
func (s *Stability) recordAspirationReSearch(nodes uint64) {
	s.aspReSearchNodes += nodes
}

// recordLMR records the outcome of an LMR-reduced null-window search.
// `failedHigh` is true when the reduced search returned value > alpha — i.e.
// the move beat the null window, meaning the reduction was potentially too
// aggressive. (For nested null-window searches, beta == alpha+1, so a true
// "in-band" re-search is impossible — `failedHigh` is the only meaningful
// signal of LMR error rate.)
func (s *Stability) recordLMR(failedHigh bool) {
	s.lmrSearches++
	if failedHigh {
		s.lmrResearches++
	}
}

func (s *Stability) reset() { *s = Stability{} }

func (s *Stability) String() string {
	if s.iterations == 0 {
		return ""
	}
	avgDelta := uint64(0)
	if s.iterations > 1 {
		avgDelta = s.scoreDeltaSum / (s.iterations - 1)
	}
	aspTotal := s.aspirationOk + s.aspirationFail
	aspFailRate := uint64(0)
	if aspTotal > 0 {
		aspFailRate = (100 * s.aspirationFail) / aspTotal
	}
	aspAvgReNodes := uint64(0)
	if s.aspirationFail > 0 {
		aspAvgReNodes = s.aspReSearchNodes / s.aspirationFail
	}
	lmrRate := uint64(0)
	if s.lmrSearches > 0 {
		lmrRate = (100 * s.lmrResearches) / s.lmrSearches
	}
	return fmt.Sprintf("stability: pv_changes %d avg_delta %dcp asp_fail %d%% (%d/%d) asp_re_nodes %d (avg %d) lmr_re %d%% (%d/%d)",
		s.pvChanges, avgDelta,
		aspFailRate, s.aspirationFail, aspTotal,
		s.aspReSearchNodes, aspAvgReNodes,
		lmrRate, s.lmrResearches, s.lmrSearches,
	)
}

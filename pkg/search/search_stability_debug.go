//go:build debug

package search

import (
	"fmt"

	"github.com/likeawizard/tofiks/pkg/board"
)

// Stability tracks search stability metrics across iterative deepening iterations.
type Stability struct {
	prevBestMove   board.Move
	prevEval       int16
	pvChanges      uint64
	aspirationOk   uint64
	aspirationFail uint64
	lmrSearches    uint64
	lmrResearches  uint64
	scoreDeltaSum  uint64
	iterations     uint64
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

func (s *Stability) recordLMR(researched bool) {
	s.lmrSearches++
	if researched {
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
	lmrRate := uint64(0)
	if s.lmrSearches > 0 {
		lmrRate = (100 * s.lmrResearches) / s.lmrSearches
	}
	return fmt.Sprintf("stability: pv_changes %d avg_delta %dcp asp_fail %d%% (%d/%d) lmr_re %d%% (%d/%d)",
		s.pvChanges, avgDelta,
		aspFailRate, s.aspirationFail, aspTotal,
		lmrRate, s.lmrResearches, s.lmrSearches,
	)
}

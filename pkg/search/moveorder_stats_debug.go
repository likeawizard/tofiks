//go:build debug

package search

import "fmt"

// MoveOrderStats tracks move ordering efficiency per search iteration.
type MoveOrderStats struct {
	failHighs      uint64
	failHighFirsts uint64
	pvsReSearches  uint64
}

func (s *MoveOrderStats) recordFailHigh(first bool) {
	s.failHighs++
	if first {
		s.failHighFirsts++
	}
}

// recordPVSReSearch records a null-window → full-window re-search in the PVS
// main loop, triggered when a null-window child returned a value in
// (alpha, beta). Under fail-soft TT cutoffs, null-window fails can return
// directly beta-cutoff-worthy values, skipping this re-search — which is
// exactly one of the mechanisms we're trying to measure.
func (s *MoveOrderStats) recordPVSReSearch() {
	s.pvsReSearches++
}

func (s *MoveOrderStats) reset() { *s = MoveOrderStats{} }

func (s *MoveOrderStats) String() string {
	if s.failHighs == 0 {
		return ""
	}
	rate := (100 * s.failHighFirsts) / s.failHighs
	return fmt.Sprintf("ordering: fh %d fhf %d%% (%d/%d) pvs_re %d",
		s.failHighs, rate, s.failHighFirsts, s.failHighs, s.pvsReSearches)
}

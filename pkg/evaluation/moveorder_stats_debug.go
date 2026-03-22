//go:build debug

package eval

import "fmt"

// MoveOrderStats tracks move ordering efficiency per search iteration.
type MoveOrderStats struct {
	failHighs      uint64
	failHighFirsts uint64
}

func (s *MoveOrderStats) recordFailHigh(first bool) {
	s.failHighs++
	if first {
		s.failHighFirsts++
	}
}

func (s *MoveOrderStats) reset() { *s = MoveOrderStats{} }

func (s *MoveOrderStats) String() string {
	if s.failHighs == 0 {
		return ""
	}
	rate := (100 * s.failHighFirsts) / s.failHighs
	return fmt.Sprintf("ordering: fh %d fhf %d%% (%d/%d)", s.failHighs, rate, s.failHighFirsts, s.failHighs)
}

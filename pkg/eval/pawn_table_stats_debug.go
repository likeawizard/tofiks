//go:build debug

package eval

import "fmt"

// PawnTableStats tracks pawn hash table probe statistics per search iteration.
type PawnTableStats struct {
	probes uint64
	hits   uint64
}

func (s *PawnTableStats) recordProbe() { s.probes++ }
func (s *PawnTableStats) recordHit()   { s.hits++ }
func (s *PawnTableStats) Reset()       { *s = PawnTableStats{} }

func (s *PawnTableStats) String() string {
	if s.probes == 0 {
		return ""
	}
	hitRate := (100 * s.hits) / s.probes
	return fmt.Sprintf("pawn_ht: probes %d hit %d%% miss %d",
		s.probes, hitRate, s.probes-s.hits,
	)
}

//go:build !debug

package eval

// PawnTableStats is a no-op in release builds.
type PawnTableStats struct{}

func (s *PawnTableStats) recordProbe()   {}
func (s *PawnTableStats) recordHit()     {}
func (s *PawnTableStats) reset()         {}
func (s *PawnTableStats) String() string { return "" }

//go:build !debug

package eval

// PawnTableStats is a no-op in release builds.
type PawnTableStats struct{}

func (s *PawnTableStats) recordProbe()   {}
func (s *PawnTableStats) recordHit()     {}
func (s *PawnTableStats) Reset()         {}
func (s *PawnTableStats) String() string { return "" }

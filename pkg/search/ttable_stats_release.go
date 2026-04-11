//go:build !debug

package search

// TTStats is a no-op in release builds.
type TTStats struct{}

func (s *TTStats) recordProbe()                            {}
func (s *TTStats) recordHit(_ int)                         {}
func (s *TTStats) recordCutoff(_ EntryType, _, _, _ int16) {}
func (s *TTStats) recordMoveHit()                          {}
func (s *TTStats) recordNewWrite()                         {}
func (s *TTStats) recordOverWrite()                        {}
func (s *TTStats) recordRejected()                         {}
func (s *TTStats) reset()                                  {}
func (s *TTStats) String() string                          { return "" }

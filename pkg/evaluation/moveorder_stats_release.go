//go:build !debug

package eval

// MoveOrderStats is a no-op in release builds.
type MoveOrderStats struct{}

func (s *MoveOrderStats) recordFailHigh(_ bool) {}
func (s *MoveOrderStats) reset()                {}
func (s *MoveOrderStats) String() string        { return "" }

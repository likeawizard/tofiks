//go:build !debug

package search

// PruneStats is a no-op in release builds.
type PruneStats struct{}

func (s *PruneStats) recordNMP(_ bool)   {}
func (s *PruneStats) recordRFP(_ bool)   {}
func (s *PruneStats) recordFP(_ bool)    {}
func (s *PruneStats) recordLMP(_ bool)   {}
func (s *PruneStats) recordSE(_, _ bool) {}
func (s *PruneStats) recordIIR()         {}
func (s *PruneStats) reset()             {}
func (s *PruneStats) String() string     { return "" }

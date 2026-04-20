//go:build debug

package search

import "fmt"

// PruneStats tracks per-technique pruning and reduction counters so we can
// tell which heuristics are actually producing cutoffs vs. just burning time.
// Attempts count how often the outer guard passed (i.e. the technique was
// eligible); fires count how often it actually triggered.
type PruneStats struct {
	nmpAttempts uint64
	nmpCutoffs  uint64
	rfpAttempts uint64
	rfpCutoffs  uint64
	fpAttempts  uint64
	fpPrunes    uint64
	lmpAttempts uint64
	lmpPrunes   uint64
	seAttempts  uint64
	seApplied   uint64
	seMultiCut  uint64
	iirFires    uint64
}

func (s *PruneStats) recordNMP(cutoff bool) {
	s.nmpAttempts++
	if cutoff {
		s.nmpCutoffs++
	}
}

func (s *PruneStats) recordRFP(cutoff bool) {
	s.rfpAttempts++
	if cutoff {
		s.rfpCutoffs++
	}
}

func (s *PruneStats) recordFP(pruned bool) {
	s.fpAttempts++
	if pruned {
		s.fpPrunes++
	}
}

func (s *PruneStats) recordLMP(pruned bool) {
	s.lmpAttempts++
	if pruned {
		s.lmpPrunes++
	}
}

func (s *PruneStats) recordSE(applied, multiCut bool) {
	s.seAttempts++
	if applied {
		s.seApplied++
	}
	if multiCut {
		s.seMultiCut++
	}
}

func (s *PruneStats) recordIIR() { s.iirFires++ }

func (s *PruneStats) reset() { *s = PruneStats{} }

func (s *PruneStats) String() string {
	total := s.nmpAttempts + s.rfpAttempts + s.fpAttempts + s.lmpAttempts + s.seAttempts + s.iirFires
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("prune: nmp %s rfp %s fp %s lmp %s se %d/%d mc %d iir %d",
		pruneRate(s.nmpCutoffs, s.nmpAttempts),
		pruneRate(s.rfpCutoffs, s.rfpAttempts),
		pruneRate(s.fpPrunes, s.fpAttempts),
		pruneRate(s.lmpPrunes, s.lmpAttempts),
		s.seApplied, s.seAttempts,
		s.seMultiCut,
		s.iirFires,
	)
}

func pruneRate(num, denom uint64) string {
	if denom == 0 {
		return "-"
	}
	return fmt.Sprintf("%d%%(%d/%d)", (100*num)/denom, num, denom)
}

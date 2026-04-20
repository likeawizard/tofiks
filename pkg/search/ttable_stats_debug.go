//go:build debug

package search

import "fmt"

// TTStats tracks transposition table probe and store statistics per search iteration.
type TTStats struct {
	probes             uint64
	hits               uint64
	scoreCuts          uint64
	scoreCutsExact     uint64
	scoreCutsLower     uint64
	scoreCutsUpper     uint64
	mateCutoffs        uint64
	cutoffOvershootSum uint64
	cutoffOvershootMax uint64
	moveHits           uint64
	newWrites          uint64
	overWrites         uint64
	rejected           uint64
	depthSum           uint64
}

func (s *TTStats) recordProbe()        { s.probes++ }
func (s *TTStats) recordHit(depth int) { s.hits++; s.depthSum += uint64(depth) }

// recordCutoff records a TT cutoff split by bound type, tracking overshoot
// (how far the stored eval is past the window edge) and mate-score detection.
// rawEval must be the pre-clamp stored score — GetScore returns a value
// clamped to alpha/beta on non-EXACT bounds, which would make overshoot always
// zero and hide mate-score cutoffs on LOWER/UPPER entries.
// Overshoot is always >= 0 and is 0 for EXACT bounds.
func (s *TTStats) recordCutoff(bound EntryType, rawEval, alpha, beta int16) {
	s.scoreCuts++
	overshoot := 0
	switch bound {
	case TT_EXACT:
		s.scoreCutsExact++
	case TT_LOWER:
		s.scoreCutsLower++
		overshoot = int(rawEval) - int(beta)
	case TT_UPPER:
		s.scoreCutsUpper++
		overshoot = int(alpha) - int(rawEval)
	}
	if overshoot < 0 {
		overshoot = 0
	}
	s.cutoffOvershootSum += uint64(overshoot)
	if uint64(overshoot) > s.cutoffOvershootMax {
		s.cutoffOvershootMax = uint64(overshoot)
	}
	if rawEval >= CheckmateThreshold || rawEval <= -CheckmateThreshold {
		s.mateCutoffs++
	}
}

func (s *TTStats) recordMoveHit()   { s.moveHits++ }
func (s *TTStats) recordNewWrite()  { s.newWrites++ }
func (s *TTStats) recordOverWrite() { s.overWrites++ }
func (s *TTStats) recordRejected()  { s.rejected++ }
func (s *TTStats) reset()           { *s = TTStats{} }

// String returns a human-readable summary of TT performance for the current iteration.
func (s *TTStats) String() string {
	if s.probes == 0 {
		return "tt: no probes"
	}
	hitRate := (100 * s.hits) / s.probes
	cutRate := uint64(0)
	avgDepth := uint64(0)
	if s.hits > 0 {
		cutRate = (100 * s.scoreCuts) / s.hits
		avgDepth = s.depthSum / s.hits
	}
	avgOvershoot := uint64(0)
	nonExactCuts := s.scoreCutsLower + s.scoreCutsUpper
	if nonExactCuts > 0 {
		avgOvershoot = s.cutoffOvershootSum / nonExactCuts
	}
	totalStores := s.newWrites + s.overWrites + s.rejected
	overwriteRate := uint64(0)
	rejectRate := uint64(0)
	if totalStores > 0 {
		overwriteRate = (100 * s.overWrites) / totalStores
		rejectRate = (100 * s.rejected) / totalStores
	}
	return fmt.Sprintf("tt: probes %d hit %d%% cut %d%% (E:%d L:%d U:%d mate:%d) avg_over %dcp max_over %dcp move %d avgdep %d | stores %d new %d over %d%% rej %d%%",
		s.probes, hitRate, cutRate,
		s.scoreCutsExact, s.scoreCutsLower, s.scoreCutsUpper, s.mateCutoffs,
		avgOvershoot, s.cutoffOvershootMax,
		s.moveHits, avgDepth,
		totalStores, s.newWrites, overwriteRate, rejectRate,
	)
}

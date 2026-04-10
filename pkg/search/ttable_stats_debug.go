//go:build debug

package search

import "fmt"

// TTStats tracks transposition table probe and store statistics per search iteration.
type TTStats struct {
	probes     uint64
	hits       uint64
	scoreCuts  uint64
	moveHits   uint64
	newWrites  uint64
	overWrites uint64
	rejected   uint64
	depthSum   uint64
}

func (s *TTStats) recordProbe()        { s.probes++ }
func (s *TTStats) recordHit(depth int) { s.hits++; s.depthSum += uint64(depth) }
func (s *TTStats) recordCutoff()       { s.scoreCuts++ }
func (s *TTStats) recordMoveHit()      { s.moveHits++ }
func (s *TTStats) recordNewWrite()     { s.newWrites++ }
func (s *TTStats) recordOverWrite()    { s.overWrites++ }
func (s *TTStats) recordRejected()     { s.rejected++ }
func (s *TTStats) reset()              { *s = TTStats{} }

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
	totalStores := s.newWrites + s.overWrites + s.rejected
	overwriteRate := uint64(0)
	rejectRate := uint64(0)
	if totalStores > 0 {
		overwriteRate = (100 * s.overWrites) / totalStores
		rejectRate = (100 * s.rejected) / totalStores
	}
	return fmt.Sprintf("tt: probes %d hit %d%% cut %d%% move %d avgdep %d | stores %d new %d over %d%% rej %d%%",
		s.probes, hitRate, cutRate, s.moveHits, avgDepth,
		totalStores, s.newWrites, overwriteRate, rejectRate,
	)
}

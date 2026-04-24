package eval

import "github.com/likeawizard/tofiks/pkg/board"

// PawnEntry stores cached pawn structure evaluation.
type PawnEntry struct {
	key   uint64
	score int16
}

// PawnTable is a fixed-size hash table for pawn structure evaluation.
// Size is always a power of 2 for fast index masking.
type PawnTable struct {
	entries []PawnEntry
	Stats   PawnTableStats
	mask    uint64
}

const pawnTableSizeMB = 2

func NewPawnTable() *PawnTable {
	count := uint64(pawnTableSizeMB*1024*1024) / 10 // ~10 bytes per entry
	// Round down to power of 2.
	size := uint64(1)
	for size*2 <= count {
		size *= 2
	}
	return &PawnTable{
		entries: make([]PawnEntry, size),
		mask:    size - 1,
	}
}

func (pt *PawnTable) Probe(key uint64) (int16, bool) {
	pt.Stats.recordProbe()
	e := &pt.entries[key&pt.mask]
	if e.key == key {
		pt.Stats.recordHit()
		return e.score, true
	}
	return 0, false
}

func (pt *PawnTable) Store(key uint64, score int16) {
	e := &pt.entries[key&pt.mask]
	e.key = key
	e.score = score
}

func (pt *PawnTable) Clear() {
	for i := range pt.entries {
		pt.entries[i] = PawnEntry{}
	}
}

// evaluatePawns computes the full pawn structure score for both sides.
func evaluatePawns(b *board.Board) int16 {
	var score int
	for color := board.White; color <= board.Black; color++ {
		side := 1
		if color == board.Black {
			side = -1
		}
		opp := color ^ 1
		ownPawns := b.Pieces[color][board.Pawns]
		oppPawns := b.Pieces[opp][board.Pawns]
		pieces := ownPawns
		for pieces > 0 {
			piece := pieces.PopLS1B()
			sq := board.Square(piece)
			s := int(sq)
			file := s % 8
			var value int

			if IsProtected(b, sq, color) {
				value = PawnProtected
			}
			if IsDoubled(b, sq, color) {
				value += PawnDoubled
			}

			isolated := IsIsolated(b, sq, color)
			if isolated {
				value += PawnIsolated
			}

			passed := IsPassed(b, sq, color)
			if passed {
				rank := 7 - s/8
				if color == board.Black {
					rank = s / 8
				}
				value += PassedPawnBonus[rank]

				// Connected passed pawns: passed pawn with a friendly passed pawn on an adjacent file.
				if board.AdjacentFiles[file] != 0 {
					adjPassers := board.AdjacentFiles[file] & ownPawns
					for adjPassers > 0 {
						adjSq := board.Square(adjPassers.PopLS1B())
						if IsPassed(b, adjSq, color) {
							value += PawnConnectedPasser
							break
						}
					}
				}
			}

			// Backward pawn: not isolated, not passed, no friendly pawn behind on adjacent files
			// that could advance to protect it, and the stop square is controlled by enemy pawns.
			if !isolated && !passed {
				stopSq := s - 8
				if color == board.Black {
					stopSq = s + 8
				}
				if stopSq >= 0 && stopSq < 64 {
					// Check if enemy pawns attack the stop square.
					stopAttacked := board.PawnAttacks[color][board.Square(stopSq)]&oppPawns != 0
					// Check if no friendly pawn behind on adjacent files can support.
					behindSupport := board.AdjacentFiles[file] & board.FrontSpan[color^1][sq] & ownPawns
					if stopAttacked && behindSupport == 0 {
						ownRank := 7 - s/8
						if color == board.Black {
							ownRank = s / 8
						}
						if ownRank <= 2 {
							value += PawnBackwardDeep
						} else {
							value += PawnBackwardMid
						}
						if board.FileMasks[file]&oppPawns == 0 {
							value += PawnBackwardOpen
						}
					}
				}
			}

			// Blocked pawn: directly blocked by an enemy pawn.
			stopSq := s - 8
			if color == board.Black {
				stopSq = s + 8
			}
			if stopSq >= 0 && stopSq < 64 && board.SquareBitboards[stopSq]&oppPawns != 0 {
				value += PawnBlocked
			}

			// Candidate passed pawn: not yet passed, but friendly supporters on adjacent files
			// outnumber or equal the enemy sentries.
			if !passed && !isolated {
				sentries := board.PassedPawns[color][sq] & oppPawns
				supporters := board.AdjacentFiles[file] & board.FrontSpan[color][sq] & ownPawns
				// Use PawnAttacks to also count pawns that can protect via capture.
				helpers := board.AdjacentFiles[file] & board.FrontSpan[color^1][sq] & ownPawns
				totalSupport := supporters.Count() + helpers.Count()
				if sentries != 0 && totalSupport >= sentries.Count() {
					value += PawnCandidate
				}
			}

			score += side * value
		}
	}
	return int16(score)
}

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
	var score int16
	for color := board.WHITE; color <= board.BLACK; color++ {
		side := int16(1)
		if color == board.BLACK {
			side = -1
		}
		opp := color ^ 1
		ownPawns := b.Pieces[color][board.PAWNS]
		oppPawns := b.Pieces[opp][board.PAWNS]
		pieces := ownPawns
		for pieces > 0 {
			piece := pieces.PopLS1B()
			sq := board.Square(piece)
			file := int(sq) % 8
			var value int16

			if IsProtected(b, sq, color) {
				value = int16(W_P_PROTECTED)
			}
			if IsDoubled(b, sq, color) {
				value += int16(W_P_DOUBLED)
			}

			isolated := IsIsolated(b, sq, color)
			if isolated {
				value += int16(W_P_ISOLATED)
			}

			passed := IsPassed(b, sq, color)
			if passed {
				rank := 7 - int(sq)/8
				if color == board.BLACK {
					rank = int(sq) / 8
				}
				value += int16(passedPawnBonus[rank])

				// Connected passed pawns: passed pawn with a friendly passed pawn on an adjacent file.
				if board.AdjacentFiles[file] != 0 {
					adjPassers := board.AdjacentFiles[file] & ownPawns
					for adjPassers > 0 {
						adjSq := board.Square(adjPassers.PopLS1B())
						if IsPassed(b, adjSq, color) {
							value += int16(W_P_CONNECTED_PASS)
							break
						}
					}
				}
			}

			// Backward pawn: not isolated, not passed, no friendly pawn behind on adjacent files
			// that could advance to protect it, and the stop square is controlled by enemy pawns.
			if !isolated && !passed {
				stopSq := int(sq) - 8
				if color == board.BLACK {
					stopSq = int(sq) + 8
				}
				if stopSq >= 0 && stopSq < 64 {
					// Check if enemy pawns attack the stop square.
					stopAttacked := board.PawnAttacks[color][board.Square(stopSq)]&oppPawns != 0
					// Check if no friendly pawn behind on adjacent files can support.
					behindSupport := board.AdjacentFiles[file] & board.FrontSpan[color^1][sq] & ownPawns
					if stopAttacked && behindSupport == 0 {
						value += int16(W_P_BACKWARD)
					}
				}
			}

			// Blocked pawn: directly blocked by an enemy pawn.
			stopSq := int(sq) - 8
			if color == board.BLACK {
				stopSq = int(sq) + 8
			}
			if stopSq >= 0 && stopSq < 64 && board.SquareBitboards[stopSq]&oppPawns != 0 {
				value += int16(W_P_BLOCKED)
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
					value += int16(W_P_CANDIDATE)
				}
			}

			score += side * value
		}
	}
	return score
}

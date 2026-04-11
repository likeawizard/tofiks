package texel

import (
	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/eval"
)

// TraceCoeff is a single non-zero coefficient in a sparse trace.
type TraceCoeff struct {
	Index uint16
	Value float32
}

// Trace holds the non-zero coefficients for a position in sparse format.
// This reduces memory from ~7.7 KB (dense [968]float64) to ~500 bytes per position.
type Trace []TraceCoeff

// denseTrace is used internally during trace computation, then compacted to sparse.
type denseTrace [NumParams]float64

// TraceEvaluate computes the coefficient trace for a position.
// The resulting trace, when dotted with the weight vector, reproduces the eval.
// Returns the phase (0-256) for external use.
func TraceEvaluate(b *board.Board) (Trace, int) {
	var t denseTrace
	phase := b.GetGamePhase()
	mgPhase := float64(256-phase) / 256.0
	egPhase := float64(phase) / 256.0

	oppKing := [2]board.BBoard{
		board.KingAttacks[board.Square(b.Pieces[board.Black][board.Kings].LS1B())],
		board.KingAttacks[board.Square(b.Pieces[board.White][board.Kings].LS1B())],
	}

	for color := board.White; color <= board.Black; color++ {
		sign := 1.0
		if color == board.Black {
			sign = -1.0
		}
		numPawns := b.Pieces[color][board.Pawns].Count()

		for pieceType := board.Pawns; pieceType <= board.Kings; pieceType++ {
			pieces := b.Pieces[color][pieceType]
			for pieces > 0 {
				piece := pieces.PopLS1B()
				sq := piece
				// For black, mirror the square for PST lookup (same as InitPSTs).
				pstSq := sq
				if color == board.Black {
					pstSq = (7-sq/8)*8 + sq%8
				}

				// PST coefficients (phase-interpolated).
				t[pstIndex(0, pieceType, pstSq)] += sign * mgPhase
				t[pstIndex(1, pieceType, pstSq)] += sign * egPhase

				// Piece weight coefficient (not phase-dependent).
				if pieceType < pieceWeightCount {
					t[pieceWeightStart+pieceType] += sign
				}

				// Piece-specific eval traces.
				switch pieceType {
				case board.Pawns:
					// Handled separately in tracePawns.
				case board.Knights:
					traceKnight(b, board.Square(sq), color, oppKing[color], sign, numPawns, &t)
				case board.Bishops:
					traceBishop(b, board.Square(sq), color, oppKing[color], sign, &t)
				case board.Rooks:
					traceRook(b, board.Square(sq), color, oppKing[color], sign, numPawns, &t)
				case board.Queens:
					traceQueen(b, board.Square(sq), color, oppKing[color], sign, &t)
				case board.Kings:
					traceKing(b, board.Square(sq), color, sign, phase, &t)
				}
			}
		}
	}

	// Pawn structure (not phase-dependent).
	tracePawns(b, &t)

	// Compact to sparse representation.
	var sparse Trace
	for i, v := range t {
		if v != 0 {
			sparse = append(sparse, TraceCoeff{Index: uint16(i), Value: float32(v)})
		}
	}
	return sparse, phase
}

// EvalFromTrace reconstructs the eval from a trace and weight vector.
func EvalFromTrace(t *Trace, w *[NumParams]float64) float64 {
	var sum float64
	for _, c := range *t {
		sum += float64(c.Value) * w[c.Index]
	}
	return sum
}

func traceKnight(b *board.Board, sq board.Square, side int, oppKing board.BBoard, sign float64, numPawns int, t *denseTrace) {
	moves := board.KnightAttacks[sq] & ^b.Occupancy[side]
	moveCount := float64(moves.Count())
	captureCount := float64((moves & b.Occupancy[side^1]).Count())
	threatCount := float64((moves & oppKing).Count())

	t[mobilityStart+3] += sign * moveCount // KnightMobility
	t[captureStart] += sign * captureCount // CaptureBonus
	t[threatStart+3] += sign * threatCount // KnightThreat

	// Kaufman knight-pawn slope: bonus = slope * (numPawns - 5).
	t[knightPawnSlopeStart] += sign * float64(numPawns-5)

	// Outpost.
	if board.Outposts[side][sq]&b.Pieces[side^1][board.Pawns] == 0 &&
		board.PawnAttacks[side^1][sq]&b.Pieces[side][board.Pawns] != 0 {
		outSq := int(sq)
		if side == board.Black {
			outSq = (7-outSq/8)*8 + outSq%8
		}
		t[outpostStart+outSq] += sign
	}
}

func traceBishop(b *board.Board, sq board.Square, side int, oppKing board.BBoard, sign float64, t *denseTrace) {
	moves := board.GetBishopAttacks(int(sq), b.Occupancy[board.Both])
	moveCount := float64(moves.Count())
	captureCount := float64((moves & b.Occupancy[side^1]).Count())
	threatCount := float64((moves & oppKing).Count())

	t[mobilityStart+2] += sign * moveCount // BishopMobility
	t[captureStart] += sign * captureCount // CaptureBonus
	t[threatStart+2] += sign * threatCount // BishopThreat

	// Outpost.
	if board.Outposts[side][sq]&b.Pieces[side^1][board.Pawns] == 0 &&
		board.PawnAttacks[side^1][sq]&b.Pieces[side][board.Pawns] != 0 {
		outSq := int(sq)
		if side == board.Black {
			outSq = (7-outSq/8)*8 + outSq%8
		}
		t[outpostStart+64+outSq] += sign
	}

	// Bishop pair.
	if b.Pieces[side][board.Bishops].Count() > 1 {
		t[bishopPairStart] += sign
	}
}

func traceRook(b *board.Board, sq board.Square, side int, oppKing board.BBoard, sign float64, numPawns int, t *denseTrace) {
	moves := board.GetRookAttacks(int(sq), b.Occupancy[board.Both])
	moveCount := float64(moves.Count())
	captureCount := float64((moves & b.Occupancy[side^1]).Count())
	threatCount := float64((moves & oppKing).Count())

	t[mobilityStart+1] += sign * moveCount // RookMobility
	t[captureStart] += sign * captureCount // CaptureBonus
	t[threatStart+1] += sign * threatCount // RookThreat

	// Kaufman rook-pawn slope: bonus = slope * (numPawns - 5).
	t[rookPawnSlopeStart] += sign * float64(numPawns-5)

	// Rook on open / semi-open file.
	file := board.FileMasks[sq%8]
	if file&b.Pieces[side][board.Pawns] == 0 {
		if file&b.Pieces[side^1][board.Pawns] == 0 {
			t[rookFileStart+0] += sign // open file
		} else {
			t[rookFileStart+1] += sign // semi-open file
		}
	}
}

func traceQueen(b *board.Board, sq board.Square, _ int, oppKing board.BBoard, sign float64, t *denseTrace) {
	moves := board.GetQueenAttacks(int(sq), b.Occupancy[board.Both])
	moveCount := float64(moves.Count())

	side := board.White
	if sign < 0 {
		side = board.Black
	}
	captures := float64((moves & b.Occupancy[side^1]).Count())
	threatCount := float64((moves & oppKing).Count())

	t[mobilityStart+0] += sign * moveCount // QueenMobility
	t[captureStart] += sign * captures     // CaptureBonus
	t[threatStart+0] += sign * threatCount // QueenThreat
}

func traceKing(b *board.Board, king board.Square, side int, sign float64, phase int, t *denseTrace) {
	mgPhase := float64(256-phase) / 256.0
	egPhase := float64(phase) / 256.0

	// King safety (MG): distCenter, pawnShield, friendlyNearKing.
	// Note: enemyNearKing was removed — it correlated with material count,
	// causing the tuner to flip its sign. Per-piece threats already capture
	// enemy pressure on the king zone more accurately.
	// Note: king mobility (MG and EG) was removed — a noisy proxy that
	// conflated castled kings (few moves, good) with mating-net kings (few
	// moves, losing), and active endgame kings with exposed center kings.
	distC := float64(eval.DistCenter(int(king)))
	pawnShield := float64((board.KingSafetyMask[side][king] & b.Pieces[side][board.Pawns]).Count())
	allFriendly := float64((board.KingSafetyMask[side][king] & b.Occupancy[side]).Count())
	friendlyNonPawn := allFriendly - pawnShield

	t[kingSafetyStart+0] += sign * distC * mgPhase
	t[kingSafetyStart+1] += sign * pawnShield * mgPhase
	t[kingSafetyStart+2] += sign * friendlyNonPawn * mgPhase

	// King activity (EG): distCenter, distSquares.
	distS := float64(eval.DistSquares(int(king), b.Pieces[side^1][board.Kings].LS1B()))
	t[kingActivityStart+0] += sign * distC * egPhase
	t[kingActivityStart+1] += sign * distS * egPhase
}

// tracePawns computes pawn structure coefficients for both sides.
func tracePawns(b *board.Board, t *denseTrace) {
	for color := board.White; color <= board.Black; color++ {
		sign := 1.0
		if color == board.Black {
			sign = -1.0
		}
		opp := color ^ 1
		ownPawns := b.Pieces[color][board.Pawns]
		oppPawns := b.Pieces[opp][board.Pawns]
		pieces := ownPawns

		for pieces > 0 {
			piece := pieces.PopLS1B()
			sq := board.Square(piece)
			file := int(sq) % 8

			if eval.IsProtected(b, sq, color) {
				t[pawnStructStart+0] += sign // protected
			}
			if eval.IsDoubled(b, sq, color) {
				t[pawnStructStart+1] += sign // doubled
			}

			isolated := eval.IsIsolated(b, sq, color)
			if isolated {
				t[pawnStructStart+2] += sign // isolated
			}

			passed := eval.IsPassed(b, sq, color)
			if passed {
				rank := 7 - int(sq)/8
				if color == board.Black {
					rank = int(sq) / 8
				}
				// Passed pawn bonus (ranks 1-6 map to indices 0-5).
				if rank >= 1 && rank <= 6 {
					t[passedPawnStart+rank-1] += sign
				}

				// Connected passed pawns.
				if board.AdjacentFiles[file] != 0 {
					adjPassers := board.AdjacentFiles[file] & ownPawns
					for adjPassers > 0 {
						adjSq := board.Square(adjPassers.PopLS1B())
						if eval.IsPassed(b, adjSq, color) {
							t[pawnStructStart+5] += sign // connectedPass
							break
						}
					}
				}
			}

			// Backward pawn.
			if !isolated && !passed {
				stopSq := int(sq) - 8
				if color == board.Black {
					stopSq = int(sq) + 8
				}
				if stopSq >= 0 && stopSq < 64 {
					stopAttacked := board.PawnAttacks[color][board.Square(stopSq)]&oppPawns != 0
					behindSupport := board.AdjacentFiles[file] & board.FrontSpan[color^1][sq] & ownPawns
					if stopAttacked && behindSupport == 0 {
						t[pawnStructStart+3] += sign // backward
					}
				}
			}

			// Blocked pawn.
			stopSq := int(sq) - 8
			if color == board.Black {
				stopSq = int(sq) + 8
			}
			if stopSq >= 0 && stopSq < 64 && board.SquareBitboards[stopSq]&oppPawns != 0 {
				t[pawnStructStart+4] += sign // blocked
			}

			// Candidate passed pawn.
			if !passed && !isolated {
				sentries := board.PassedPawns[color][sq] & oppPawns
				supporters := board.AdjacentFiles[file] & board.FrontSpan[color][sq] & ownPawns
				helpers := board.AdjacentFiles[file] & board.FrontSpan[color^1][sq] & ownPawns
				totalSupport := supporters.Count() + helpers.Count()
				if sentries != 0 && totalSupport >= sentries.Count() {
					t[pawnStructStart+6] += sign // candidate
				}
			}
		}
	}
}

package texel

import (
	"github.com/likeawizard/tofiks/pkg/board"
	eval "github.com/likeawizard/tofiks/pkg/evaluation"
)

// Parameter layout indices into the flat weight vector.
// PSTs: 2 stages * 6 pieces * 64 squares = 768
// Then scalar eval terms follow.
const (
	// PST block: [stage][piece][sq] linearized as stage*384 + piece*64 + sq.
	pstStart = 0
	pstCount = 2 * 6 * 64 // 768

	// Piece weights (5: pawn..queen, king excluded).
	pieceWeightStart = pstStart + pstCount
	pieceWeightCount = 5

	// Mobility weights: queen, rook, bishop, knight, king.
	mobilityStart = pieceWeightStart + pieceWeightCount
	mobilityCount = 5

	// Capture bonus.
	captureStart = mobilityStart + mobilityCount
	captureCount = 1

	// King threat weights: queen, rook, bishop, knight.
	threatStart = captureStart + captureCount
	threatCount = 4

	// Pawn structure: protected, doubled, isolated, backward, blocked, connectedPass, candidate.
	pawnStructStart = threatStart + threatCount
	pawnStructCount = 7

	// Passed pawn bonus by rank (ranks 1-6; ranks 0 and 7 are fixed at 0).
	passedPawnStart = pawnStructStart + pawnStructCount
	passedPawnCount = 6

	// Rook file bonuses: open, semi-open.
	rookFileStart = passedPawnStart + passedPawnCount
	rookFileCount = 2

	// Bishop pair.
	bishopPairStart = rookFileStart + rookFileCount
	bishopPairCount = 1

	// King safety MG: distCenter, pawnShield, friendlyNearKing, mobility.
	kingSafetyStart = bishopPairStart + bishopPairCount
	kingSafetyCount = 4

	// King activity EG: distCenter, distSquares, mobility.
	kingActivityStart = kingSafetyStart + kingSafetyCount
	kingActivityCount = 3

	// Outpost tables: knight[64] + bishop[64].
	outpostStart = kingActivityStart + kingActivityCount
	outpostCount = 128

	// Total parameter count.
	NumParams = outpostStart + outpostCount
)

// PST index helpers.
func pstIndex(stage, piece, sq int) int {
	return pstStart + stage*6*64 + piece*64 + sq
}

// recenterPSTs subtracts the mean from each piece's PST so the tables stay
// centered around zero and don't absorb material value. Piece weights are NOT
// modified — the optimizer adjusts them naturally. Pawn PST entries on ranks
// 1 and 8 are pinned to zero since pawns can never occupy those squares.
func recenterPSTs(w *[NumParams]float64) {
	for piece := 0; piece < pieceWeightCount; piece++ {
		for stage := 0; stage < 2; stage++ {
			// Pin impossible pawn squares (ranks 1 and 8) to zero.
			if piece == 0 { // PAWNS
				for sq := 0; sq < 8; sq++ {
					w[pstIndex(stage, piece, sq)] = 0    // rank 8
					w[pstIndex(stage, piece, 56+sq)] = 0 // rank 1
				}
			}

			var sum float64
			var count float64
			for sq := 0; sq < 64; sq++ {
				// Skip impossible pawn squares.
				if piece == 0 && (sq < 8 || sq >= 56) {
					continue
				}
				sum += w[pstIndex(stage, piece, sq)]
				count++
			}
			mean := sum / count
			for sq := 0; sq < 64; sq++ {
				if piece == 0 && (sq < 8 || sq >= 56) {
					continue
				}
				w[pstIndex(stage, piece, sq)] -= mean
			}
		}
	}
}

// InitialParams extracts the current eval weights into a flat vector.
func InitialParams() [NumParams]float64 {
	var p [NumParams]float64

	// PSTs (read from white's perspective).
	for stage := 0; stage < 2; stage++ {
		for piece := board.PAWNS; piece <= board.KINGS; piece++ {
			for sq := 0; sq < 64; sq++ {
				p[pstIndex(stage, piece, sq)] = float64(eval.PST[stage][board.WHITE][piece][sq])
			}
		}
	}

	// Piece weights.
	for i := 0; i < pieceWeightCount; i++ {
		p[pieceWeightStart+i] = float64(eval.PieceWeights[i])
	}

	// Mobility: queen=0, rook=1, bishop=2, knight=3, king=4.
	p[mobilityStart+0] = float64(eval.MOVE_QUEEN)
	p[mobilityStart+1] = float64(eval.MOVE_ROOK)
	p[mobilityStart+2] = float64(eval.MOVE_BISHOP)
	p[mobilityStart+3] = float64(eval.MOVE_KNIGHT)
	p[mobilityStart+4] = float64(eval.MOVE_KING)

	// Capture bonus.
	p[captureStart] = float64(eval.W_CAPTURE)

	// Threats: queen=0, rook=1, bishop=2, knight=3.
	p[threatStart+0] = float64(eval.QUEEN_THREAT)
	p[threatStart+1] = float64(eval.ROOK_THREAT)
	p[threatStart+2] = float64(eval.BISHOP_THREAT)
	p[threatStart+3] = float64(eval.KNIGHT_THREAT)

	// Pawn structure.
	p[pawnStructStart+0] = float64(eval.W_P_PROTECTED)
	p[pawnStructStart+1] = float64(eval.W_P_DOUBLED)
	p[pawnStructStart+2] = float64(eval.W_P_ISOLATED)
	p[pawnStructStart+3] = float64(eval.W_P_BACKWARD)
	p[pawnStructStart+4] = float64(eval.W_P_BLOCKED)
	p[pawnStructStart+5] = float64(eval.W_P_CONNECTED_PASS)
	p[pawnStructStart+6] = float64(eval.W_P_CANDIDATE)

	// Passed pawn bonus (ranks 1-6).
	for i := 0; i < passedPawnCount; i++ {
		p[passedPawnStart+i] = float64(eval.PassedPawnBonus[i+1])
	}

	// Rook file bonuses.
	p[rookFileStart+0] = float64(eval.W_ROOK_OPEN_FILE)
	p[rookFileStart+1] = float64(eval.W_ROOK_SEMI_OPEN_FILE)

	// Bishop pair.
	p[bishopPairStart] = float64(eval.W_BISHOP_PAIR)

	// King safety MG (enemyNearKing removed — correlated with material count).
	p[kingSafetyStart+0] = float64(eval.KS_DIST_CENTER)
	p[kingSafetyStart+1] = float64(eval.KS_PAWN_SHIELD)
	p[kingSafetyStart+2] = float64(eval.KS_FRIENDLY)
	p[kingSafetyStart+3] = float64(eval.MOVE_KING)

	// King activity EG.
	p[kingActivityStart+0] = float64(eval.KA_DIST_CENTER)
	p[kingActivityStart+1] = float64(eval.KA_DIST_SQUARES)
	p[kingActivityStart+2] = 5 // -MOVE_KING in the original code

	// Outposts.
	for sq := 0; sq < 64; sq++ {
		p[outpostStart+sq] = float64(eval.OutpostsScores[board.WHITE][board.KNIGHTS][sq])
		p[outpostStart+64+sq] = float64(eval.OutpostsScores[board.WHITE][board.BISHOPS][sq])
	}

	return p
}

// ApplyParams writes tuned weights back to the eval package globals.
func ApplyParams(p *[NumParams]float64) {
	for stage := 0; stage < 2; stage++ {
		for piece := board.PAWNS; piece <= board.KINGS; piece++ {
			for sq := 0; sq < 64; sq++ {
				eval.PST[stage][board.WHITE][piece][sq] = int(p[pstIndex(stage, piece, sq)])
			}
		}
	}
	eval.InitPSTs()

	for i := 0; i < pieceWeightCount; i++ {
		eval.PieceWeights[i] = int(p[pieceWeightStart+i])
	}

	eval.MOVE_QUEEN = int(p[mobilityStart+0])
	eval.MOVE_ROOK = int(p[mobilityStart+1])
	eval.MOVE_BISHOP = int(p[mobilityStart+2])
	eval.MOVE_KNIGHT = int(p[mobilityStart+3])
	eval.MOVE_KING = int(p[mobilityStart+4])

	eval.W_CAPTURE = int(p[captureStart])

	eval.QUEEN_THREAT = int(p[threatStart+0])
	eval.ROOK_THREAT = int(p[threatStart+1])
	eval.BISHOP_THREAT = int(p[threatStart+2])
	eval.KNIGHT_THREAT = int(p[threatStart+3])

	eval.W_P_PROTECTED = int(p[pawnStructStart+0])
	eval.W_P_DOUBLED = int(p[pawnStructStart+1])
	eval.W_P_ISOLATED = int(p[pawnStructStart+2])
	eval.W_P_BACKWARD = int(p[pawnStructStart+3])
	eval.W_P_BLOCKED = int(p[pawnStructStart+4])
	eval.W_P_CONNECTED_PASS = int(p[pawnStructStart+5])
	eval.W_P_CANDIDATE = int(p[pawnStructStart+6])

	for i := 0; i < passedPawnCount; i++ {
		eval.PassedPawnBonus[i+1] = int(p[passedPawnStart+i])
	}

	eval.W_ROOK_OPEN_FILE = int(p[rookFileStart+0])
	eval.W_ROOK_SEMI_OPEN_FILE = int(p[rookFileStart+1])

	eval.W_BISHOP_PAIR = int(p[bishopPairStart])

	eval.KS_DIST_CENTER = int(p[kingSafetyStart+0])
	eval.KS_PAWN_SHIELD = int(p[kingSafetyStart+1])
	eval.KS_FRIENDLY = int(p[kingSafetyStart+2])
	// kingSafetyStart+3 is king MG mobility, same as MOVE_KING (already set above).

	eval.KA_DIST_CENTER = int(p[kingActivityStart+0])
	eval.KA_DIST_SQUARES = int(p[kingActivityStart+1])
	// kingActivityStart+2 is king EG mobility — separate from MOVE_KING.
	for sq := 0; sq < 64; sq++ {
		eval.OutpostsScores[board.WHITE][board.KNIGHTS][sq] = int(p[outpostStart+sq])
		eval.OutpostsScores[board.WHITE][board.BISHOPS][sq] = int(p[outpostStart+64+sq])
	}
	// Rebuild black-side outpost tables.
	invert := func(sq int) int { return (7-sq/8)*8 + sq%8 }
	for piece := board.PAWNS; piece <= board.KINGS; piece++ {
		for sq := 0; sq < 64; sq++ {
			eval.OutpostsScores[board.BLACK][piece][sq] = eval.OutpostsScores[board.WHITE][piece][invert(sq)]
		}
	}
}

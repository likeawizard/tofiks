package texel

import (
	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/eval"
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
			if piece == 0 { // Pawns
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
		for piece := board.Pawns; piece <= board.Kings; piece++ {
			for sq := 0; sq < 64; sq++ {
				p[pstIndex(stage, piece, sq)] = float64(eval.PST[stage][board.White][piece][sq])
			}
		}
	}

	// Piece weights.
	for i := 0; i < pieceWeightCount; i++ {
		p[pieceWeightStart+i] = float64(eval.PieceWeights[i])
	}

	// Mobility: queen=0, rook=1, bishop=2, knight=3, king=4.
	p[mobilityStart+0] = float64(eval.QueenMobility)
	p[mobilityStart+1] = float64(eval.RookMobility)
	p[mobilityStart+2] = float64(eval.BishopMobility)
	p[mobilityStart+3] = float64(eval.KnightMobility)
	p[mobilityStart+4] = float64(eval.KingMobility)

	// Capture bonus.
	p[captureStart] = float64(eval.CaptureBonus)

	// Threats: queen=0, rook=1, bishop=2, knight=3.
	p[threatStart+0] = float64(eval.QueenThreat)
	p[threatStart+1] = float64(eval.RookThreat)
	p[threatStart+2] = float64(eval.BishopThreat)
	p[threatStart+3] = float64(eval.KnightThreat)

	// Pawn structure.
	p[pawnStructStart+0] = float64(eval.PawnProtected)
	p[pawnStructStart+1] = float64(eval.PawnDoubled)
	p[pawnStructStart+2] = float64(eval.PawnIsolated)
	p[pawnStructStart+3] = float64(eval.PawnBackward)
	p[pawnStructStart+4] = float64(eval.PawnBlocked)
	p[pawnStructStart+5] = float64(eval.PawnConnectedPasser)
	p[pawnStructStart+6] = float64(eval.PawnCandidate)

	// Passed pawn bonus (ranks 1-6).
	for i := 0; i < passedPawnCount; i++ {
		p[passedPawnStart+i] = float64(eval.PassedPawnBonus[i+1])
	}

	// Rook file bonuses.
	p[rookFileStart+0] = float64(eval.RookOpenFile)
	p[rookFileStart+1] = float64(eval.RookSemiOpenFile)

	// Bishop pair.
	p[bishopPairStart] = float64(eval.BishopPair)

	// King safety MG (enemyNearKing removed — correlated with material count).
	p[kingSafetyStart+0] = float64(eval.KingSafetyDistCenter)
	p[kingSafetyStart+1] = float64(eval.KingSafetyPawnShield)
	p[kingSafetyStart+2] = float64(eval.KingSafetyFriendly)
	p[kingSafetyStart+3] = float64(eval.KingMobility)

	// King activity EG.
	p[kingActivityStart+0] = float64(eval.KingActivityDistCenter)
	p[kingActivityStart+1] = float64(eval.KingActivityDistSquares)
	p[kingActivityStart+2] = 5 // -KingMobility in the original code

	// Outposts.
	for sq := 0; sq < 64; sq++ {
		p[outpostStart+sq] = float64(eval.OutpostsScores[board.White][board.Knights][sq])
		p[outpostStart+64+sq] = float64(eval.OutpostsScores[board.White][board.Bishops][sq])
	}

	return p
}

// ApplyParams writes tuned weights back to the eval package globals.
func ApplyParams(p *[NumParams]float64) {
	for stage := 0; stage < 2; stage++ {
		for piece := board.Pawns; piece <= board.Kings; piece++ {
			for sq := 0; sq < 64; sq++ {
				eval.PST[stage][board.White][piece][sq] = int(p[pstIndex(stage, piece, sq)])
			}
		}
	}
	eval.InitPSTs()

	for i := 0; i < pieceWeightCount; i++ {
		eval.PieceWeights[i] = int(p[pieceWeightStart+i])
	}

	eval.QueenMobility = int(p[mobilityStart+0])
	eval.RookMobility = int(p[mobilityStart+1])
	eval.BishopMobility = int(p[mobilityStart+2])
	eval.KnightMobility = int(p[mobilityStart+3])
	eval.KingMobility = int(p[mobilityStart+4])

	eval.CaptureBonus = int(p[captureStart])

	eval.QueenThreat = int(p[threatStart+0])
	eval.RookThreat = int(p[threatStart+1])
	eval.BishopThreat = int(p[threatStart+2])
	eval.KnightThreat = int(p[threatStart+3])

	eval.PawnProtected = int(p[pawnStructStart+0])
	eval.PawnDoubled = int(p[pawnStructStart+1])
	eval.PawnIsolated = int(p[pawnStructStart+2])
	eval.PawnBackward = int(p[pawnStructStart+3])
	eval.PawnBlocked = int(p[pawnStructStart+4])
	eval.PawnConnectedPasser = int(p[pawnStructStart+5])
	eval.PawnCandidate = int(p[pawnStructStart+6])

	for i := 0; i < passedPawnCount; i++ {
		eval.PassedPawnBonus[i+1] = int(p[passedPawnStart+i])
	}

	eval.RookOpenFile = int(p[rookFileStart+0])
	eval.RookSemiOpenFile = int(p[rookFileStart+1])

	eval.BishopPair = int(p[bishopPairStart])

	eval.KingSafetyDistCenter = int(p[kingSafetyStart+0])
	eval.KingSafetyPawnShield = int(p[kingSafetyStart+1])
	eval.KingSafetyFriendly = int(p[kingSafetyStart+2])
	// kingSafetyStart+3 is king MG mobility, same as KingMobility (already set above).

	eval.KingActivityDistCenter = int(p[kingActivityStart+0])
	eval.KingActivityDistSquares = int(p[kingActivityStart+1])
	// kingActivityStart+2 is king EG mobility — separate from KingMobility.
	for sq := 0; sq < 64; sq++ {
		eval.OutpostsScores[board.White][board.Knights][sq] = int(p[outpostStart+sq])
		eval.OutpostsScores[board.White][board.Bishops][sq] = int(p[outpostStart+64+sq])
	}
	// Rebuild black-side outpost tables.
	invert := func(sq int) int { return (7-sq/8)*8 + sq%8 }
	for piece := board.Pawns; piece <= board.Kings; piece++ {
		for sq := 0; sq < 64; sq++ {
			eval.OutpostsScores[board.Black][piece][sq] = eval.OutpostsScores[board.White][piece][invert(sq)]
		}
	}
}

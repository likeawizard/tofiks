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

	// Mobility weights: queen, rook, bishop, knight.
	// King mobility lives in the king safety (MG) and king activity (EG) blocks,
	// since it's phase-dependent — see kingSafetyStart+3 and kingActivityStart+2.
	mobilityStart = pieceWeightStart + pieceWeightCount
	mobilityCount = 4

	// King threat weights: queen, rook, bishop, knight.
	threatStart = mobilityStart + mobilityCount
	threatCount = 4

	// Pawn structure: protected, doubled, isolated, backwardDeep, backwardMid,
	// backwardOpen, blocked, connectedPass, candidate.
	pawnStructStart = threatStart + threatCount
	pawnStructCount = 9

	// Passed pawn bonus by rank (ranks 1-6; ranks 0 and 7 are fixed at 0).
	passedPawnStart = pawnStructStart + pawnStructCount
	passedPawnCount = 6

	// Rook file bonuses: open, semi-open.
	rookFileStart = passedPawnStart + passedPawnCount
	rookFileCount = 2

	// Bishop pair.
	bishopPairStart = rookFileStart + rookFileCount
	bishopPairCount = 1

	// King safety MG: distCenter, pawnShield, friendlyNearKing.
	// King mobility was removed — bad proxy conflating safety with mating nets.
	kingSafetyStart = bishopPairStart + bishopPairCount
	kingSafetyCount = 3

	// King activity EG: distCenter, distSquares.
	kingActivityStart = kingSafetyStart + kingSafetyCount
	kingActivityCount = 2

	// Outpost tables: knight[64] + bishop[64].
	outpostStart = kingActivityStart + kingActivityCount
	outpostCount = 128

	// Kaufman piece-value slopes: single linear rule per piece.
	// bonus = slope * (numPawns - 5). Pivot at 5 matches Kaufman's convention.
	knightPawnSlopeStart = outpostStart + outpostCount
	rookPawnSlopeStart   = knightPawnSlopeStart + 1

	// Passed-pawn king proximity: enemy-king-dist coefficient, friendly-king-dist
	// coefficient. Both multiplied by (rank × distance × egPhase) at eval time.
	passerKingProxStart = rookPawnSlopeStart + 1
	passerKingProxCount = 2

	// Tempo bonus for side to move.
	tempoStart = passerKingProxStart + passerKingProxCount
	tempoCount = 1

	// Victim-aware threats: pawn-on-minor, pawn-on-major, minor-on-rook,
	// minor-on-queen, rook-on-queen.
	threatsStart = tempoStart + tempoCount
	threatsCount = 5

	badBishopStart = threatsStart + threatsCount
	badBishopCount = 1

	// Pawn break: count of pawn pushes landing on an empty square that attacks
	// an enemy pawn (i.e. threatens to create a lever).
	pawnBreakStart = badBishopStart + badBishopCount
	pawnBreakCount = 1

	// Total parameter count.
	NumParams = pawnBreakStart + pawnBreakCount
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

	// Mobility: queen=0, rook=1, bishop=2, knight=3.
	// King mobility is phase-dependent and lives in the king-safety / king-activity blocks below.
	p[mobilityStart+0] = float64(eval.QueenMobility)
	p[mobilityStart+1] = float64(eval.RookMobility)
	p[mobilityStart+2] = float64(eval.BishopMobility)
	p[mobilityStart+3] = float64(eval.KnightMobility)

	// Threats: queen=0, rook=1, bishop=2, knight=3.
	p[threatStart+0] = float64(eval.QueenThreat)
	p[threatStart+1] = float64(eval.RookThreat)
	p[threatStart+2] = float64(eval.BishopThreat)
	p[threatStart+3] = float64(eval.KnightThreat)

	// Pawn structure.
	p[pawnStructStart+0] = float64(eval.PawnProtected)
	p[pawnStructStart+1] = float64(eval.PawnDoubled)
	p[pawnStructStart+2] = float64(eval.PawnIsolated)
	p[pawnStructStart+3] = float64(eval.PawnBackwardDeep)
	p[pawnStructStart+4] = float64(eval.PawnBackwardMid)
	p[pawnStructStart+5] = float64(eval.PawnBackwardOpen)
	p[pawnStructStart+6] = float64(eval.PawnBlocked)
	p[pawnStructStart+7] = float64(eval.PawnConnectedPasser)
	p[pawnStructStart+8] = float64(eval.PawnCandidate)

	// Passed pawn bonus (ranks 1-6).
	for i := 0; i < passedPawnCount; i++ {
		p[passedPawnStart+i] = float64(eval.PassedPawnBonus[i+1])
	}

	// Rook file bonuses.
	p[rookFileStart+0] = float64(eval.RookOpenFile)
	p[rookFileStart+1] = float64(eval.RookSemiOpenFile)

	// Bishop pair.
	p[bishopPairStart] = float64(eval.BishopPair)

	// King safety MG (enemyNearKing + king mobility removed).
	p[kingSafetyStart+0] = float64(eval.KingSafetyDistCenter)
	p[kingSafetyStart+1] = float64(eval.KingSafetyPawnShield)
	p[kingSafetyStart+2] = float64(eval.KingSafetyFriendly)

	// King activity EG.
	p[kingActivityStart+0] = float64(eval.KingActivityDistCenter)
	p[kingActivityStart+1] = float64(eval.KingActivityDistSquares)

	// Outposts.
	for sq := 0; sq < 64; sq++ {
		p[outpostStart+sq] = float64(eval.OutpostsScores[board.White][board.Knights][sq])
		p[outpostStart+64+sq] = float64(eval.OutpostsScores[board.White][board.Bishops][sq])
	}

	// Kaufman piece-value slopes.
	p[knightPawnSlopeStart] = float64(eval.KnightPawnSlope)
	p[rookPawnSlopeStart] = float64(eval.RookPawnSlope)

	// Passed-pawn king proximity (EG).
	p[passerKingProxStart+0] = float64(eval.PasserEnemyKingDist)
	p[passerKingProxStart+1] = float64(eval.PasserFriendlyKingDist)

	// Tempo.
	p[tempoStart] = float64(eval.Tempo)

	// Victim-aware threats.
	p[threatsStart+0] = float64(eval.ThreatPawnOnMinor)
	p[threatsStart+1] = float64(eval.ThreatPawnOnMajor)
	p[threatsStart+2] = float64(eval.ThreatMinorOnRook)
	p[threatsStart+3] = float64(eval.ThreatMinorOnQueen)
	p[threatsStart+4] = float64(eval.ThreatRookOnQueen)

	// Bad bishop.
	p[badBishopStart] = float64(eval.BadBishop)

	// Pawn break.
	p[pawnBreakStart] = float64(eval.PawnBreak)

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

	eval.QueenThreat = int(p[threatStart+0])
	eval.RookThreat = int(p[threatStart+1])
	eval.BishopThreat = int(p[threatStart+2])
	eval.KnightThreat = int(p[threatStart+3])

	eval.PawnProtected = int(p[pawnStructStart+0])
	eval.PawnDoubled = int(p[pawnStructStart+1])
	eval.PawnIsolated = int(p[pawnStructStart+2])
	eval.PawnBackwardDeep = int(p[pawnStructStart+3])
	eval.PawnBackwardMid = int(p[pawnStructStart+4])
	eval.PawnBackwardOpen = int(p[pawnStructStart+5])
	eval.PawnBlocked = int(p[pawnStructStart+6])
	eval.PawnConnectedPasser = int(p[pawnStructStart+7])
	eval.PawnCandidate = int(p[pawnStructStart+8])

	for i := 0; i < passedPawnCount; i++ {
		eval.PassedPawnBonus[i+1] = int(p[passedPawnStart+i])
	}

	eval.RookOpenFile = int(p[rookFileStart+0])
	eval.RookSemiOpenFile = int(p[rookFileStart+1])

	eval.BishopPair = int(p[bishopPairStart])

	eval.KingSafetyDistCenter = int(p[kingSafetyStart+0])
	eval.KingSafetyPawnShield = int(p[kingSafetyStart+1])
	eval.KingSafetyFriendly = int(p[kingSafetyStart+2])

	eval.KingActivityDistCenter = int(p[kingActivityStart+0])
	eval.KingActivityDistSquares = int(p[kingActivityStart+1])
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

	// Kaufman piece-value slopes.
	eval.KnightPawnSlope = int(p[knightPawnSlopeStart])
	eval.RookPawnSlope = int(p[rookPawnSlopeStart])

	// Passed-pawn king proximity.
	eval.PasserEnemyKingDist = int(p[passerKingProxStart+0])
	eval.PasserFriendlyKingDist = int(p[passerKingProxStart+1])

	// Tempo.
	eval.Tempo = int(p[tempoStart])

	// Victim-aware threats.
	eval.ThreatPawnOnMinor = int(p[threatsStart+0])
	eval.ThreatPawnOnMajor = int(p[threatsStart+1])
	eval.ThreatMinorOnRook = int(p[threatsStart+2])
	eval.ThreatMinorOnQueen = int(p[threatsStart+3])
	eval.ThreatRookOnQueen = int(p[threatsStart+4])

	eval.BadBishop = int(p[badBishopStart])

	eval.PawnBreak = int(p[pawnBreakStart])
}

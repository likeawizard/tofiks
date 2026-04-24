package eval

import (
	"github.com/likeawizard/tofiks/pkg/board"
)

// Eval holds the stateful side of evaluation: the pawn structure cache and
// any per-game scratch space. Created once per Engine and reused across calls.
type Eval struct {
	PawnTable *PawnTable
}

// New constructs an Eval with a fresh pawn table.
func New() *Eval {
	return &Eval{
		PawnTable: NewPawnTable(),
	}
}

var (
	// PieceWeights represents the base value of each piece.
	PieceWeights = [6]int{134, 393, 389, 628, 1261, 10000}

	// KnightPawnSlope and RookPawnSlope are L. Kaufman's piece-value adjustments
	// Tuner persistently finds both opposite-sign from Kaufman's theory — likely
	// a structural interaction with other eval terms rather than a data artifact.
	KnightPawnSlope = -1
	RookPawnSlope   = 2

	dist = [64]int{
		4, 3, 3, 3, 3, 3, 3, 4,
		3, 3, 2, 2, 2, 2, 3, 3,
		3, 2, 2, 1, 1, 2, 2, 3,
		3, 2, 1, 0, 0, 1, 2, 3,
		3, 2, 1, 0, 0, 1, 2, 3,
		3, 2, 2, 1, 1, 2, 2, 3,
		3, 3, 2, 2, 2, 2, 3, 3,
		4, 3, 3, 3, 3, 3, 3, 4,
	}
)

var (
	QueenMobility  = 2
	RookMobility   = 3
	BishopMobility = 7
	KnightMobility = 0

	QueenThreat  = 15
	RookThreat   = 4
	BishopThreat = 4
	KnightThreat = 1

	PawnProtected       = 18
	PawnDoubled         = -16
	PawnIsolated        = -10
	PawnBackwardDeep    = -15
	PawnBackwardMid     = -12
	PawnBackwardOpen    = 8
	PawnBlocked         = -5
	PawnConnectedPasser = 8
	PawnCandidate       = 8
	PawnBreak           = 1

	RookOpenFile     = 27
	RookSemiOpenFile = 28

	BishopPair = 21

	BadBishop = -10

	KingSafetyDistCenter = -13
	KingSafetyPawnShield = 29
	KingSafetyFriendly   = 6

	KingActivityDistCenter  = -24
	KingActivityDistSquares = -1

	PassedPawnBonus = [8]int{0, -27, -37, -22, 14, 60, 205, 0}

	// PasserEnemyKingDist / PasserFriendlyKingDist scale rank × Manhattan
	// distance to each king and are applied EG-only. Enemy-king-far is good,
	// friendly-king-close is good — signs set accordingly.
	PasserEnemyKingDist    = 5
	PasserFriendlyKingDist = -2

	// Tempo is a flat bonus for the side to move.
	Tempo = 29

	// Victim-aware threats.
	ThreatPawnOnMinor  = 127
	ThreatPawnOnMajor  = 96
	ThreatMinorOnRook  = 76
	ThreatMinorOnQueen = 13
	ThreatRookOnQueen  = 57
)

// Piece protected a pawn.
func IsProtected(b *board.Board, sq board.Square, side int) bool {
	return board.PawnAttacks[side^1][sq]&b.Pieces[side][board.Pawns] != 0
}

func IsDoubled(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.Pawns]&board.DoubledPawns[sq] != 0
}

// Has no friendly pawns on neighboring files.
func IsIsolated(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.Pawns]&board.IsolatedPawns[sq] == 0
}

// Has no opponent opposing pawns in front (same or neighbor files).
func IsPassed(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side^1][board.Pawns]&board.PassedPawns[side][sq] == 0
}

// GetEvaluation returns the static evaluation of the position from white's
// perspective in centipawns. Caller is responsible for any side-relative
// negation. Uses the pawn cache when available.
func (e *Eval) GetEvaluation(b *board.Board) int {
	b.Phase = b.GetGamePhase()

	// Pawn structure evaluation via hash table.
	var pawnScore int16
	if cached, ok := e.PawnTable.Probe(b.PawnHash); ok {
		pawnScore = cached
	} else {
		pawnScore = evaluatePawns(b)
		e.PawnTable.Store(b.PawnHash, pawnScore)
	}

	var (
		eval     = int(pawnScore)
		side     = -1
		numPawns int
		pieces   board.BBoard
		piece    int
		oppKing  board.BBoard
	)

	for color := board.White; color <= board.Black; color++ {
		side *= -1
		friendlyKingSq := b.Pieces[color][board.Kings].LS1B()
		enemyKingSq := b.Pieces[color^1][board.Kings].LS1B()
		oppKing = board.KingAttacks[enemyKingSq]
		numPawns = b.Pieces[color][board.Pawns].Count()

		for pieceType := board.Pawns; pieceType <= board.Kings; pieceType++ {
			pieces = b.Pieces[color][pieceType]
			for pieces > 0 {
				piece = pieces.PopLS1B()
				var pieceEval int
				switch pieceType {
				case board.Pawns:
					// Structure already evaluated via pawn table.
				case board.Bishops:
					pieceEval = bishopEval(b, piece, color, oppKing)
				case board.Knights:
					pieceEval = knightEval(b, piece, color, oppKing, numPawns)
				case board.Rooks:
					pieceEval = rookEval(b, piece, color, oppKing, numPawns)
				case board.Queens:
					pieceEval = queenEval(b, piece, oppKing)
				case board.Kings:
					pieceEval = kingEval(b, piece, color, oppKing)
				}
				eval += side * (PieceWeights[pieceType] +
					pieceEval +
					(PST[0][color][pieceType][piece]*(256-b.Phase)+
						PST[1][color][pieceType][piece]*b.Phase)/256)
			}
		}

		// Pawn threats on enemy non-pawn pieces. Kept outside the pawn-hash
		// cache because the score depends on enemy piece positions.
		ownPawns := b.Pieces[color][board.Pawns]
		var pawnAttackBB board.BBoard
		if color == board.White {
			pawnAttackBB = ((ownPawns & ^board.FileMasks[7]) >> 7) | ((ownPawns & ^board.FileMasks[0]) >> 9)
		} else {
			pawnAttackBB = ((ownPawns & ^board.FileMasks[0]) << 7) | ((ownPawns & ^board.FileMasks[7]) << 9)
		}
		enemyMinors := b.Pieces[color^1][board.Knights] | b.Pieces[color^1][board.Bishops]
		enemyMajors := b.Pieces[color^1][board.Rooks] | b.Pieces[color^1][board.Queens]
		eval += side * ((pawnAttackBB&enemyMinors).Count()*ThreatPawnOnMinor +
			(pawnAttackBB&enemyMajors).Count()*ThreatPawnOnMajor)

		// Passed-pawn king proximity (EG-only). Kept outside the pawn-hash
		// cache because the score depends on king squares, not pawn structure.
		for pawns := b.Pieces[color][board.Pawns]; pawns > 0; {
			sq := pawns.PopLS1B()
			if !IsPassed(b, board.Square(sq), color) {
				continue
			}
			rank := 7 - sq/8
			if color == board.Black {
				rank = sq / 8
			}
			kingProx := rank * (DistSquares(enemyKingSq, sq)*PasserEnemyKingDist +
				DistSquares(friendlyKingSq, sq)*PasserFriendlyKingDist)
			eval += side * kingProx * b.Phase / 256
		}
	}

	eval += evaluatePawnBreaks(b)

	// Tempo bonus for side to move (from White's perspective).
	if b.Side == board.White {
		eval += Tempo
	} else {
		eval -= Tempo
	}

	return eval
}

// evaluatePawnBreaks counts pawn pushes whose destination is empty and attacks an enemy pawn.
func evaluatePawnBreaks(b *board.Board) int {
	empty := ^b.Occupancy[board.Both]

	wPawns := b.Pieces[board.White][board.Pawns]
	bPawns := b.Pieces[board.Black][board.Pawns]

	wBreakTarget := ((bPawns & ^board.FileMasks[0]) << 7) | ((bPawns & ^board.FileMasks[7]) << 9)
	bBreakTarget := ((wPawns & ^board.FileMasks[7]) >> 7) | ((wPawns & ^board.FileMasks[0]) >> 9)

	wSingle := (wPawns >> 8) & empty
	wDouble := (((wPawns & board.Rank2) >> 8) & empty) >> 8 & empty
	bSingle := (bPawns << 8) & empty
	bDouble := (((bPawns & board.Rank7) << 8) & empty) << 8 & empty

	wBreaks := ((wSingle | wDouble) & wBreakTarget).Count()
	bBreaks := ((bSingle | bDouble) & bBreakTarget).Count()

	return (wBreaks - bBreaks) * PawnBreak
}

func DistCenter(sq int) int {
	return dist[sq]
}

func DistSquares(us, them int) int {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	return abs(us/8-them/8) + abs(us%8-them%8)
}

// King safety score as a measure of distance from the board center and the number of adjacent enemy pieces and friendly pieces
// TODO: naive initial approach
// use board direction to value own pieces: i.e. a King in front of 3 pawns is not the same as a king behind 3 pawns
// consider using opponent piece attacks around the king instead of actual pieces. use piece weights for opponent threat levels: a queen near our king should be a larger concern than a bishop.
func getKingSafety(b *board.Board, king int, side int) (kingSafety int) {
	kingSafety += KingSafetyDistCenter * DistCenter(king)
	pawnShield := (board.KingSafetyMask[side][king] & b.Pieces[side][board.Pawns]).Count()
	kingSafety += KingSafetyPawnShield*pawnShield + KingSafetyFriendly*((board.KingSafetyMask[side][king]&b.Occupancy[side]).Count()-pawnShield)
	return kingSafety
}

func getKingActivity(b *board.Board, king int, side int) (kingActivity int) {
	kingActivity = KingActivityDistCenter*DistCenter(king) + KingActivityDistSquares*DistSquares(king, b.Pieces[side^1][board.Kings].LS1B())
	return kingActivity
}

func knightEval(b *board.Board, sq int, side int, oppKing board.BBoard, numPawns int) int {
	var eval int
	moves := board.KnightAttacks[sq] & ^b.Occupancy[side]
	if board.Outposts[side][sq]&b.Pieces[side^1][board.Pawns] == 0 &&
		board.PawnAttacks[side^1][sq]&b.Pieces[side][board.Pawns] != 0 {
		eval = OutpostsScores[side][board.Knights][sq]
	}
	return eval + moves.Count()*KnightMobility +
		(moves&oppKing).Count()*KnightThreat +
		(moves&b.Pieces[side^1][board.Rooks]).Count()*ThreatMinorOnRook +
		(moves&b.Pieces[side^1][board.Queens]).Count()*ThreatMinorOnQueen +
		KnightPawnSlope*(numPawns-5)
}

func bishopEval(b *board.Board, sq int, side int, oppKing board.BBoard) int {
	var eval int
	moves := board.GetBishopAttacks(sq, b.Occupancy[board.Both])

	if board.Outposts[side][sq]&b.Pieces[side^1][board.Pawns] == 0 &&
		board.PawnAttacks[side^1][sq]&b.Pieces[side][board.Pawns] != 0 {
		eval = OutpostsScores[side][board.Bishops][sq]
	}
	if b.Pieces[side][board.Bishops].Count() > 1 {
		eval += BishopPair
	}
	eval += BadBishop * (b.Pieces[side][board.Pawns] & board.SquareColorMask[sq]).Count()
	return eval + moves.Count()*BishopMobility +
		(moves&oppKing).Count()*BishopThreat +
		(moves&b.Pieces[side^1][board.Rooks]).Count()*ThreatMinorOnRook +
		(moves&b.Pieces[side^1][board.Queens]).Count()*ThreatMinorOnQueen
}

// Evaluation for rooks - mobility, captures, king threats, and (semi)open files.
func rookEval(b *board.Board, sq int, side int, oppKing board.BBoard, numPawns int) int {
	moves := board.GetRookAttacks(sq, b.Occupancy[board.Both])
	eval := moves.Count()*RookMobility +
		(moves&oppKing).Count()*RookThreat +
		(moves&b.Pieces[side^1][board.Queens]).Count()*ThreatRookOnQueen +
		RookPawnSlope*(numPawns-5)

	file := board.FileMasks[sq%8]
	if file&b.Pieces[side][board.Pawns] == 0 {
		if file&b.Pieces[side^1][board.Pawns] == 0 {
			eval += RookOpenFile
		} else {
			eval += RookSemiOpenFile
		}
	}

	return eval
}

func queenEval(b *board.Board, sq int, oppKing board.BBoard) int {
	moves := board.GetQueenAttacks(sq, b.Occupancy[board.Both])
	return moves.Count()*QueenMobility +
		(moves&oppKing).Count()*QueenThreat
}

func kingEval(b *board.Board, king int, side int, _ board.BBoard) int {
	return (getKingSafety(b, king, side)*(256-b.Phase) + getKingActivity(b, king, side)*b.Phase) / 256
}

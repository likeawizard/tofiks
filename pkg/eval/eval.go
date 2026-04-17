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
	PieceWeights = [6]int{124, 351, 383, 626, 1268, 10000}

	// KnightPawnSlope and RookPawnSlope are L. Kaufman's piece-value adjustments
	// expressed as a single linear rule: each own pawn above 5 adjusts the piece
	// value by `slope`, and each pawn below 5 by `-slope`. The pivot at 5 pawns
	// is the Kaufman convention ("no adjustment" point).
	//
	//   knightBonus = KnightPawnSlope * (numPawns - 5)
	//   rookBonus   = RookPawnSlope   * (numPawns - 5)
	//
	// Defaults derive from Kaufman's original ratios (knight_value / 16 and
	// -rook_value / 8). Both are applied as flat per-piece value adjustments.
	KnightPawnSlope = 4  // knights gain value in pawn-heavy positions
	RookPawnSlope   = -5 // rooks lose value in pawn-heavy positions

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
	BishopMobility = 9
	KnightMobility = -1
	CaptureBonus   = 9

	QueenThreat  = 17
	RookThreat   = 4
	BishopThreat = 4
	KnightThreat = 1

	PawnProtected       = 17
	PawnDoubled         = -16
	PawnIsolated        = -10
	PawnBackward        = -7
	PawnBlocked         = -5
	PawnConnectedPasser = 9
	PawnCandidate       = 8

	RookOpenFile     = 27
	RookSemiOpenFile = 21

	BishopPair = 22

	KingSafetyDistCenter = 6
	KingSafetyPawnShield = 35
	KingSafetyFriendly   = 3

	KingActivityDistCenter  = -23
	KingActivityDistSquares = -1

	PassedPawnBonus = [8]int{0, -22, -32, -14, 20, 68, 206, 0}

	// PasserEnemyKingDist / PasserFriendlyKingDist scale rank × Manhattan
	// distance to each king and are applied EG-only. Enemy-king-far is good,
	// friendly-king-close is good — signs set accordingly.
	PasserEnemyKingDist    = 5
	PasserFriendlyKingDist = -2

	// Tempo is a flat bonus for the side to move.
	Tempo = 15
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
					pieceEval = queenEval(b, piece, color, oppKing)
				case board.Kings:
					pieceEval = kingEval(b, piece, color, oppKing)
				}
				eval += side * (PieceWeights[pieceType] +
					pieceEval +
					(PST[0][color][pieceType][piece]*(256-b.Phase)+
						PST[1][color][pieceType][piece]*b.Phase)/256)
			}
		}

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

	// Tempo bonus for side to move (from White's perspective).
	if b.Side == board.White {
		eval += Tempo
	} else {
		eval -= Tempo
	}

	return eval
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
		(moves&b.Occupancy[side^1]).Count()*CaptureBonus +
		(moves&oppKing).Count()*KnightThreat +
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
	return eval + moves.Count()*BishopMobility +
		(moves&b.Occupancy[side^1]).Count()*CaptureBonus +
		(moves&oppKing).Count()*BishopThreat
}

// Evaluation for rooks - mobility, captures, king threats, and (semi)open files.
func rookEval(b *board.Board, sq int, side int, oppKing board.BBoard, numPawns int) int {
	moves := board.GetRookAttacks(sq, b.Occupancy[board.Both])
	eval := moves.Count()*RookMobility +
		(moves&b.Occupancy[side^1]).Count()*CaptureBonus +
		(moves&oppKing).Count()*RookThreat +
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

func queenEval(b *board.Board, sq int, side int, oppKing board.BBoard) int {
	moves := board.GetQueenAttacks(sq, b.Occupancy[board.Both])
	return moves.Count()*QueenMobility +
		(moves&b.Occupancy[side^1]).Count()*CaptureBonus +
		(moves&oppKing).Count()*QueenThreat
}

func kingEval(b *board.Board, king int, side int, _ board.BBoard) int {
	return (getKingSafety(b, king, side)*(256-b.Phase) + getKingActivity(b, king, side)*b.Phase) / 256
}

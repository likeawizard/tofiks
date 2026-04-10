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
	PieceWeights = [6]int{111, 263, 288, 498, 991, 10000}

	// Based on L. Kaufman - rook and knight values are adjusted by the number of pawns on the board.
	PiecePawnBonus = [6][9]int{
		{},
		{},
		{-25, -19, -13, -6, 0, 6, 13, 19, 25},
		{50, 37, 25, 12, 0, -12, -25, -37, -50},
		{},
		{},
	}

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
	BishopMobility = 6
	KnightMobility = -2
	KingMobility   = -5
	CaptureBonus   = 8

	QueenThreat  = 16
	RookThreat   = 1
	BishopThreat = 7
	KnightThreat = 3

	PawnProtected       = 10
	PawnDoubled         = -12
	PawnIsolated        = -5
	PawnBackward        = -1
	PawnBlocked         = -5
	PawnConnectedPasser = 11
	PawnCandidate       = 8

	RookOpenFile     = 23
	RookSemiOpenFile = 17

	BishopPair = 15

	KingSafetyDistCenter = 24
	KingSafetyPawnShield = 31
	KingSafetyFriendly   = 8

	KingActivityDistCenter  = -24
	KingActivityDistSquares = -1

	PassedPawnBonus = [8]int{0, 0, 0, 20, 44, 67, 156, 0}
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
		oppKing = board.KingAttacks[b.Pieces[color^1][board.Kings].LS1B()]
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
					pieceEval = knightEval(b, piece, color, oppKing)
				case board.Rooks:
					pieceEval = rookEval(b, piece, color, oppKing)
				case board.Queens:
					pieceEval = queenEval(b, piece, color, oppKing)
				case board.Kings:
					pieceEval = kingEval(b, piece, color, oppKing)
				}
				eval += side * (PieceWeights[pieceType] +
					pieceEval +
					(PST[0][color][pieceType][piece]*(256-b.Phase)+
						(PST[1][color][pieceType][piece]+PiecePawnBonus[pieceType][numPawns])*b.Phase)/256)
			}
		}
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

func knightEval(b *board.Board, sq int, side int, oppKing board.BBoard) int {
	var eval int
	moves := board.KnightAttacks[sq] & ^b.Occupancy[side]
	if board.Outposts[side][sq]&b.Pieces[side^1][board.Pawns] == 0 &&
		board.PawnAttacks[side^1][sq]&b.Pieces[side][board.Pawns] != 0 {
		eval = OutpostsScores[side][board.Knights][sq]
	}
	return eval + moves.Count()*KnightMobility +
		(moves&b.Occupancy[side^1]).Count()*CaptureBonus +
		(moves&oppKing).Count()*KnightThreat
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
func rookEval(b *board.Board, sq int, side int, oppKing board.BBoard) int {
	moves := board.GetRookAttacks(sq, b.Occupancy[board.Both])
	eval := moves.Count()*RookMobility +
		(moves&b.Occupancy[side^1]).Count()*CaptureBonus +
		(moves&oppKing).Count()*RookThreat

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
	moves := board.KingAttacks[king] & ^b.Occupancy[side]
	return ((getKingSafety(b, king, side)+moves.Count()*KingMobility)*(256-b.Phase) + (getKingActivity(b, king, side)-moves.Count()*KingMobility)*b.Phase) / 256
}

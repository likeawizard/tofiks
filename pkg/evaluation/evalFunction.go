package eval

import (
	"github.com/likeawizard/tofiks/pkg/board"
)

var (
	// PieceWeights represents the base value of each piece.
	PieceWeights = [6]int{100, 325, 325, 500, 975, 10000}

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

const (
	// Mobility related weights.
	MOVE_QUEEN      = 1
	MOVE_ROOK       = 2
	MOVE_BISHOP     = 3
	MOVE_KNIGHT     = 5
	MOVE_KING       = -5
	W_CAPTURE   int = 4

	// King threat weights. How much a piece contributes to the king safety evaluation.
	QUEEN_THREAT  = 10
	ROOK_THREAT   = 6
	BISHOP_THREAT = 4
	KNIGHT_THREAT = 6

	// Pawn structure weights.
	W_P_PROTECTED int = 15
	W_P_DOUBLED   int = -15
	W_P_ISOLATED  int = -20

	// Rook file bonuses.
	W_ROOK_OPEN_FILE      = 20
	W_ROOK_SEMI_OPEN_FILE = 10
)

// Passed pawn bonus indexed by rank from the advancing side's perspective (0=back rank, 7=promotion rank).
var passedPawnBonus = [8]int{0, 5, 10, 20, 40, 70, 120, 0}

// Piece protected a pawn.
func IsProtected(b *board.Board, sq board.Square, side int) bool {
	return board.PawnAttacks[side^1][sq]^b.Pieces[side][board.PAWNS] != 0
}

func IsDoubled(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.PAWNS]&board.DoubledPawns[sq] != 0
}

// Has no friendly pawns on neighboring files.
func IsIsolated(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.PAWNS]&board.IsolatedPawns[sq] == 0
}

// Has no opponent opposing pawns in front (same or neighbor files).
func IsPassed(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.PAWNS]&board.PassedPawns[side][sq] == 0
}

func (e *Engine) GetEvaluation(b *board.Board) int {
	e.Stats.evals++
	e.Board.Phase = e.Board.GetGamePhase()

	var (
		eval     int
		side     = -1
		numPawns int
		pieces   board.BBoard
		piece    int
		oppKing  board.BBoard
	)

	for color := board.WHITE; color <= board.BLACK; color++ {
		side *= -1
		oppKing = board.KingAttacks[board.Square(b.Pieces[color^1][board.KINGS].LS1B())]
		numPawns = b.Pieces[color][board.PAWNS].Count()

		for pieceType := board.PAWNS; pieceType <= board.KINGS; pieceType++ {
			pieces = b.Pieces[color][pieceType]
			for pieces > 0 {
				piece = pieces.PopLS1B()
				sq := board.Square(piece)
				var pieceEval int
				switch pieceType {
				case board.PAWNS:
					pieceEval = pawnEval(b, sq, color, oppKing)
				case board.BISHOPS:
					pieceEval = bishopEval(b, sq, color, oppKing)
				case board.KNIGHTS:
					pieceEval = knightEval(b, sq, color, oppKing)
				case board.ROOKS:
					pieceEval = rookEval(b, sq, color, oppKing)
				case board.QUEENS:
					pieceEval = queenEval(b, sq, color, oppKing)
				case board.KINGS:
					pieceEval = kingEval(b, sq, color, oppKing)
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

func distCenter(sq board.Square) int {
	return dist[sq]
}

func distSquares(us, them board.Square) int {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	u, t := int(us), int(them)
	return abs(u/8-t/8) + abs(u%8-t%8)
}

// King safety score as a measure of distance from the board center and the number of adjacent enemy pieces and friendly pieces
// TODO: naive initial approach
// use board direction to value own pieces: i.e. a King in front of 3 pawns is not the same as a king behind 3 pawns
// consider using opponent piece attacks around the king instead of actual pieces. use piece weights for opponent threat levels: a queen near our king should be a larger concern than a bishop.
func getKingSafety(b *board.Board, king board.Square, side int) (kingSafety int) {
	kingSafety += 2 * distCenter(king)
	kingSafety += 5*(board.KingSafetyMask[side][king]&b.Occupancy[side]).Count() - 15*(board.KingAttacks[king]&b.Occupancy[side^1]).Count()
	return kingSafety
}

func getKingActivity(b *board.Board, king board.Square, side int) (kingActivity int) {
	kingActivity = -(distCenter(king) + distSquares(king, board.Square(b.Pieces[side^1][board.KINGS].LS1B())))
	return kingActivity
}

func pawnEval(b *board.Board, sq board.Square, side int, _ board.BBoard) int {
	var value int
	if IsProtected(b, sq, side) {
		value = W_P_PROTECTED
	}
	if IsDoubled(b, sq, side) {
		value += W_P_DOUBLED
	}

	if IsIsolated(b, sq, side) {
		value += W_P_ISOLATED
	}
	if IsPassed(b, sq, side) {
		rank := 7 - int(sq)/8
		if side == board.BLACK {
			rank = int(sq) / 8
		}
		value += passedPawnBonus[rank]
	}

	return value
}

func knightEval(b *board.Board, sq board.Square, side int, oppKing board.BBoard) int {
	var eval int
	moves := board.KnightAttacks[sq] & ^b.Occupancy[side]
	if board.Outposts[side][sq]&b.Pieces[side^1][board.PAWNS] == 0 &&
		board.PawnAttacks[side^1][sq]&b.Pieces[side][board.PAWNS] != 0 {
		eval = OutpostsScores[side][board.KNIGHTS][sq]
	}
	return eval + moves.Count()*MOVE_KNIGHT +
		(moves&b.Occupancy[side^1]).Count()*W_CAPTURE +
		(moves&oppKing).Count()*KNIGHT_THREAT
}

func bishopEval(b *board.Board, sq board.Square, side int, oppKing board.BBoard) int {
	var eval int
	moves := board.GetBishopAttacks(int(sq), b.Occupancy[board.BOTH])

	if board.Outposts[side][sq]&b.Pieces[side^1][board.PAWNS] == 0 &&
		board.PawnAttacks[side^1][sq]&b.Pieces[side][board.PAWNS] != 0 {
		eval = OutpostsScores[side][board.BISHOPS][sq]
	}
	if b.Pieces[side][board.BISHOPS].Count() > 1 {
		eval += 35
	}
	return eval + moves.Count()*MOVE_BISHOP +
		(moves&b.Occupancy[side^1]).Count()*W_CAPTURE +
		(moves&oppKing).Count()*BISHOP_THREAT
}

// Evaluation for rooks - mobility, captures, king threats, and (semi)open files.
func rookEval(b *board.Board, sq board.Square, side int, oppKing board.BBoard) int {
	moves := board.GetRookAttacks(int(sq), b.Occupancy[board.BOTH])
	eval := moves.Count()*MOVE_ROOK +
		(moves&b.Occupancy[side^1]).Count()*W_CAPTURE +
		(moves&oppKing).Count()*ROOK_THREAT

	file := board.FileMasks[sq%8]
	if file&b.Pieces[side][board.PAWNS] == 0 {
		if file&b.Pieces[side^1][board.PAWNS] == 0 {
			eval += W_ROOK_OPEN_FILE
		} else {
			eval += W_ROOK_SEMI_OPEN_FILE
		}
	}

	return eval
}

func queenEval(b *board.Board, sq board.Square, side int, oppKing board.BBoard) int {
	moves := board.GetQueenAttacks(int(sq), b.Occupancy[board.BOTH])
	return moves.Count()*MOVE_QUEEN +
		(moves&b.Occupancy[side^1]).Count()*W_CAPTURE +
		(moves&oppKing).Count()*QUEEN_THREAT
}

func kingEval(b *board.Board, king board.Square, side int, _ board.BBoard) int {
	moves := board.KingAttacks[king] & ^b.Occupancy[side]
	return ((getKingSafety(b, king, side)+moves.Count()*MOVE_KING)*(256-b.Phase) + (getKingActivity(b, king, side)-moves.Count()*MOVE_KING)*b.Phase) / 256
}

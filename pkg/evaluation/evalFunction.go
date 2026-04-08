package eval

import (
	"github.com/likeawizard/tofiks/pkg/board"
)

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
	// Mobility related weights (texel-tuned).
	MOVE_QUEEN  = 2
	MOVE_ROOK   = 3
	MOVE_BISHOP = 6
	MOVE_KNIGHT = -2
	MOVE_KING   = -5
	W_CAPTURE   = 8

	// King threat weights (texel-tuned).
	QUEEN_THREAT  = 16
	ROOK_THREAT   = 1
	BISHOP_THREAT = 7
	KNIGHT_THREAT = 3

	// Pawn structure weights (texel-tuned).
	W_P_PROTECTED      = 10
	W_P_DOUBLED        = -12
	W_P_ISOLATED       = -5
	W_P_BACKWARD       = -1
	W_P_BLOCKED        = -5
	W_P_CONNECTED_PASS = 11
	W_P_CANDIDATE      = 8

	// Rook file bonuses (texel-tuned).
	W_ROOK_OPEN_FILE      = 23
	W_ROOK_SEMI_OPEN_FILE = 17

	// Bishop pair bonus (texel-tuned).
	W_BISHOP_PAIR = 15

	// King safety (MG) weights (texel-tuned).
	KS_DIST_CENTER = 24
	KS_PAWN_SHIELD = 31
	KS_FRIENDLY    = 8

	// King activity (EG) weights (texel-tuned).
	KA_DIST_CENTER  = -24
	KA_DIST_SQUARES = -1

	// Passed pawn bonus (texel-tuned, ranks 1-2 clamped to 0).
	PassedPawnBonus = [8]int{0, 0, 0, 20, 44, 67, 156, 0}
)

// Piece protected a pawn.
func IsProtected(b *board.Board, sq board.Square, side int) bool {
	return board.PawnAttacks[side^1][sq]&b.Pieces[side][board.PAWNS] != 0
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
	return b.Pieces[side^1][board.PAWNS]&board.PassedPawns[side][sq] == 0
}

func (e *Engine) GetEvaluation(b *board.Board) int {
	e.Stats.evals++
	e.Board.Phase = e.Board.GetGamePhase()

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
					// Structure already evaluated via pawn table.
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

func DistCenter(sq board.Square) int {
	return dist[sq]
}

func DistSquares(us, them board.Square) int {
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
	kingSafety += KS_DIST_CENTER * DistCenter(king)
	pawnShield := (board.KingSafetyMask[side][king] & b.Pieces[side][board.PAWNS]).Count()
	kingSafety += KS_PAWN_SHIELD*pawnShield + KS_FRIENDLY*((board.KingSafetyMask[side][king]&b.Occupancy[side]).Count()-pawnShield)
	return kingSafety
}

func getKingActivity(b *board.Board, king board.Square, side int) (kingActivity int) {
	kingActivity = KA_DIST_CENTER*DistCenter(king) + KA_DIST_SQUARES*DistSquares(king, board.Square(b.Pieces[side^1][board.KINGS].LS1B()))
	return kingActivity
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
		eval += W_BISHOP_PAIR
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

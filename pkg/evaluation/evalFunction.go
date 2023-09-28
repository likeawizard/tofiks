package eval

import (
	"github.com/likeawizard/tofiks/pkg/board"
)

var PieceWeights = [6]int{100, 325, 325, 500, 975, 10000}

// Based on L. Kaufman - rook and knight values are adjusted by the number of pawns on the board.
var PiecePawnBonus = [6][9]int{
	{},
	{},
	{-25, -19, -13, -6, 0, 6, 13, 19, 25},
	{50, 37, 25, 12, 0, -12, -25, -37, -50},
	{},
	{},
}

const (
	// Mobility related weights.
	W_MOVE    int = 2
	W_CAPTURE int = 4

	// Knight.
	W_N_C22       int = 30
	W_N_C44       int = 20
	W_N_INNER_RIM int = -5
	W_N_OUTER_RIM int = -20

	// Bishop.
	W_B_MAJD int = 20
	W_B_MIND int = 10

	// Pawn.
	W_P_PASSED    int = 10
	W_P_PROTECTED int = 15
	W_P_DOUBLED   int = -15
	W_P_ISOLATED  int = -20
)

type pieceEvalFn func(*board.Board, board.Square, int) int

var pieceEvals = [6]pieceEvalFn{pawnEval, bishopEval, knightEval, rookEval, queenEval, kingEval}

func pawnEval(b *board.Board, sq board.Square, side int) int {
	value := 0
	if IsProtected(b, sq, side) {
		value += W_P_PROTECTED
	}
	if IsDoubled(b, sq, side) {
		value += W_P_DOUBLED
	}

	if IsIsolated(b, sq, side) {
		value += W_P_ISOLATED
	}
	if IsPassed(b, sq, side) {
		value += W_P_PASSED
	}

	return value
}

func queenEval(b *board.Board, sq board.Square, side int) int {
	moves := board.GetQueenAttacks(int(sq), b.Occupancy[board.BOTH])
	return moves.Count()*W_MOVE + (moves&b.Occupancy[side^1]).Count()*W_MOVE
}

// TODO: combine all pawn functions in one with multi value return
// Piece protected a pawn.
func IsProtected(b *board.Board, sq board.Square, side int) bool {
	return board.PawnAttacks[side^1][sq]^b.Pieces[side][board.PAWNS] != 0
}

// TODO: create a lookup table for files to avoid branching.
func IsDoubled(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.PAWNS]&board.DoubledPawns[sq] != 0
}

// Has no friendly pawns on neighboring files.
func IsIsolated(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.PAWNS]&board.IsolatedPawns[sq] == 0
}

// Has no opponent opposing pawns in front (same or neighbor files)
// TODO: stub.
func IsPassed(b *board.Board, sq board.Square, side int) bool {
	return b.Pieces[side][board.PAWNS]&board.PassedPawns[side][sq] == 0
}

func knightEval(b *board.Board, sq board.Square, side int) int {
	moves := board.KnightAttacks[sq] & ^b.Occupancy[side]
	return moves.Count()*W_MOVE + (moves&b.Occupancy[side^1]).Count()*W_CAPTURE
}

func bishopPairEval(b *board.Board, side int) int {
	if b.Pieces[side][board.BISHOPS].Count() > 1 {
		return 50
	}
	return 0
}

func bishopEval(b *board.Board, sq board.Square, side int) int {
	moves := board.GetBishopAttacks(int(sq), b.Occupancy[board.BOTH])
	return bishopPairEval(b, side) + moves.Count()*W_MOVE + (moves&b.Occupancy[side^1]).Count()*W_CAPTURE +
		(board.SquareBitboards[sq]&board.MajorDiag).Count()*W_B_MAJD +
		(board.SquareBitboards[sq]&board.MinorDiag).Count()*W_B_MIND
}

func (e *EvalEngine) GetEvaluation(b *board.Board) int {
	e.Stats.evals++
	var eval, pieceEval int

	// TODO: ensure no move gen is dependent on b.IsWhite internally
	side := -1
	for color := board.WHITE; color <= board.BLACK; color++ {
		side *= -1
		numPawns := b.Pieces[color][board.PAWNS].Count()
		for pieceType := board.PAWNS; pieceType <= board.KINGS; pieceType++ {
			pieces := b.Pieces[color][pieceType]
			for pieces > 0 {
				piece := pieces.PopLS1B()
				pieceEval = PieceWeights[pieceType]
				// Tapered eval - more bias towards PST in the opening and more bias to individual eval functions towards the endgame
				pieceEval += (PST[0][color][pieceType][piece]*(256-b.Phase)+
					(PST[1][color][pieceType][piece]+PiecePawnBonus[pieceType][numPawns])*b.Phase)/256 +
					pieceEvals[pieceType](b, board.Square(piece), color)
				eval += side * pieceEval
			}
		}
	}

	return eval
}

// TODO: try branchless: eliminate min/max and use branchless abs().
func distCenter(sq board.Square) int {
	c := int(sq)
	return max(3-c/8, c/8-4) + max(3-c%8, c%8-4)
}

func distSqares(us, them board.Square) int {
	u, t := int(us), int(them)
	return max((u-t)/8, (t-u)/8) + max((u-t)%8, (t-u)%8)
}

// King safety score as a measure of distance from the board center and the number of adjacent enemy pieces and friendly pieces
// TODO: naive initial approach
// use board direction to value own pieces: i.e. a King in front of 3 pawns is not the same as a king behind 3 pawns
// consider using opponent piece attacks around the king instead of actual pieces. use piece weights for opponent threat levels: a queen near our king should be a larger concern than a bishop.
func getKingSafety(b *board.Board, king board.Square, side int) (kingSafety int) {
	kingSafety += 2 * distCenter(king)
	kingSafety += 5*(board.KingSafetyMask[side][king]&b.Occupancy[side]).Count() - 15*(board.KingAttacks[king]&b.Occupancy[side^1]).Count()
	return
}

func getKingActivity(b *board.Board, king board.Square) (kingActivity int) {
	oppKing := b.Pieces[b.Side^1][board.KINGS].LS1B()
	kingActivity = -(distCenter(king) + distSqares(king, board.Square(oppKing)))
	return
}

func kingEval(b *board.Board, king board.Square, side int) int {
	return (getKingSafety(b, king, side)*(256-b.Phase) + getKingActivity(b, king)*b.Phase) / 256
}

// Evaluation for rooks - connected & (semi)open files.
func rookEval(b *board.Board, sq board.Square, side int) (rookScore int) {
	moves := board.GetRookAttacks(int(sq), b.Occupancy[board.BOTH])
	rookScore = moves.Count()*W_MOVE + (moves&b.Occupancy[side^1]).Count()*W_CAPTURE
	return
}

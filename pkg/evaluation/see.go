package eval

import "github.com/likeawizard/tofiks/pkg/board"

var seeValues = [7]int{100, 325, 325, 500, 975, 10000, 0}

// SEE evaluates a capture on toSq by the piece moving from fromSq.
// Returns the material gain/loss of the full exchange sequence.
func (e *Engine) SEE(fromSq, toSq board.Square) int {
	b := e.Board
	var gain [32]int
	depth := 0
	side := b.Side

	occ := b.Occupancy[board.BOTH]

	attackerPiece := e.pieceOnSquare(fromSq, side)
	gain[0] = seeValues[b.PieceAtSquare(toSq)]

	for {
		depth++
		gain[depth] = seeValues[attackerPiece] - gain[depth-1]

		if max(-gain[depth-1], gain[depth]) < 0 {
			break
		}

		occ &^= board.SquareBitboards[fromSq]
		side ^= 1

		attackerPiece, fromSq = e.leastValuableAttacker(toSq, side, occ)
		if attackerPiece == board.NO_PIECE {
			break
		}
	}

	for depth--; depth > 0; depth-- {
		gain[depth-1] = -max(-gain[depth-1], gain[depth])
	}

	return gain[0]
}

func (e *Engine) pieceOnSquare(sq board.Square, side int8) int {
	bb := board.SquareBitboards[sq]
	for piece := board.PAWNS; piece <= board.KINGS; piece++ {
		if e.Board.Pieces[side][piece]&bb != 0 {
			return piece
		}
	}
	return board.NO_PIECE
}

func (e *Engine) leastValuableAttacker(toSq board.Square, side int8, occ board.BBoard) (int, board.Square) {
	b := e.Board
	sq := int(toSq)

	attackers := board.PawnAttacks[side^1][toSq] & b.Pieces[side][board.PAWNS] & occ
	if attackers != 0 {
		return board.PAWNS, board.Square(attackers.LS1B())
	}

	attackers = board.KnightAttacks[toSq] & b.Pieces[side][board.KNIGHTS] & occ
	if attackers != 0 {
		return board.KNIGHTS, board.Square(attackers.LS1B())
	}

	bishopAttacks := board.GetBishopAttacks(sq, occ)
	attackers = bishopAttacks & b.Pieces[side][board.BISHOPS] & occ
	if attackers != 0 {
		return board.BISHOPS, board.Square(attackers.LS1B())
	}

	rookAttacks := board.GetRookAttacks(sq, occ)
	attackers = rookAttacks & b.Pieces[side][board.ROOKS] & occ
	if attackers != 0 {
		return board.ROOKS, board.Square(attackers.LS1B())
	}

	attackers = (bishopAttacks | rookAttacks) & b.Pieces[side][board.QUEENS] & occ
	if attackers != 0 {
		return board.QUEENS, board.Square(attackers.LS1B())
	}

	attackers = board.KingAttacks[toSq] & b.Pieces[side][board.KINGS] & occ
	if attackers != 0 {
		return board.KINGS, board.Square(attackers.LS1B())
	}

	return board.NO_PIECE, 0
}

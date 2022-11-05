package board

import (
	"fmt"
	"math/bits"
	"strings"
)

// Get a human readable string represantiation of a bitboard.
func (bb BBoard) String() string {
	s := ""
	for r := 0; r < 8; r++ {
		s += fmt.Sprintf(" %d ", 8-r)
		for f := 0; f < 8; f++ {
			sq := r*8 + f
			s += fmt.Sprintf(" %d", bb.Get(sq))
		}
		s += "\n"
	}
	s += "\n    a b c d e f g h"
	s += fmt.Sprintf("\n\n Bitboard: %d", bb)
	return s
}

// Get the bit at position
func (bb *BBoard) Get(sq int) BBoard {
	return *bb >> sq & 1
}

// Set a bit to one at position
func (bb *BBoard) Set(sq int) {
	*bb |= SquareBitboards[sq]
}

// Set a bit to zero at position
func (bb *BBoard) Clear(sq int) {
	*bb &= ^SquareBitboards[sq]
}

// Return population count (number of 1's)
func (bb BBoard) Count() int {
	return bits.OnesCount64(uint64(bb))
}

// Get the position of the Least Signficant
func (bb BBoard) LS1B() int {
	return bits.TrailingZeros64(uint64(bb))
}

func (bb *BBoard) PopLS1B() int {
	ls1b := bits.TrailingZeros64(uint64(*bb))
	bb.Clear(ls1b)
	return ls1b
}

// Get bishop attack mask with blocker occupancy
func GetBishopAttacks(sq int, occ BBoard) BBoard {
	occ &= BishopAttackMasks[sq]
	occ *= BishopMagics[sq]
	occ >>= 64 - BishopOccBitCount[sq]
	return BishopAttacks[sq][occ]
}

// Get Rook attack mask with blocker occupancy
func GetRookAttacks(sq int, occ BBoard) BBoard {
	occ &= RookAttackMasks[sq]
	occ *= RookMagics[sq]
	occ >>= 64 - RookOccBitCount[sq]
	return RookAttacks[sq][occ]
}

// Get Queen attacks as a Bishop and Rook superposition
func GetQueenAttacks(sq int, occ BBoard) BBoard {
	return GetBishopAttacks(sq, occ) | GetRookAttacks(sq, occ)
}

// Get pinned piece square and pin attack mask which are the only legal destination squares for the pinned piece.
// Basic assumptions:
// A piece can only be pinned by one attacker.
// An attacker that delivers check can not also pin a piece.
// A knight can never unpin itself. A pinned knight has no legal moves.
// A bishop can not unpin itself from rook attacks and vice versa.
func (b *Board) GetPinsBB(side int) map[int]BBoard {
	king := b.Pieces[side][KINGS].LS1B()
	pins := make(map[int]BBoard)
	var directAttackers, xRayAttackers, attackMask, pinnedPieces BBoard
	var attackerSq int
	directAttackers = GetBishopAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][BISHOPS] | b.Pieces[side^1][QUEENS])
	xRayAttackers = GetBishopAttacks(king, b.Occupancy[side^1]) & (b.Pieces[side^1][BISHOPS] | b.Pieces[side^1][QUEENS]) &^ directAttackers
	for xRayAttackers > 0 {
		attackerSq = xRayAttackers.PopLS1B()
		attackMask = (GetBishopAttacks(attackerSq, b.Pieces[side][KINGS]) & GetBishopAttacks(king, SquareBitboards[attackerSq])) | SquareBitboards[attackerSq]&^b.Pieces[side][KINGS]
		pinnedPieces = attackMask & b.Occupancy[side]
		if pinnedPieces > 0 && pinnedPieces.Count() == 1 {
			pins[pinnedPieces.LS1B()] = attackMask
		}
	}

	directAttackers = GetRookAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][ROOKS] | b.Pieces[side^1][QUEENS])
	xRayAttackers = GetRookAttacks(king, b.Occupancy[side^1]) & (b.Pieces[side^1][ROOKS] | b.Pieces[side^1][QUEENS]) &^ directAttackers
	for xRayAttackers > 0 {
		attackerSq = xRayAttackers.PopLS1B()
		attackMask = (GetRookAttacks(attackerSq, b.Pieces[side][KINGS]) & GetRookAttacks(king, SquareBitboards[attackerSq])) | SquareBitboards[attackerSq]&^b.Pieces[side][KINGS]
		pinnedPieces = attackMask & b.Occupancy[side]
		if pinnedPieces > 0 && pinnedPieces.Count() == 1 {
			pins[pinnedPieces.LS1B()] = attackMask
		}
	}
	return pins
}

// Get checkers and check attack vectors and true if the check is a double check. A zero bitboard indicates no check.
// Slider piece checks return a bitboard containing squares that are legal destinations which either capture the checker or block its attack.
// A knight checker returns only the position of the knight to be captured as blocking is impossible unlike sliding pieces.
// In case of a double check only the king can move and the resulting bitboard can not be used for determining the legality of other piece moves.
func (b *Board) GetChecksBB(side int) (BBoard, bool) {
	var numChecks int
	var checks, attacker BBoard
	var pawnCheck bool
	king := b.Pieces[side][KINGS].LS1B()

	attacker = PawnAttacks[side][king] & b.Pieces[side^1][PAWNS]
	if attacker != 0 {
		pawnCheck = true
		checks |= attacker
		numChecks++
	}

	attacker = GetRookAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][ROOKS] | b.Pieces[side^1][QUEENS])
	if attacker != 0 {
		checks |= (GetRookAttacks(attacker.LS1B(), b.Pieces[side][KINGS]) & GetRookAttacks(king, attacker)) | attacker&^b.Pieces[side][KINGS]
		numChecks += attacker.Count()
	}

	// A pawn can check by moving forward or capturing. Only a capture move that clears a file for a rook attack can create a double check. So only check Knight and Bishop checks if no pawn check is present
	if !pawnCheck {
		attacker = KnightAttacks[king] & b.Pieces[side^1][KNIGHTS]
		if attacker != 0 {
			checks |= attacker
			numChecks++
		}

		attacker = GetBishopAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][BISHOPS] | b.Pieces[side^1][QUEENS])
		if attacker != 0 {
			checks |= (GetBishopAttacks(attacker.LS1B(), b.Pieces[side][KINGS]) & GetBishopAttacks(king, attacker)) | attacker&^b.Pieces[side][KINGS]
			numChecks++
		}
	}

	return checks, numChecks > 1
}

// Determine if a square is attacked by the opposing side
func (b *Board) IsAttacked(sq, side int, occ BBoard) bool {
	var isAttacked bool

	if PawnAttacks[side][sq]&b.Pieces[side^1][PAWNS] != 0 {
		return true
	}

	if KnightAttacks[sq]&b.Pieces[side^1][KNIGHTS] != 0 {
		return true
	}

	if KingAttacks[sq]&b.Pieces[side^1][KINGS] != 0 {
		return true
	}

	if GetBishopAttacks(sq, occ)&(b.Pieces[side^1][BISHOPS]|b.Pieces[side^1][QUEENS]) != 0 {
		return true
	}

	if GetRookAttacks(sq, occ)&(b.Pieces[side^1][ROOKS]|b.Pieces[side^1][QUEENS]) != 0 {
		return true
	}

	return isAttacked
}

// Determine if the king for the given side is in check
func (b *Board) IsChecked(side int) bool {
	king := b.Pieces[side][KINGS].LS1B()

	return b.IsAttacked(king, side, b.Occupancy[BOTH])
}

// Get a bitboard of all the squares attacked by the opposition
func (b *Board) AttackedSquares(side int, occ BBoard) BBoard {
	attacked := BBoard(0)

	for sq := 0; sq < 64; sq++ {
		if b.IsAttacked(sq, side, occ) {
			attacked = attacked | SquareBitboards[sq]
		}
	}

	return attacked
}

func (b *Board) PseudoMoveGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	var moves []Move
	var move Move

	offset := Move(0)
	if b.Side == BLACK {
		offset = 6
	}

	if b.Side == 0 {
		pieces = b.Pieces[WHITE][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = PawnAttacks[WHITE][from] & b.Occupancy[BLACK]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | IS_CAPTURE | 1<<12

				if from >= A7 && from <= H7 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(to|from<<6) | 1<<12
				if from >= A7 && from <= H7 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 16
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from-8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(to|from<<6)|IS_DOUBLE|1<<12)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(int(b.EnPassantTarget)|from<<6) | IS_CAPTURE | IS_ENPASSANT | 1<<12
				moves = append(moves, move)
			}
		}

	} else {
		pieces = b.Pieces[BLACK][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()

			attacks = PawnAttacks[BLACK][from] & b.Occupancy[WHITE]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | IS_CAPTURE | 7<<12

				if from >= A2 && from <= H2 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(to|from<<6) | 7<<12
				if from >= A2 && from <= H2 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 16
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from+8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(to|from<<6)|IS_DOUBLE|7<<12)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(int(b.EnPassantTarget)|from<<6) | IS_CAPTURE | IS_ENPASSANT | 7<<12
				moves = append(moves, move)
			}
		}
	}

	pieces = b.Pieces[b.Side][KNIGHTS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = KnightAttacks[from] & ^b.Occupancy[b.Side]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (3+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)

		}
	}

	pieces = b.Pieces[b.Side][BISHOPS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (2+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)

		}
	}

	pieces = b.Pieces[b.Side][ROOKS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (4+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)

		}
	}

	pieces = b.Pieces[b.Side][QUEENS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (5+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)

		}
	}

	return append(moves, b.MoveGenKing()...)
}

func (b *Board) PseudoCaptureGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	var moves []Move
	var move Move

	offset := Move(0)
	if b.Side == BLACK {
		offset = 6
	}

	if b.Side == 0 {
		pieces = b.Pieces[WHITE][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = PawnAttacks[WHITE][from] & b.Occupancy[BLACK]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | IS_CAPTURE | 1<<12

				if from >= A7 && from <= H7 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(int(b.EnPassantTarget)|from<<6) | IS_CAPTURE | IS_ENPASSANT | 1<<12
				moves = append(moves, move)
			}
		}

	} else {
		pieces = b.Pieces[BLACK][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()

			attacks = PawnAttacks[BLACK][from] & b.Occupancy[WHITE]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | IS_CAPTURE | 7<<12

				if from >= A2 && from <= H2 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(int(b.EnPassantTarget)|from<<6) | IS_CAPTURE | IS_ENPASSANT | 7<<12
				moves = append(moves, move)
			}
		}
	}

	pieces = b.Pieces[b.Side][KNIGHTS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = KnightAttacks[from] & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (3+offset)<<12 | IS_CAPTURE
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][BISHOPS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (2+offset)<<12 | IS_CAPTURE
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][ROOKS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (4+offset)<<12 | IS_CAPTURE
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][QUEENS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (5+offset)<<12 | IS_CAPTURE
			moves = append(moves, move)
		}
	}

	king := b.Pieces[b.Side][KINGS].LS1B()
	attacks = KingAttacks[king] & b.Occupancy[b.Side^1]
	for attacks > 0 {
		to = attacks.PopLS1B()
		move = Move(to|king<<6) | (6+offset)<<12 | IS_CAPTURE
		moves = append(moves, move)
	}

	return moves
}

// Generate all legal moves for the current side to move
func (b *Board) MoveGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	var moves []Move
	var move Move
	checks, doubleCheck := b.GetChecksBB(b.Side)

	// In case of a double check only King moves are legal
	if doubleCheck {
		return b.MoveGenKing()
	}

	pins := b.GetPinsBB(b.Side)

	if b.Side == 0 {
		pieces = b.Pieces[WHITE][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()

			legalDestinations := ^BBoard(0)
			if checks != 0 {
				legalDestinations &= checks
			}
			if pin, ok := pins[from]; ok {
				legalDestinations &= pin
			}

			attacks = PawnAttacks[WHITE][from] & b.Occupancy[BLACK] & legalDestinations
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | IS_CAPTURE | 1<<12

				if from >= A7 && from <= H7 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to]&legalDestinations != 0 {
				move = Move(to|from<<6) | 1<<12
				if from >= A7 && from <= H7 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 16
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from-8]) == 0 && SquareBitboards[to]&legalDestinations != 0 {
				moves = append(moves, Move(to|from<<6)|IS_DOUBLE|1<<12)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(int(b.EnPassantTarget)|from<<6) | IS_CAPTURE | IS_ENPASSANT | 1<<12
				umake := b.MakeMove(move)
				if !b.IsChecked(b.Side ^ 1) {
					moves = append(moves, move)
				}
				umake()
			}
		}

	} else {
		pieces = b.Pieces[BLACK][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()

			legalDestinations := ^BBoard(0)
			if checks != 0 {
				legalDestinations &= checks
			}
			if pin, ok := pins[from]; ok {
				legalDestinations &= pin
			}

			attacks = PawnAttacks[BLACK][from] & b.Occupancy[WHITE] & legalDestinations
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | IS_CAPTURE | 7<<12

				if from >= A2 && from <= H2 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to]&legalDestinations != 0 {
				move = Move(to|from<<6) | 7<<12
				if from >= A2 && from <= H2 {
					moves = append(moves, move|PROMO_QUEEN<<16, move|PROMO_KNIGHT<<16, move|PROMO_ROOK<<16, move|PROMO_BISHOP<<16)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 16
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from+8]) == 0 && SquareBitboards[to]&legalDestinations != 0 {
				moves = append(moves, Move(to|from<<6)|IS_DOUBLE|7<<12)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(int(b.EnPassantTarget)|from<<6) | IS_CAPTURE | IS_ENPASSANT | 7<<12
				umake := b.MakeMove(move)
				if !b.IsChecked(b.Side ^ 1) {
					moves = append(moves, move)
				}
				umake()
			}
		}
	}

	offset := Move(0)
	if b.Side == BLACK {
		offset = 6
	}

	pieces = b.Pieces[b.Side][KNIGHTS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = KnightAttacks[from] & ^b.Occupancy[b.Side]
		if checks != 0 {
			attacks &= checks
		}
		if pin, ok := pins[from]; ok {
			attacks &= pin
		}
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (3+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)

		}
	}

	pieces = b.Pieces[b.Side][BISHOPS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		if checks != 0 {
			attacks &= checks
		}
		if pin, ok := pins[from]; ok {
			attacks &= pin
		}
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (2+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)

		}
	}

	pieces = b.Pieces[b.Side][ROOKS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		if checks != 0 {
			attacks &= checks
		}
		if pin, ok := pins[from]; ok {
			attacks &= pin
		}
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (4+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)

		}
	}

	pieces = b.Pieces[b.Side][QUEENS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		if checks != 0 {
			attacks &= checks
		}
		if pin, ok := pins[from]; ok {
			attacks &= pin
		}
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(to|from<<6) | (5+offset)<<12
			if b.Occupancy[b.Side^1].Get(to) != 0 {
				move |= IS_CAPTURE
			}
			moves = append(moves, move)
		}
	}

	moves = append(moves, b.MoveGenKing()...)

	return moves
	// return b.RemoveIllegal(moves)
}

// Return king moves for the current side to move
func (b *Board) MoveGenKing() []Move {
	var from, to int
	var pieces, attacks, attackedSquares BBoard
	var isInCheck bool
	var moves []Move
	var move Move
	if b.Side == WHITE {
		attackedSquares = b.AttackedSquares(b.Side, b.Occupancy[BOTH]&^b.Pieces[b.Side][KINGS])
		pieces = b.Pieces[WHITE][KINGS]
		isInCheck = attackedSquares&pieces != 0
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = KingAttacks[from] & ^(b.Occupancy[WHITE] | attackedSquares)
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | 6<<12
				if b.Occupancy[BLACK].Get(to) != 0 {
					move |= IS_CAPTURE
				}
				moves = append(moves, move)

			}
		}

		if !isInCheck {
			if b.CastlingRights&WOO != 0 && (b.Occupancy[BOTH]|attackedSquares)&F1G1 == 0 {
				moves = append(moves, WCastleKing)
			}
			if b.CastlingRights&WOOO != 0 && b.Occupancy[BOTH]&D1B1 == 0 && attackedSquares&D1C1 == 0 {
				moves = append(moves, WCastleQueen)
			}
		}

	} else {
		attackedSquares = b.AttackedSquares(b.Side, b.Occupancy[BOTH]&^b.Pieces[b.Side][KINGS])
		pieces = b.Pieces[BLACK][KINGS]
		isInCheck = attackedSquares&pieces != 0
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = KingAttacks[from] & ^(b.Occupancy[BLACK] | attackedSquares)
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(to|from<<6) | 12<<12
				if b.Occupancy[WHITE].Get(to) != 0 {
					move |= IS_CAPTURE
				}
				moves = append(moves, move)

			}
		}

		if !isInCheck {
			if b.CastlingRights&BOO != 0 && (b.Occupancy[BOTH]|attackedSquares)&F8G8 == 0 {
				moves = append(moves, BCastleKing)
			}
			if b.CastlingRights&BOOO != 0 && b.Occupancy[BOTH]&D8B8 == 0 && attackedSquares&D8C8 == 0 {
				moves = append(moves, BCastleQueen)
			}
		}
	}

	return moves
}

// Get all legal capture moves for current side to move
// TODO: performance: implement independetly of MoveGen to reduce redundancy
func (b *Board) MoveGenCaptures() []Move {
	all := b.MoveGen()
	captures := make([]Move, 0)
	for _, move := range all {
		if move.IsCapture() {
			captures = append(captures, move)
		}
	}
	return captures
}

// Generate a function to return the board state the it's current state
func (b *Board) GetUnmake() func() {
	copy := b.Copy()
	return func() {
		b.Hash = copy.Hash
		b.Pieces = copy.Pieces
		b.Occupancy = copy.Occupancy
		b.Side = copy.Side
		b.CastlingRights = copy.CastlingRights
		b.EnPassantTarget = copy.EnPassantTarget
		b.HalfMoveCounter = copy.HalfMoveCounter
		b.FullMoveCounter = copy.FullMoveCounter
	}
}

// Make a legal move in position and update board state - castling rights, en passant, move count, side to move etc. Returns a function to take back the move made.
func (b *Board) MakeMove(move Move) func() {
	umove := b.GetUnmake()
	if move.IsCapture() || move.Piece() == 1 {
		b.HalfMoveCounter = 0
	} else {
		b.HalfMoveCounter++
	}

	if b.EnPassantTarget > 0 {
		b.ZobristEnPassant(b.EnPassantTarget)
	}

	bitboard := b.GetBitBoard(move.Piece(), move)

	switch {
	case move.IsEnPassant():
		b.ZobristEPCapture(move)
		b.EnPassantTarget = -1
		direction := 8
		if b.Side == WHITE {
			direction = -8
		}
		b.RemoveCaptured(int(move.To()) - direction)
	case move.IsCapture():
		b.EnPassantTarget = -1
		b.ZobristCapture(move)
		b.RemoveCaptured(int(move.To()))
	case move.IsCastling():
		b.EnPassantTarget = -1
		b.ZobristSimpleMove(move)
		b.CompleteCastling(move)
	case move.IsDouble():
		b.ZobristSimpleMove(move)
		b.EnPassantTarget = (move.To() + move.From()) / 2
		b.ZobristEnPassant(b.EnPassantTarget)
	default:
		b.EnPassantTarget = -1
		b.ZobristSimpleMove(move)
	}
	bitboard.Set(int(move.To()))
	bitboard.Clear(int(move.From()))

	b.Promote(move)

	for side := WHITE; side <= BLACK; side++ {
		b.Occupancy[side] = b.Pieces[side][KINGS]
		for piece := PAWNS; piece < KINGS; piece++ {
			b.Occupancy[side] |= b.Pieces[side][piece]
		}
	}
	b.Occupancy[BOTH] = b.Occupancy[WHITE] | b.Occupancy[BLACK]

	b.updateCastlingRights(move)
	if b.Side == BLACK {
		b.FullMoveCounter++
	}

	b.ZobristSideToMove()
	b.Side ^= 1
	return umove
}

// Attempt to play a UCI move in position. Returns unmake closure and ok
func (b *Board) MoveUCI(uciMove string) (func(), bool) {
	all := b.MoveGen()

	for _, move := range all {
		if uciMove == move.String() {
			return b.MakeMove(move), true
		}
	}
	return nil, false
}

// Play out a line of UCI moves in succession. Returns success.
func (b *Board) PlayMovesUCI(uciMoves string) bool {
	moveSlice := strings.Fields(uciMoves)

	for _, uciMove := range moveSlice {
		_, ok := b.MoveUCI(uciMove)
		if !ok {
			return false
		}
	}

	return true
}

// Return a pointer to the bitboard of the piece moved
func (b *Board) GetBitBoard(piece int, move Move) *BBoard {
	side := WHITE
	if piece > 6 {
		side = BLACK
	}
	return &b.Pieces[side][(piece-1)%6]
}

// Remove a piece captured by a move from the opposing bitboard
func (b *Board) RemoveCaptured(sq int) {
	b.Occupancy[b.Side^1].Clear(sq)
	for piece := PAWNS; piece <= KINGS; piece++ {
		b.Pieces[b.Side^1][piece] &= b.Occupancy[b.Side^1]
	}
}

// Make the complimentary rook move when castling
func (b *Board) CompleteCastling(move Move) {
	bitboard := &b.Pieces[b.Side][ROOKS]
	var rookMove Move
	switch move {
	case WCastleKing:
		rookMove = WCastleKingRook
	case WCastleQueen:
		rookMove = WCastleQueenRook
	case BCastleKing:
		rookMove = BCastleKingRook
	case BCastleQueen:
		rookMove = BCastleQueenRook
	}
	b.ZobristSimpleMove(rookMove)
	bitboard.Set(int(rookMove.To()))
	bitboard.Clear(int(rookMove.From()))
}

// Get the piece at square as a collection of values: found, color, piece
func (b *Board) PieceAtSquare(sq Square) (bool, int, int) {
	for color := WHITE; color <= BLACK; color++ {
		for pieceType := PAWNS; pieceType <= KINGS; pieceType++ {
			if SquareBitboards[sq]&b.Pieces[color][pieceType] != 0 {
				return true, color, pieceType
			}
		}
	}

	return false, 0, 0
}

// Replace a pawn on the 8th/1st rank with the promotion piece
func (b *Board) Promote(move Move) {
	if move.Promotion() == 0 {
		return
	}

	var pawnBitBoard, promotionBitBoard *BBoard
	pawnBitBoard = b.GetBitBoard(move.Piece(), move)
	var pieceIdx int

	switch move.Promotion() {
	case PROMO_QUEEN:
		pieceIdx = QUEENS
	case PROMO_KNIGHT:
		pieceIdx = KNIGHTS
	case PROMO_ROOK:
		pieceIdx = ROOKS
	case PROMO_BISHOP:
		pieceIdx = BISHOPS
	}
	promotionBitBoard = &b.Pieces[b.Side][pieceIdx]

	pawnBitBoard.Clear(int(move.To()))
	promotionBitBoard.Set(int(move.To()))
	b.ZobristPromotion(move)
}

// Prune Illegal moves by making the move and verifying that the resulting position doesn't leave own king in check
// TODO: Due to removal. Make the MoveGen generate only legal moves using check and pin restrictions on piece movement
func (b *Board) RemoveIllegal(moves []Move) []Move {
	legal := make([]Move, 0)
	for _, move := range moves {
		umove := b.MakeMove(move)
		if !b.IsChecked(b.Side ^ 1) {
			legal = append(legal, move)
		}
		umove()
	}

	return legal
}

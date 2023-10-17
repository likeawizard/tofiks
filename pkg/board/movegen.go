package board

func (b *Board) PseudoMoveGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	moves := make([]Move, 0, 64)
	var move Move

	if b.Side == 0 {
		pieces = b.Pieces[WHITE][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = PawnAttacks[WHITE][from] & b.Occupancy[BLACK]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(from | to<<toShift | IsCapture | PAWNS<<PieceShift)

				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(from | to<<toShift | PAWNS<<PieceShift)
				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 16
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from-8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift|PAWNS<<PieceShift|IsDouble))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<PieceShift
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
				move = Move(from | to<<toShift | IsCapture | PAWNS<<PieceShift)

				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(from | to<<toShift | PAWNS<<PieceShift)
				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 16
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from+8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift|PAWNS<<PieceShift|IsDouble))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<PieceShift
				moves = append(moves, move)
			}
		}
	}

	enemies := b.Occupancy[b.Side^1]
	var caps, quiets BBoard
	pieces = b.Pieces[b.Side][KNIGHTS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = KnightAttacks[from] & ^b.Occupancy[b.Side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | KNIGHTS<<PieceShift)
		for quiets > 0 {
			to = quiets.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
		move |= IsCapture
		for caps > 0 {
			to = caps.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
	}

	pieces = b.Pieces[b.Side][BISHOPS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | BISHOPS<<PieceShift)
		for quiets > 0 {
			to = quiets.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
		move |= IsCapture
		for caps > 0 {
			to = caps.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
	}

	pieces = b.Pieces[b.Side][ROOKS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | ROOKS<<PieceShift)
		for quiets > 0 {
			to = quiets.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
		move |= IsCapture
		for caps > 0 {
			to = caps.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
	}

	pieces = b.Pieces[b.Side][QUEENS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[b.Side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | QUEENS<<PieceShift)
		for quiets > 0 {
			to = quiets.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
		move |= IsCapture
		for caps > 0 {
			to = caps.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
	}

	return append(moves, b.MoveGenKing()...)
}

func (b *Board) PseudoCaptureAndQueenPromoGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	var moves []Move
	var move Move

	if b.Side == 0 {
		pieces = b.Pieces[WHITE][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = PawnAttacks[WHITE][from] & b.Occupancy[BLACK]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(from | to<<toShift | IsCapture | PAWNS<<PieceShift)

				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift|PAWNS<<PieceShift)
				} else {
					moves = append(moves, move)
				}
			}

			to = from - 8
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift)|QUEENS<<promoShift|PAWNS<<PieceShift)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<PieceShift
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
				move = Move(from | to<<toShift | IsCapture | PAWNS<<PieceShift)

				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift|PAWNS<<PieceShift)
				} else {
					moves = append(moves, move)
				}
			}

			to = from + 8
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift)|QUEENS<<promoShift|PAWNS<<PieceShift)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<PieceShift
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
			move = Move(from | to<<toShift | IsCapture | KNIGHTS<<PieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][BISHOPS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | BISHOPS<<PieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][ROOKS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | ROOKS<<PieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][QUEENS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | QUEENS<<PieceShift)
			moves = append(moves, move)
		}
	}

	king := b.Pieces[b.Side][KINGS].LS1B()
	attacks = KingAttacks[king] & b.Occupancy[b.Side^1]
	for attacks > 0 {
		to = attacks.PopLS1B()
		move = Move(king | to<<toShift | IsCapture | KINGS<<PieceShift)
		moves = append(moves, move)
	}

	return moves
}

// Generate all legal moves for the current side to move.
func (b *Board) MoveGenLegal() []Move {
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
				move = Move(from | to<<toShift | PAWNS<<PieceShift)

				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to]&legalDestinations != 0 {
				move = Move(from | to<<toShift | PAWNS<<PieceShift)
				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 16
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from-8]) == 0 && SquareBitboards[to]&legalDestinations != 0 {
				moves = append(moves, Move(from|to<<toShift|PAWNS<<PieceShift))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | PAWNS<<PieceShift
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
				move = Move(from | to<<toShift | PAWNS<<PieceShift)

				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to]&legalDestinations != 0 {
				move = Move(from | to<<toShift | PAWNS<<PieceShift)
				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 16
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from+8]) == 0 && SquareBitboards[to]&legalDestinations != 0 {
				moves = append(moves, Move(from|to<<toShift|PAWNS<<PieceShift))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | PAWNS<<PieceShift
				umake := b.MakeMove(move)
				if !b.IsChecked(b.Side ^ 1) {
					moves = append(moves, move)
				}
				umake()
			}
		}
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
			move = Move(from | to<<toShift | KNIGHTS<<PieceShift)
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
			move = Move(from | to<<toShift | BISHOPS<<PieceShift)
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
			move = Move(from | to<<toShift | ROOKS<<PieceShift)
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
			move = Move(from | to<<toShift | QUEENS<<PieceShift)
			moves = append(moves, move)
		}
	}

	moves = append(moves, b.MoveGenKing()...)

	return moves
	// return b.RemoveIllegal(moves)
}

// Return king moves for the current side to move.
func (b *Board) MoveGenKing() []Move {
	var from, to int
	var pieces, attacks, attackedSquares BBoard
	var isInCheck bool
	moves := make([]Move, 0, 9)
	var move Move
	enemies := b.Occupancy[b.Side^1]
	var caps, quiets BBoard
	if b.Side == WHITE {
		attackedSquares = b.AttackedSquares(b.Side, b.Occupancy[BOTH]&^b.Pieces[b.Side][KINGS])
		pieces = b.Pieces[WHITE][KINGS]
		isInCheck = attackedSquares&pieces != 0
		from = pieces.PopLS1B()
		attacks = KingAttacks[from] & ^(b.Occupancy[WHITE] | attackedSquares)
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | KINGS<<PieceShift)
		for quiets > 0 {
			to = quiets.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
		move |= IsCapture
		for caps > 0 {
			to = caps.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
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
		from = pieces.PopLS1B()
		attacks = KingAttacks[from] & ^(b.Occupancy[BLACK] | attackedSquares)
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | KINGS<<PieceShift)
		for quiets > 0 {
			to = quiets.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
		}
		move |= IsCapture
		for caps > 0 {
			to = caps.PopLS1B()
			moves = append(moves, move|Move(to<<toShift))
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

package board

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

func (b *Board) PseudoCaptureAndQueenPromoGen() []Move {
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
					moves = append(moves, move|PROMO_QUEEN<<16)
				} else {
					moves = append(moves, move)
				}
			}

			to = from - 8
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(to|from<<6)|1<<12|PROMO_QUEEN<<16)
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
					moves = append(moves, move|PROMO_QUEEN<<16)
				} else {
					moves = append(moves, move)
				}
			}

			to = from + 8
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(to|from<<6)|7<<12|PROMO_QUEEN<<16)
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
	all := b.MoveGenLegal()
	captures := make([]Move, 0)
	for _, move := range all {
		if move.IsCapture() {
			captures = append(captures, move)
		}
	}
	return captures
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

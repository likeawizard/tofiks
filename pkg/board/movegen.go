package board

func (b *Board) PseudoMoveGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	moves := make([]Move, 0, 64)
	var move Move
	side := b.Side

	if side == WHITE {
		pieces = b.Pieces[WHITE][PAWNS]
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = PawnAttacks[WHITE][from] & b.Occupancy[BLACK]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(from | to<<toShift | IsCapture | PAWNS<<pieceShift)

				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(from | to<<toShift | PAWNS<<pieceShift)
				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 16
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from-8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift|PAWNS<<pieceShift|IsDouble))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<pieceShift
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
				move = Move(from | to<<toShift | IsCapture | PAWNS<<pieceShift)

				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 8
			if to >= 0 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(from | to<<toShift | PAWNS<<pieceShift)
				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift, move|KNIGHTS<<promoShift, move|ROOKS<<promoShift, move|BISHOPS<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 16
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&(SquareBitboards[to]|SquareBitboards[from+8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift|PAWNS<<pieceShift|IsDouble))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<pieceShift
				moves = append(moves, move)
			}
		}
	}

	enemies := b.Occupancy[side^1]
	var caps, quiets BBoard
	pieces = b.Pieces[side][KNIGHTS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = KnightAttacks[from] & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | KNIGHTS<<pieceShift)
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

	pieces = b.Pieces[side][BISHOPS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | BISHOPS<<pieceShift)
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

	pieces = b.Pieces[side][ROOKS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | ROOKS<<pieceShift)
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

	pieces = b.Pieces[side][QUEENS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[BOTH]) & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | QUEENS<<pieceShift)
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

	from = b.Pieces[side][KINGS].LS1B()
	move = Move(from | KINGS<<pieceShift)
	attacks = KingAttacks[from] & ^b.Occupancy[side]
	caps = attacks & enemies
	quiets = attacks &^ enemies
	for quiets > 0 {
		to = quiets.PopLS1B()
		moves = append(moves, move|Move(to<<toShift))
	}
	move |= IsCapture
	for caps > 0 {
		to = caps.PopLS1B()
		moves = append(moves, move|Move(to<<toShift))
	}

	if !b.InCheck {
		if side == WHITE {
			attackedSquares := b.AttackedSquares(side, F1G1|D1B1|D1C1, b.Occupancy[BOTH]&^pieces)
			if b.CastlingRights&WOO != 0 && (b.Occupancy[BOTH]|attackedSquares)&F1G1 == 0 {
				moves = append(moves, WCastleKing)
			}
			if b.CastlingRights&WOOO != 0 && b.Occupancy[BOTH]&D1B1 == 0 && attackedSquares&D1C1 == 0 {
				moves = append(moves, WCastleQueen)
			}
		} else {
			attackedSquares := b.AttackedSquares(side, F8G8|D8B8|D8C8, b.Occupancy[BOTH]&^pieces)
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
				move = Move(from | to<<toShift | IsCapture | PAWNS<<pieceShift)

				if from >= A7 && from <= H7 {
					moves = append(moves, move|QUEENS<<promoShift|PAWNS<<pieceShift)
				} else {
					moves = append(moves, move)
				}
			}

			to = from - 8
			if from >= A7 && from <= H7 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift)|QUEENS<<promoShift|PAWNS<<pieceShift)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[WHITE][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<pieceShift
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
				move = Move(from | to<<toShift | IsCapture | PAWNS<<pieceShift)

				if from >= A2 && from <= H2 {
					moves = append(moves, move|QUEENS<<promoShift|PAWNS<<pieceShift)
				} else {
					moves = append(moves, move)
				}
			}

			to = from + 8
			if from >= A2 && from <= H2 && b.Occupancy[BOTH]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift)|QUEENS<<promoShift|PAWNS<<pieceShift)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[BLACK][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | PAWNS<<pieceShift
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
			move = Move(from | to<<toShift | IsCapture | KNIGHTS<<pieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][BISHOPS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | BISHOPS<<pieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][ROOKS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | ROOKS<<pieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][QUEENS]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[BOTH]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | QUEENS<<pieceShift)
			moves = append(moves, move)
		}
	}

	king := b.Pieces[b.Side][KINGS].LS1B()
	attacks = KingAttacks[king] & b.Occupancy[b.Side^1]
	for attacks > 0 {
		to = attacks.PopLS1B()
		move = Move(king | to<<toShift | IsCapture | KINGS<<pieceShift)
		moves = append(moves, move)
	}

	return moves
}

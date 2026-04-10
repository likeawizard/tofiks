package board

func (b *Board) PseudoMoveGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	moves := make([]Move, 0, 64)
	var move Move
	side := b.Side

	if side == White {
		pieces = b.Pieces[White][Pawns]
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = PawnAttacks[White][from] & b.Occupancy[Black]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(from | to<<toShift | IsCapture | Pawns<<pieceShift)

				if from >= A7 && from <= H7 {
					moves = append(moves, move|Queens<<promoShift, move|Knights<<promoShift, move|Rooks<<promoShift, move|Bishops<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 8
			if to >= 0 && b.Occupancy[Both]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(from | to<<toShift | Pawns<<pieceShift)
				if from >= A7 && from <= H7 {
					moves = append(moves, move|Queens<<promoShift, move|Knights<<promoShift, move|Rooks<<promoShift, move|Bishops<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from - 16
			if from >= A2 && from <= H2 && b.Occupancy[Both]&(SquareBitboards[to]|SquareBitboards[from-8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift|Pawns<<pieceShift|IsDouble))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[White][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | Pawns<<pieceShift
				moves = append(moves, move)
			}
		}
	} else {
		pieces = b.Pieces[Black][Pawns]
		for pieces > 0 {
			from = pieces.PopLS1B()

			attacks = PawnAttacks[Black][from] & b.Occupancy[White]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(from | to<<toShift | IsCapture | Pawns<<pieceShift)

				if from >= A2 && from <= H2 {
					moves = append(moves, move|Queens<<promoShift, move|Knights<<promoShift, move|Rooks<<promoShift, move|Bishops<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 8
			if to >= 0 && b.Occupancy[Both]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				move = Move(from | to<<toShift | Pawns<<pieceShift)
				if from >= A2 && from <= H2 {
					moves = append(moves, move|Queens<<promoShift, move|Knights<<promoShift, move|Rooks<<promoShift, move|Bishops<<promoShift)
				} else {
					moves = append(moves, move)
				}
			}
			to = from + 16
			if from >= A7 && from <= H7 && b.Occupancy[Both]&(SquareBitboards[to]|SquareBitboards[from+8]) == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift|Pawns<<pieceShift|IsDouble))
			}

			if b.EnPassantTarget > 0 && PawnAttacks[Black][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | Pawns<<pieceShift
				moves = append(moves, move)
			}
		}
	}

	enemies := b.Occupancy[side^1]
	var caps, quiets BBoard
	pieces = b.Pieces[side][Knights]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = KnightAttacks[from] & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | Knights<<pieceShift)
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

	pieces = b.Pieces[side][Bishops]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[Both]) & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | Bishops<<pieceShift)
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

	pieces = b.Pieces[side][Rooks]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[Both]) & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | Rooks<<pieceShift)
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

	pieces = b.Pieces[side][Queens]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[Both]) & ^b.Occupancy[side]
		caps = attacks & enemies
		quiets = attacks &^ enemies
		move = Move(from | Queens<<pieceShift)
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

	from = b.Pieces[side][Kings].LS1B()
	move = Move(from | Kings<<pieceShift)
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
		kingBB := b.Pieces[side][Kings]
		if side == White {
			attackedSquares := b.AttackedSquares(side, F1G1|D1B1|D1C1, b.Occupancy[Both]&^kingBB)
			if b.CastlingRights&WOO != 0 && (b.Occupancy[Both]|attackedSquares)&F1G1 == 0 {
				moves = append(moves, WCastleKing)
			}
			if b.CastlingRights&WOOO != 0 && b.Occupancy[Both]&D1B1 == 0 && attackedSquares&D1C1 == 0 {
				moves = append(moves, WCastleQueen)
			}
		} else {
			attackedSquares := b.AttackedSquares(side, F8G8|D8B8|D8C8, b.Occupancy[Both]&^kingBB)
			if b.CastlingRights&BOO != 0 && (b.Occupancy[Both]|attackedSquares)&F8G8 == 0 {
				moves = append(moves, BCastleKing)
			}
			if b.CastlingRights&BOOO != 0 && b.Occupancy[Both]&D8B8 == 0 && attackedSquares&D8C8 == 0 {
				moves = append(moves, BCastleQueen)
			}
		}
	}

	return moves
}

func (b *Board) PseudoCaptureAndQueenPromoGen() []Move {
	var from, to int
	var pieces, attacks BBoard
	moves := make([]Move, 0, 16)
	var move Move

	if b.Side == 0 {
		pieces = b.Pieces[White][Pawns]
		for pieces > 0 {
			from = pieces.PopLS1B()
			attacks = PawnAttacks[White][from] & b.Occupancy[Black]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(from | to<<toShift | IsCapture | Pawns<<pieceShift)

				if from >= A7 && from <= H7 {
					moves = append(moves, move|Queens<<promoShift|Pawns<<pieceShift)
				} else {
					moves = append(moves, move)
				}
			}

			to = from - 8
			if from >= A7 && from <= H7 && b.Occupancy[Both]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift)|Queens<<promoShift|Pawns<<pieceShift)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[White][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | Pawns<<pieceShift
				moves = append(moves, move)
			}
		}
	} else {
		pieces = b.Pieces[Black][Pawns]
		for pieces > 0 {
			from = pieces.PopLS1B()

			attacks = PawnAttacks[Black][from] & b.Occupancy[White]
			for attacks > 0 {
				to = attacks.PopLS1B()
				move = Move(from | to<<toShift | IsCapture | Pawns<<pieceShift)

				if from >= A2 && from <= H2 {
					moves = append(moves, move|Queens<<promoShift|Pawns<<pieceShift)
				} else {
					moves = append(moves, move)
				}
			}

			to = from + 8
			if from >= A2 && from <= H2 && b.Occupancy[Both]&SquareBitboards[to] == 0 && SquareBitboards[to] != 0 {
				moves = append(moves, Move(from|to<<toShift)|Queens<<promoShift|Pawns<<pieceShift)
			}

			if b.EnPassantTarget > 0 && PawnAttacks[Black][from]&SquareBitboards[b.EnPassantTarget] != 0 {
				move = Move(from|int(b.EnPassantTarget)<<toShift) | IsEnpassant | IsCapture | Pawns<<pieceShift
				moves = append(moves, move)
			}
		}
	}

	pieces = b.Pieces[b.Side][Knights]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = KnightAttacks[from] & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | Knights<<pieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][Bishops]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetBishopAttacks(from, b.Occupancy[Both]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | Bishops<<pieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][Rooks]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetRookAttacks(from, b.Occupancy[Both]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | Rooks<<pieceShift)
			moves = append(moves, move)
		}
	}

	pieces = b.Pieces[b.Side][Queens]
	for pieces > 0 {
		from = pieces.PopLS1B()
		attacks = GetQueenAttacks(from, b.Occupancy[Both]) & b.Occupancy[b.Side^1]
		for attacks > 0 {
			to = attacks.PopLS1B()
			move = Move(from | to<<toShift | IsCapture | Queens<<pieceShift)
			moves = append(moves, move)
		}
	}

	king := b.Pieces[b.Side][Kings].LS1B()
	attacks = KingAttacks[king] & b.Occupancy[b.Side^1]
	for attacks > 0 {
		to = attacks.PopLS1B()
		move = Move(king | to<<toShift | IsCapture | Kings<<pieceShift)
		moves = append(moves, move)
	}

	return moves
}

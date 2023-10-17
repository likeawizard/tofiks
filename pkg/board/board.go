package board

func NewBoard(position string) *Board {
	b := Board{}
	switch position {
	case "startpos", "":
		if err := b.ImportFEN(startingFEN); err != nil {
			panic(err)
		}
	default:
		if err := b.ImportFEN(position); err != nil {
			panic(err)
		}
	}
	return &b
}

func (b *Board) Copy() *Board {
	cp := Board{
		Hash:            b.Hash,
		Pieces:          b.Pieces,
		Occupancy:       b.Occupancy,
		Side:            b.Side,
		Phase:           b.Phase,
		InCheck:         b.InCheck,
		CastlingRights:  b.CastlingRights,
		EnPassantTarget: b.EnPassantTarget,
		HalfMoveCounter: b.HalfMoveCounter,
		FullMoveCounter: b.FullMoveCounter,
	}

	return &cp
}

func (b *Board) IsDouble(move Move, piece int) bool {
	if piece != PAWNS {
		return false
	}
	from, to := move.FromTo()
	return (from/8 == 1 || from/8 == 6) && (to/8 == 3 || to/8 == 4)
}

func (b *Board) updateCastlingRights(move Move) {
	if b.CastlingRights == 0 {
		return
	}
	from, to := move.FromTo()

	switch {
	case b.CastlingRights&(WOOO|WOO) != 0 && from == WCastleQueen.From():
		if b.CastlingRights&WOO != 0 {
			b.ZobristCastlingRights(WOO)
		}
		if b.CastlingRights&WOOO != 0 {
			b.ZobristCastlingRights(WOOO)
		}

		b.CastlingRights &^= WOOO
		b.CastlingRights &^= WOO

	case b.CastlingRights&(BOOO|BOO) != 0 && from == BCastleQueen.From():
		if b.CastlingRights&BOOO != 0 {
			b.ZobristCastlingRights(BOOO)
		}
		if b.CastlingRights&BOO != 0 {
			b.ZobristCastlingRights(BOO)
		}

		b.CastlingRights &^= BOOO
		b.CastlingRights &^= BOO

	case b.CastlingRights&WOOO != 0 && (from == WCastleQueenRook.From() || to == WCastleQueenRook.From()):
		b.ZobristCastlingRights(WOOO)
		b.CastlingRights &^= WOOO

	case b.CastlingRights&WOO != 0 && (from == WCastleKingRook.From() || to == WCastleKingRook.From()):
		b.ZobristCastlingRights(WOO)
		b.CastlingRights &^= WOO

	case b.CastlingRights&BOOO != 0 && (from == BCastleQueenRook.From() || to == BCastleQueenRook.From()):
		b.ZobristCastlingRights(BOOO)
		b.CastlingRights &^= BOOO

	case b.CastlingRights&BOO != 0 && (from == BCastleKingRook.From() || to == BCastleKingRook.From()):
		b.ZobristCastlingRights(BOO)
		b.CastlingRights &^= BOO
	}
}

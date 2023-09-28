package board

var Pieces = [6]string{"P", "B", "N", "R", "Q", "K"}

func NewBoard(position string) *Board {
	b := Board{}
	switch position {
	case "startpos", "":
		b.ImportFEN(startingFEN)
	default:
		b.ImportFEN(position)
	}
	return &b
}

func (b *Board) Copy() *Board {
	copy := Board{
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

	return &copy
}

// Get the type of piece moved.
func (b *Board) Piece(move Move) int {
	from := SquareBitboards[move.From()]
	for bb := PAWNS; bb <= KINGS; bb++ {
		if b.Pieces[0][bb]&from != 0 || b.Pieces[1][bb]&from != 0 {
			return bb
		}
	}

	return 0
}

// Check if destination square is occupied and implies capture.
func (b *Board) IsCapture(move Move) bool {
	return SquareBitboards[move.To()]&b.Occupancy[BOTH] != 0 || move.IsEnPassant()
}

func (b *Board) IsCastling(move Move, piece int) bool {
	if piece != KINGS {
		return false
	}
	switch move {
	case WCastleKing, WCastleQueen, BCastleKing, BCastleQueen:
		return true
	default:
		return false
	}
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

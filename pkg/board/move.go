package board

import (
	"fmt"
	"strconv"
)

var (
	// Castling moves. Used for recognizing castling and moving king during castling
	WCastleKing  = MoveFromString("e1g1") | IS_CASTLING | 6<<12
	WCastleQueen = MoveFromString("e1c1") | IS_CASTLING | 6<<12
	BCastleKing  = MoveFromString("e8g8") | IS_CASTLING | 12<<12
	BCastleQueen = MoveFromString("e8c8") | IS_CASTLING | 12<<12

	// Complimentary castling moves. Used during castling to reposition rook
	WCastleKingRook  = MoveFromString("h1f1") | 4<<12
	WCastleQueenRook = MoveFromString("a1d1") | 4<<12
	BCastleKingRook  = MoveFromString("h8f8") | 10<<12
	BCastleQueenRook = MoveFromString("a8d8") | 10<<12
)

// 0..7 a8 to h8
// 0..63 to a8 to h1 mapping
type Square int

func SquareFromString(s string) Square {
	file := int(s[0] - 'a')
	rank, _ := strconv.Atoi(s[1:])
	rank = 8 - rank
	return Square(file + rank*8)
}

func (s Square) String() string {
	rank := (7 - s/8) + 1
	file := s % 8
	return fmt.Sprintf("%c%d", file+'a', rank)
}

type Move uint64

const (
	PROMO_QUEEN = 1 + iota
	PROMO_KNIGHT
	PROMO_BISHOP
	PROMO_ROOK

	FROM         = 0x3f
	TO           = 0xfc0
	FROMTO       = FROM + TO
	PIECE        = 0xf000
	PROMO        = 0xf0000
	IS_CAPTURE   = 0x100000
	IS_DOUBLE    = 0x200000
	IS_ENPASSANT = 0x400000
	IS_CASTLING  = 0x800000
)

func NewMove(from, to, piece, promo int, capture, double, ep, castling bool) Move {
	move := Move(from | to<<6 | piece<<12)
	if capture {
		move |= IS_CAPTURE
	}
	if double {
		move |= IS_DOUBLE
	}
	if ep {
		move |= IS_ENPASSANT
	}
	if castling {
		move |= IS_CASTLING
	}

	return move
}

func MoveFromString(s string) Move {
	from := SquareFromString(s[:2]) << 6
	to := SquareFromString(s[2:4])
	promotion := 0
	if len(s) == 5 {
		promotion = int(s[4]) << 12
	}
	return Move(from + to + Square(promotion))
}

func MoveFromSquares(from, to Square) Move {
	return Move(to | from<<6)
}

func (m Move) From() Square {
	return Square(m>>6) & 63
}

func (m Move) To() Square {
	return Square(m) & 63
}

func (m Move) Promotion() uint8 {
	return uint8(m & PROMO >> 16)
}

func (m Move) SetPromotion(prom uint8) Move {
	return m&4095 | Move(prom)<<12
}

func (m Move) FromTo() (Square, Square) {
	return Square(m>>6) & 63, Square(m) & 63
}

func (m Move) Reverse() Move {
	return m>>6&63 | m&63<<6

}

func (m Move) IsCastling() bool {
	return m&IS_CASTLING != 0
}

func (m Move) IsCapture() bool {
	return m&IS_CAPTURE != 0
}

func (m Move) IsDouble() bool {
	return m&IS_DOUBLE != 0
}

func (m Move) IsEnPassant() bool {
	return m&IS_ENPASSANT != 0
}

func (m Move) Piece() int {
	return int(m & PIECE >> 12)
}

func (m Move) String() string {
	promo := ""
	switch m.Promotion() {
	case PROMO_BISHOP:
		promo = "b"
	case PROMO_KNIGHT:
		promo = "n"
	case PROMO_QUEEN:
		promo = "q"
	case PROMO_ROOK:
		promo = "r"
	}
	return fmt.Sprintf("%v%v%s", Square(m.From()), Square(m.To()), promo)

}

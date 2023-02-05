package board

import (
	"fmt"
	"strconv"
)

var (
	// Castling moves. Used for recognizing castling and moving king during castling
	WCastleKing  = MoveFromString("e1g1")
	WCastleQueen = MoveFromString("e1c1")
	BCastleKing  = MoveFromString("e8g8")
	BCastleQueen = MoveFromString("e8c8")

	// Complimentary castling moves. Used during castling to reposition rook
	WCastleKingRook  = MoveFromString("h1f1")
	WCastleQueenRook = MoveFromString("a1d1")
	BCastleKingRook  = MoveFromString("h8f8")
	BCastleQueenRook = MoveFromString("a8d8")
)

// 0..7 a8 to h8
// 0..63 to a8 to h1 mapping
type Square int

// LSB 0..5 from 6..11 to 12..14 promotion 15 IS_ENPASSANT MSB
type Move uint16

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

const (
	PROMO_QUEEN = 1 + iota
	PROMO_KNIGHT
	PROMO_BISHOP
	PROMO_ROOK

	FROM         = 0x3f
	TO           = 0xfc0
	FROMTO       = FROM + TO
	PROMO        = 0xf000
	PROMO_SHIFT  = 12
	IS_ENPASSANT = 1 << 15
)

// TODO: delete or fix: ep flag is not being set
func NewMove(from, to, promo int) Move {
	move := Move(from | to<<6 | promo<<PROMO_SHIFT)
	return move
}

func MoveFromString(s string) Move {
	from := SquareFromString(s[:2]) << 6
	to := SquareFromString(s[2:4])
	promotion := 0
	if len(s) == 5 {
		promotion = int(s[4]) << PROMO_SHIFT
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
	return uint8(m & PROMO >> PROMO_SHIFT)
}

func (m Move) SetPromotion(prom uint8) Move {
	return m&4095 | Move(prom)<<PROMO_SHIFT
}

func (m Move) FromTo() (Square, Square) {
	return Square(m>>6) & 63, Square(m) & 63
}

func (m Move) Reverse() Move {
	return m>>6&63 | m&63<<6

}

func (m Move) IsEnPassant() bool {
	return m&IS_ENPASSANT != 0
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

package board

import (
	"fmt"
	"strconv"
)

var (
	// Castling moves. Used for recognizing castling and moving king during castling.
	WCastleKing  = MoveFromString("e1g1") | IsCastling | KINGS<<PieceShift
	WCastleQueen = MoveFromString("e1c1") | IsCastling | KINGS<<PieceShift
	BCastleKing  = MoveFromString("e8g8") | IsCastling | KINGS<<PieceShift
	BCastleQueen = MoveFromString("e8c8") | IsCastling | KINGS<<PieceShift

	// Complimentary castling moves. Used during castling to reposition rook.
	WCastleKingRook  = MoveFromString("h1f1")
	WCastleQueenRook = MoveFromString("a1d1")
	BCastleKingRook  = MoveFromString("h8f8")
	BCastleQueenRook = MoveFromString("a8d8")
)

// 0..7 a8 to h8
// 0..63 to a8 to h1 mapping.
type Square int16

// LSB 0..5 from 6..11 to 12..14 promotion 15 IsEnpassant 16 IsCapture 17 IsCastling 18 IsDouble 19..21 Piece 22..31 unused MSB.
type Move uint32

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
	fromMask   = 1<<6 - 1
	fromShift  = 0
	toMask     = 1<<6 - 1
	toShift    = 6
	fromToMask = 1<<12 - 1
	promoMask  = 1<<3 - 1
	promoShift = 12

	IsEnpassant = 1 << 15
	IsCapture   = 1 << 16
	IsCastling  = 1 << 17
	IsDouble    = 1 << 18
	PieceMask   = 1<<3 - 1
	PieceShift  = 19
)

func MoveFromString(s string) Move {
	from := SquareFromString(s[:2])
	to := SquareFromString(s[2:4])
	promotion := Square(0)
	if len(s) == 5 {
		promotion = Square(s[4])
	}
	return Move(from | to<<toShift | promotion<<promoShift)
}

func (m Move) From() Square {
	return Square(m) & fromMask
}

func (m Move) To() Square {
	return Square(m>>toShift) & toMask
}

func (m Move) Promotion() uint8 {
	return uint8(m>>promoShift) & promoMask
}

func (m Move) SetPromotion(prom uint8) Move {
	return m&fromToMask | Move(prom)<<promoShift
}

func (m Move) FromTo() (Square, Square) {
	return Square(m) & fromMask, Square(m>>toShift) & toMask
}

func (m Move) IsEnPassant() bool {
	return m&IsEnpassant != 0
}

func (m Move) IsCapture() bool {
	return m&IsCapture != 0
}

func (m Move) IsCastling() bool {
	return m&IsCastling != 0
}

func (m Move) IsDouble() bool {
	return m&IsDouble != 0
}

func (m Move) Piece() uint8 {
	return uint8(m>>PieceShift) & PieceMask
}

func (m Move) String() string {
	promo := ""
	switch m.Promotion() {
	case BISHOPS:
		promo = "b"
	case KNIGHTS:
		promo = "n"
	case ROOKS:
		promo = "r"
	case QUEENS:
		promo = "q"
	}
	return fmt.Sprintf("%v%v%s", m.From(), m.To(), promo)
}

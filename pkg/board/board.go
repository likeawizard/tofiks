package board

import (
	"github.com/likeawizard/tofiks/pkg/config"
)

var Pieces = [6]string{"P", "B", "N", "R", "Q", "K"}

func (b *Board) Init(c *config.Config) {
	fen := c.Init.StartingFen
	if fen == "" {
		fen = startingFEN
	}
	b.ImportFEN(fen)
}

func (b *Board) InitDefault() {
	b.ImportFEN(startingFEN)
}

func (b *Board) Copy() *Board {
	copy := Board{
		Hash:            b.Hash,
		Pieces:          b.Pieces,
		Occupancy:       b.Occupancy,
		Side:            b.Side,
		CastlingRights:  b.CastlingRights,
		EnPassantTarget: b.EnPassantTarget,
		HalfMoveCounter: b.HalfMoveCounter,
		FullMoveCounter: b.FullMoveCounter,
	}

	return &copy
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

		b.CastlingRights = b.CastlingRights &^ WOOO
		b.CastlingRights = b.CastlingRights &^ WOO

	case b.CastlingRights&(BOOO|BOO) != 0 && from == BCastleQueen.From():
		if b.CastlingRights&BOOO != 0 {
			b.ZobristCastlingRights(BOOO)
		}
		if b.CastlingRights&BOO != 0 {
			b.ZobristCastlingRights(BOO)
		}

		b.CastlingRights = b.CastlingRights &^ BOOO
		b.CastlingRights = b.CastlingRights &^ BOO

	case b.CastlingRights&WOOO != 0 && (from == WCastleQueenRook.From() || to == WCastleQueenRook.From()):
		b.ZobristCastlingRights(WOOO)
		b.CastlingRights = b.CastlingRights &^ WOOO

	case b.CastlingRights&WOO != 0 && (from == WCastleKingRook.From() || to == WCastleKingRook.From()):
		b.ZobristCastlingRights(WOO)
		b.CastlingRights = b.CastlingRights &^ WOO

	case b.CastlingRights&BOOO != 0 && (from == BCastleQueenRook.From() || to == BCastleQueenRook.From()):
		b.ZobristCastlingRights(BOOO)
		b.CastlingRights = b.CastlingRights &^ BOOO

	case b.CastlingRights&BOO != 0 && (from == BCastleKingRook.From() || to == BCastleKingRook.From()):
		b.ZobristCastlingRights(BOO)
		b.CastlingRights = b.CastlingRights &^ BOO
	}
}

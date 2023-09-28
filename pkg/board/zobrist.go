package board

import (
	"math/rand"
)

var (
	seed          uint64
	pieceKeys     [2][6][64]uint64
	castlingKeys  map[CastlingRights]uint64
	swapSide      uint64
	enPassantKeys [64]uint64
)

func init() {
	seed = rand.Uint64()
	castlingKeys = make(map[CastlingRights]uint64)
	for sq := 0; sq < 64; sq++ {
		for color := WHITE; color <= BLACK; color++ {
			for pieceType := PAWNS; pieceType <= KINGS; pieceType++ {
				pieceKeys[color][pieceType][sq] = rand.Uint64()
			}
		}
		for i := 1; i <= 12; i++ {
		}
		enPassantKeys[sq] = rand.Uint64()
	}

	castlingKeys[WOO] = rand.Uint64()
	castlingKeys[WOOO] = rand.Uint64()
	castlingKeys[BOO] = rand.Uint64()
	castlingKeys[BOOO] = rand.Uint64()

	swapSide = rand.Uint64()
}

// Calculate Zborist hash of the position.
func (b *Board) SeedHash() uint64 {
	hash := seed

	for color := WHITE; color <= BLACK; color++ {
		for pieceType := PAWNS; pieceType <= KINGS; pieceType++ {
			pieces := b.Pieces[color][pieceType]
			for pieces > 0 {
				sq := pieces.PopLS1B()
				hash ^= pieceKeys[color][pieceType][sq]
			}
		}
	}

	for right, cr := range castlingKeys {
		if b.CastlingRights&right != 0 {
			hash ^= cr
		}
	}

	if b.EnPassantTarget != -1 {
		hash ^= enPassantKeys[b.EnPassantTarget]
	}

	if b.Side == BLACK {
		hash ^= swapSide
	}

	return hash
}

// Incrementally update Zborist hash after a move
// TODO: optimize - remove use of expensive PieceAtSquare function.
func (b *Board) ZobristSimpleMove(move Move, piece int) {
	from, to := move.From(), move.To()
	b.Hash ^= pieceKeys[b.Side][piece][to]
	b.Hash ^= pieceKeys[b.Side][piece][from]
}

func (b *Board) ZobristCapture(move Move, piece int) {
	from, to := move.From(), move.To()
	_, _, capturedPiece := b.PieceAtSquare(to)
	b.Hash ^= pieceKeys[b.Side^1][capturedPiece][to]
	b.Hash ^= pieceKeys[b.Side][piece][to]
	b.Hash ^= pieceKeys[b.Side][piece][from]
}

func (b *Board) ZobristEPCapture(move Move) {
	from, to := move.From(), move.To()
	direction := Square(8)
	if b.Side == WHITE {
		direction = -8
	}
	b.Hash ^= pieceKeys[b.Side^1][PAWNS][to-direction]
	b.Hash ^= pieceKeys[b.Side][PAWNS][to]
	b.Hash ^= pieceKeys[b.Side][PAWNS][from]
}

// Update Zobirst hash with flipping side to move.
func (b *Board) ZobristSideToMove() {
	b.Hash ^= swapSide
}

// Update Zobrist hash with castling rights.
func (b *Board) ZobristCastlingRights(right CastlingRights) {
	b.Hash ^= castlingKeys[right]
}

// Update Zobrist hash with castling move.
func (b *Board) ZobristCastling(right CastlingRights) {
	switch right {
	case WOO:
		b.ZobristSimpleMove(WCastleKing, KINGS)
		b.ZobristSimpleMove(WCastleKingRook, ROOKS)
	case WOOO:
		b.ZobristSimpleMove(WCastleQueen, KINGS)
		b.ZobristSimpleMove(WCastleQueenRook, ROOKS)
	case BOO:
		b.ZobristSimpleMove(BCastleKing, KINGS)
		b.ZobristSimpleMove(BCastleKingRook, ROOKS)
	case BOOO:
		b.ZobristSimpleMove(BCastleQueen, KINGS)
		b.ZobristSimpleMove(BCastleQueenRook, ROOKS)
	}
}

// Update Zobrist hash when promoting a piece.
func (b *Board) ZobristPromotion(move Move) {
	var promotion int
	switch move.Promotion() {
	case PROMO_QUEEN:
		promotion = QUEENS
	case PROMO_KNIGHT:
		promotion = KNIGHTS
	case PROMO_ROOK:
		promotion = ROOKS
	case PROMO_BISHOP:
		promotion = BISHOPS
	}
	to := move.To()

	// set destination with newly promoted piece
	b.Hash ^= pieceKeys[b.Side][promotion][to]
	b.Hash ^= pieceKeys[b.Side][PAWNS][to]
}

// Update Zobrist hash with En Passant square.
func (b *Board) ZobristEnPassant(square Square) {
	if b.EnPassantTarget != -1 {
		b.Hash ^= enPassantKeys[square]
	}
}

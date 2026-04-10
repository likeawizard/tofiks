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
	rng := rand.New(rand.NewSource(0x4F4649_4B5321))
	seed = rng.Uint64()
	castlingKeys = make(map[CastlingRights]uint64)
	for sq := 0; sq < 64; sq++ {
		for color := White; color <= Black; color++ {
			for pieceType := Pawns; pieceType <= Kings; pieceType++ {
				pieceKeys[color][pieceType][sq] = rng.Uint64()
			}
		}

		enPassantKeys[sq] = rng.Uint64()
	}

	castlingKeys[WOO] = rng.Uint64()
	castlingKeys[WOOO] = rng.Uint64()
	castlingKeys[BOO] = rng.Uint64()
	castlingKeys[BOOO] = rng.Uint64()

	swapSide = rng.Uint64()
}

// SeedPawnHash calculates a zobrist hash of only the pawn positions.
func (b *Board) SeedPawnHash() uint64 {
	var hash uint64
	for color := White; color <= Black; color++ {
		pieces := b.Pieces[color][Pawns]
		for pieces > 0 {
			sq := pieces.PopLS1B()
			hash ^= pieceKeys[color][Pawns][sq]
		}
	}
	return hash
}

// Calculate Zborist hash of the position.
func (b *Board) SeedHash() uint64 {
	hash := seed

	for color := White; color <= Black; color++ {
		for pieceType := Pawns; pieceType <= Kings; pieceType++ {
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

	if b.Side == Black {
		hash ^= swapSide
	}

	return hash
}

// Incrementally update Zborist hash after a move.
func (b *Board) ZobristSimpleMove(move Move, piece int) {
	from, to := move.From(), move.To()
	b.Hash ^= pieceKeys[b.Side][piece][to]
	b.Hash ^= pieceKeys[b.Side][piece][from]
}

func (b *Board) ZobristCapture(move Move, piece int) {
	from, to := move.From(), move.To()
	capturedPiece := b.PieceAtSquare(to)
	b.Hash ^= pieceKeys[b.Side^1][capturedPiece][to]
	b.Hash ^= pieceKeys[b.Side][piece][to]
	b.Hash ^= pieceKeys[b.Side][piece][from]
}

func (b *Board) ZobristEPCapture(move Move) {
	from, to := move.From(), move.To()
	direction := Square(8)
	if b.Side == White {
		direction = -8
	}
	b.Hash ^= pieceKeys[b.Side^1][Pawns][to-direction]
	b.Hash ^= pieceKeys[b.Side][Pawns][to]
	b.Hash ^= pieceKeys[b.Side][Pawns][from]
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
		b.ZobristSimpleMove(WCastleKing, Kings)
		b.ZobristSimpleMove(WCastleKingRook, Rooks)
	case WOOO:
		b.ZobristSimpleMove(WCastleQueen, Kings)
		b.ZobristSimpleMove(WCastleQueenRook, Rooks)
	case BOO:
		b.ZobristSimpleMove(BCastleKing, Kings)
		b.ZobristSimpleMove(BCastleKingRook, Rooks)
	case BOOO:
		b.ZobristSimpleMove(BCastleQueen, Kings)
		b.ZobristSimpleMove(BCastleQueenRook, Rooks)
	}
}

// Update Zobrist hash when promoting a piece.
func (b *Board) ZobristPromotion(move Move) {
	to := move.To()

	// set destination with newly promoted piece
	b.Hash ^= pieceKeys[b.Side][move.Promotion()][to]
	b.Hash ^= pieceKeys[b.Side][Pawns][to]
}

// Update Zobrist hash with En Passant square.
func (b *Board) ZobristEnPassant(square Square) {
	if b.EnPassantTarget != -1 {
		b.Hash ^= enPassantKeys[square]
	}
}

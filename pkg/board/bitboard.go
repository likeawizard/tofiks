package board

import (
	"fmt"
	"math/bits"
	"strings"
)

// String returns a human-readable string representation of a bitboard.
func (bb BBoard) String() string {
	s := ""
	for r := 0; r < 8; r++ {
		s += fmt.Sprintf(" %d ", 8-r)
		for f := 0; f < 8; f++ {
			sq := r*8 + f
			s += fmt.Sprintf(" %d", bb.Get(sq))
		}
		s += "\n"
	}
	s += "\n    a b c d e f g h"
	s += fmt.Sprintf("\n\n Bitboard: %d", bb)
	return s
}

func (bb BBoard) Flip() BBoard {
	return BBoard(bits.ReverseBytes64(uint64(bb)))
}

// Get the bit at position.
func (bb *BBoard) Get(sq int) BBoard {
	return *bb >> sq & 1
}

// Set a bit to one at position.
func (bb *BBoard) Set(sq int) {
	*bb |= SquareBitboards[sq]
}

// Set a bit to zero at position.
func (bb *BBoard) Clear(sq int) {
	*bb &= ^SquareBitboards[sq]
}

// Return population count (number of 1's).
func (bb BBoard) Count() int {
	return bits.OnesCount64(uint64(bb))
}

// Get the position of the Least Significant.
func (bb BBoard) LS1B() int {
	return bits.TrailingZeros64(uint64(bb))
}

func (bb *BBoard) PopLS1B() int {
	ls1b := bits.TrailingZeros64(uint64(*bb))
	bb.Clear(ls1b)
	return ls1b
}

// Get bishop attack mask with blocker occupancy.
func GetBishopAttacks(sq int, occ BBoard) BBoard {
	occ &= BishopAttackMasks[sq]
	occ *= BishopMagics[sq]
	occ >>= 64 - BishopOccBitCount[sq]
	return BishopAttacks[sq][occ]
}

// Get Rook attack mask with blocker occupancy.
func GetRookAttacks(sq int, occ BBoard) BBoard {
	occ &= RookAttackMasks[sq]
	occ *= RookMagics[sq]
	occ >>= 64 - RookOccBitCount[sq]
	return RookAttacks[sq][occ]
}

// Get Queen attacks as a Bishop and Rook superposition.
func GetQueenAttacks(sq int, occ BBoard) BBoard {
	return GetBishopAttacks(sq, occ) | GetRookAttacks(sq, occ)
}

// Get pinned piece square and pin attack mask which are the only legal destination squares for the pinned piece.
// Basic assumptions:
// A piece can only be pinned by one attacker.
// An attacker that delivers check can not also pin a piece.
// A knight can never unpin itself. A pinned knight has no legal moves.
// A bishop can not unpin itself from rook attacks and vice versa.
func (b *Board) GetPinsBB(side int) map[int]BBoard {
	king := b.Pieces[side][Kings].LS1B()
	pins := make(map[int]BBoard)
	var directAttackers, xRayAttackers, attackMask, pinnedPieces BBoard
	var attackerSq int
	directAttackers = GetBishopAttacks(king, b.Occupancy[Both]) & (b.Pieces[side^1][Bishops] | b.Pieces[side^1][Queens])
	xRayAttackers = GetBishopAttacks(king, b.Occupancy[side^1]) & (b.Pieces[side^1][Bishops] | b.Pieces[side^1][Queens]) &^ directAttackers
	for xRayAttackers > 0 {
		attackerSq = xRayAttackers.PopLS1B()
		attackMask = (GetBishopAttacks(attackerSq, b.Pieces[side][Kings]) & GetBishopAttacks(king, SquareBitboards[attackerSq])) | SquareBitboards[attackerSq]&^b.Pieces[side][Kings]
		pinnedPieces = attackMask & b.Occupancy[side]
		if pinnedPieces > 0 && pinnedPieces.Count() == 1 {
			pins[pinnedPieces.LS1B()] = attackMask
		}
	}

	directAttackers = GetRookAttacks(king, b.Occupancy[Both]) & (b.Pieces[side^1][Rooks] | b.Pieces[side^1][Queens])
	xRayAttackers = GetRookAttacks(king, b.Occupancy[side^1]) & (b.Pieces[side^1][Rooks] | b.Pieces[side^1][Queens]) &^ directAttackers
	for xRayAttackers > 0 {
		attackerSq = xRayAttackers.PopLS1B()
		attackMask = (GetRookAttacks(attackerSq, b.Pieces[side][Kings]) & GetRookAttacks(king, SquareBitboards[attackerSq])) | SquareBitboards[attackerSq]&^b.Pieces[side][Kings]
		pinnedPieces = attackMask & b.Occupancy[side]
		if pinnedPieces > 0 && pinnedPieces.Count() == 1 {
			pins[pinnedPieces.LS1B()] = attackMask
		}
	}
	return pins
}

// Get checkers and check attack vectors and true if the check is a double check. A zero bitboard indicates no check.
// Slider piece checks return a bitboard containing squares that are legal destinations which either capture the checker or block its attack.
// A knight checker returns only the position of the knight to be captured as blocking is impossible unlike sliding pieces.
// In case of a double check only the king can move and the resulting bitboard can not be used for determining the legality of other piece moves.
func (b *Board) GetChecksBB(side int) (BBoard, bool) {
	var numChecks int
	var checks, attacker BBoard
	var pawnCheck bool
	king := b.Pieces[side][Kings].LS1B()

	attacker = PawnAttacks[side][king] & b.Pieces[side^1][Pawns]
	if attacker != 0 {
		pawnCheck = true
		checks |= attacker
		numChecks++
	}

	attacker = GetRookAttacks(king, b.Occupancy[Both]) & (b.Pieces[side^1][Rooks] | b.Pieces[side^1][Queens])
	if attacker != 0 {
		checks |= (GetRookAttacks(attacker.LS1B(), b.Pieces[side][Kings]) & GetRookAttacks(king, attacker)) | attacker&^b.Pieces[side][Kings]
		numChecks += attacker.Count()
	}

	// A pawn can check by moving forward or capturing. Only a capture move that clears a file for a rook attack can create a double check. So only check Knight and Bishop checks if no pawn check is present
	if !pawnCheck {
		attacker = KnightAttacks[king] & b.Pieces[side^1][Knights]
		if attacker != 0 {
			checks |= attacker
			numChecks++
		}

		attacker = GetBishopAttacks(king, b.Occupancy[Both]) & (b.Pieces[side^1][Bishops] | b.Pieces[side^1][Queens])
		if attacker != 0 {
			checks |= (GetBishopAttacks(attacker.LS1B(), b.Pieces[side][Kings]) & GetBishopAttacks(king, attacker)) | attacker&^b.Pieces[side][Kings]
			numChecks++
		}
	}

	return checks, numChecks > 1
}

// Determine if a square is attacked by the opposing side.
func (b *Board) IsAttacked(sq int, side int8, occ BBoard) bool {
	return PawnAttacks[side][sq]&b.Pieces[side^1][Pawns] != 0 ||
		KnightAttacks[sq]&b.Pieces[side^1][Knights] != 0 ||
		KingAttacks[sq]&b.Pieces[side^1][Kings] != 0 ||
		GetBishopAttacks(sq, occ)&(b.Pieces[side^1][Bishops]|b.Pieces[side^1][Queens]) != 0 ||
		GetRookAttacks(sq, occ)&(b.Pieces[side^1][Rooks]|b.Pieces[side^1][Queens]) != 0
}

// Determine if the king for the given side is in check.
func (b *Board) IsChecked(side int8) bool {
	return b.IsAttacked(b.Pieces[side][Kings].LS1B(), side, b.Occupancy[Both])
}

// IsPseudoLegal performs a fast sanity check that the move's piece exists on the
// from-square for the side to move. Catches most type-2 hash collisions in TT probes.
func (b *Board) IsPseudoLegal(move Move) bool {
	return SquareBitboards[int(move.From())]&b.Pieces[b.Side][move.Piece()] != 0
}

// Get a bitboard of all the squares attacked by the opposition.
func (b *Board) AttackedSquares(side int8, mask, occ BBoard) BBoard {
	attacked := BBoard(0)
	m := mask
	var sq int
	for m > 0 {
		sq = m.PopLS1B()
		if b.IsAttacked(sq, side, occ) {
			attacked |= SquareBitboards[sq]
		}
	}

	return attacked
}

// Generate a function to return the board state the it's current state.
func (b *Board) GetUnmake() func() {
	var (
		hash            = b.Hash
		pawnHash        = b.PawnHash
		pieces          = b.Pieces
		occupancy       = b.Occupancy
		side            = b.Side
		inCheck         = b.InCheck
		castlingRights  = b.CastlingRights
		enPassantTarget = b.EnPassantTarget
		halfMoveCounter = b.HalfMoveCounter
		fullMoveCounter = b.FullMoveCounter
	)

	return func() {
		b.Hash = hash
		b.PawnHash = pawnHash
		b.Pieces = pieces
		b.Occupancy = occupancy
		b.Side = side
		b.InCheck = inCheck
		b.CastlingRights = castlingRights
		b.EnPassantTarget = enPassantTarget
		b.HalfMoveCounter = halfMoveCounter
		b.FullMoveCounter = fullMoveCounter
	}
}

// Make a legal move in position and update board state - castling rights, en passant, move count, side to move etc. Returns a function to take back the move made.
func (b *Board) MakeMove(move Move) func() {
	umove := b.GetUnmake()
	isCapture := move.IsCapture()
	piece := int(move.Piece())
	if isCapture || piece == Pawns {
		b.HalfMoveCounter = 0
	} else {
		b.HalfMoveCounter++
	}

	if b.EnPassantTarget > 0 {
		b.ZobristEnPassant(b.EnPassantTarget)
	}

	bitboard := &b.Pieces[b.Side][piece]

	from, to := move.From(), move.To()

	switch {
	case move.IsEnPassant():
		b.ZobristEPCapture(move)
		b.EnPassantTarget = -1
		direction := 8
		if b.Side == White {
			direction = -8
		}
		capSq := int(to) - direction
		b.PawnHash ^= pieceKeys[b.Side][Pawns][from] ^ pieceKeys[b.Side][Pawns][to] ^ pieceKeys[b.Side^1][Pawns][capSq]
		b.RemoveCaptured(capSq)
	case isCapture:
		b.EnPassantTarget = -1
		capturedPiece := b.PieceAtSquare(to)
		b.ZobristCapture(move, piece)
		if capturedPiece == Pawns {
			b.PawnHash ^= pieceKeys[b.Side^1][Pawns][to]
		}
		if piece == Pawns {
			b.PawnHash ^= pieceKeys[b.Side][Pawns][from] ^ pieceKeys[b.Side][Pawns][to]
		}
		b.RemoveCaptured(int(to))
	case move.IsCastling():
		b.EnPassantTarget = -1
		b.ZobristSimpleMove(move, piece)
		b.CompleteCastling(move)
	case move.IsDouble():
		b.ZobristSimpleMove(move, piece)
		b.PawnHash ^= pieceKeys[b.Side][Pawns][from] ^ pieceKeys[b.Side][Pawns][to]
		b.EnPassantTarget = (to + from) / 2
		b.ZobristEnPassant(b.EnPassantTarget)
	default:
		b.EnPassantTarget = -1
		b.ZobristSimpleMove(move, piece)
		if piece == Pawns {
			b.PawnHash ^= pieceKeys[b.Side][Pawns][from] ^ pieceKeys[b.Side][Pawns][to]
		}
	}
	bitboard.Set(int(move.To()))
	bitboard.Clear(int(move.From()))

	if move.Promotion() != 0 {
		b.PawnHash ^= pieceKeys[b.Side][Pawns][to]
	}
	b.Promote(move)

	for side := White; side <= Black; side++ {
		b.Occupancy[side] = b.Pieces[side][Kings]
		for piece := Pawns; piece < Kings; piece++ {
			b.Occupancy[side] |= b.Pieces[side][piece]
		}
	}
	b.Occupancy[Both] = b.Occupancy[White] | b.Occupancy[Black]

	b.updateCastlingRights(move)
	if b.Side == Black {
		b.FullMoveCounter++
	}

	b.ZobristSideToMove()
	b.Side ^= 1
	b.InCheck = b.IsChecked(b.Side)
	return umove
}

// Determine the game phase as a sliding factor between opening and endgame
// https://www.chessprogramming.org/Tapered_Eval#Implementation_example
func (b *Board) GetGamePhase() int {
	phase := 24 -
		b.Pieces[White][Bishops].Count() - b.Pieces[Black][Bishops].Count() -
		b.Pieces[White][Knights].Count() - b.Pieces[Black][Knights].Count() -
		2*(b.Pieces[White][Rooks].Count()+b.Pieces[Black][Rooks].Count()) -
		4*(b.Pieces[White][Queens].Count()+b.Pieces[Black][Queens].Count())

	return (phase * 268) / 24
}

func (b *Board) MakeNullMove() func() {
	type undoNull struct {
		inCheck bool
		ep      Square
	}
	undo := undoNull{
		ep: b.EnPassantTarget,
	}
	b.ZobristEnPassant(b.EnPassantTarget)
	b.EnPassantTarget = -1
	b.HalfMoveCounter++

	b.ZobristSideToMove()
	b.Side ^= 1
	b.InCheck = b.IsChecked(b.Side)
	return func() {
		b.HalfMoveCounter--
		b.EnPassantTarget = undo.ep
		b.ZobristEnPassant(undo.ep)
		b.InCheck = undo.inCheck
		b.ZobristSideToMove()
		b.Side ^= 1
	}
}

// Attempt to play a UCI move in position. Returns unmake closure and ok.
func (b *Board) MoveUCI(uciMove string) (func(), bool) {
	all := b.PseudoMoveGen()
	for _, move := range all {
		if uciMove == move.String() {
			umove := b.MakeMove(move)
			if b.IsChecked(b.Side ^ 1) {
				umove()
				return nil, false
			}
			return umove, true
		}
	}
	return nil, false
}

// Play out a line of UCI moves in succession. Returns success.
func (b *Board) PlayMovesUCI(uciMoves string) bool {
	moveSlice := strings.Fields(uciMoves)

	for _, uciMove := range moveSlice {
		_, ok := b.MoveUCI(uciMove)
		if !ok {
			return false
		}
	}

	return true
}

// Remove a piece captured by a move from the opposing bitboard.
func (b *Board) RemoveCaptured(sq int) {
	b.Occupancy[b.Side^1].Clear(sq)
	for piece := Pawns; piece <= Kings; piece++ {
		b.Pieces[b.Side^1][piece] &= b.Occupancy[b.Side^1]
	}
}

// Make the complimentary rook move when castling.
func (b *Board) CompleteCastling(move Move) {
	bitboard := &b.Pieces[b.Side][Rooks]
	var rookMove Move
	switch move {
	case WCastleKing:
		rookMove = WCastleKingRook
	case WCastleQueen:
		rookMove = WCastleQueenRook
	case BCastleKing:
		rookMove = BCastleKingRook
	case BCastleQueen:
		rookMove = BCastleQueenRook
	}
	b.ZobristSimpleMove(rookMove, Rooks)
	bitboard.Set(int(rookMove.To()))
	bitboard.Clear(int(rookMove.From()))
}

// Get the piece at square as a collection of values: found, color, piece.
func (b *Board) PieceAtSquare(sq Square) int {
	for color := White; color <= Black; color++ {
		for pieceType := Pawns; pieceType <= Kings; pieceType++ {
			if SquareBitboards[sq]&b.Pieces[color][pieceType] != 0 {
				return pieceType
			}
		}
	}

	return NoPiece
}

// Replace a pawn on the 8th/1st rank with the promotion piece.
func (b *Board) Promote(move Move) {
	promotion := move.Promotion()
	if promotion == 0 {
		return
	}

	var pawnBitBoard, promotionBitBoard *BBoard
	pawnBitBoard = &b.Pieces[b.Side][Pawns]
	promotionBitBoard = &b.Pieces[b.Side][promotion]

	pawnBitBoard.Clear(int(move.To()))
	promotionBitBoard.Set(int(move.To()))
	b.ZobristPromotion(move)
}

// Determine if the game only consists of pawns and kings.
func (b *Board) IsPawnOnly() bool {
	return b.Pieces[White][Pawns]|b.Pieces[White][Kings]|b.Pieces[Black][Pawns]|b.Pieces[Black][Kings] == b.Occupancy[Both]
}

// Determine if there is a draw by insufficient material
// This determines theoretical possibility of mate. Not KvKNN, which still can be achieved as a 'help mate'.
func (b *Board) InsufficientMaterial() bool {
	isLight := func(s int) bool {
		return ((s/8)+(s%8))%2 == 0
	}
	// If any pawn or major piece on the board can't have insufficient material
	// No game with 3 or more minors is a strict draw.
	if b.Pieces[White][Pawns] != 0 || b.Pieces[Black][Pawns] != 0 ||
		b.Pieces[White][Queens] != 0 || b.Pieces[Black][Queens] != 0 ||
		b.Pieces[White][Rooks] != 0 || b.Pieces[Black][Rooks] != 0 ||
		b.Occupancy[Both].Count() > 4 {
		return false
	}

	// We disqualified all obvious sufficient material cases above and are left with games that have at most 2 minors
	wN, wB := b.Pieces[White][Knights].Count(), b.Pieces[White][Bishops].Count()
	bN, bB := b.Pieces[Black][Knights].Count(), b.Pieces[Black][Bishops].Count()
	wM, bM := wN+wB, bN+bB

	// Check if only one (or zero) minor on the board. KvKM King v King plus one minor = draw
	if wM+bM <= 1 {
		return true
	}

	// There must be two minors. If either side has two minros - not a draw KvKNN, KvKBB, KvKBN
	if wM > 1 || bM > 1 {
		return false
	}

	if wB == 1 && bB == 1 {
		return isLight(b.Pieces[White][Bishops].LS1B()) == isLight(b.Pieces[Black][Bishops].LS1B())
	}

	return false
}

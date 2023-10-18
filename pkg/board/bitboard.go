package board

import (
	"fmt"
	"math/bits"
	"strings"
)

// Get a human readable string represantiation of a bitboard.
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
	king := b.Pieces[side][KINGS].LS1B()
	pins := make(map[int]BBoard)
	var directAttackers, xRayAttackers, attackMask, pinnedPieces BBoard
	var attackerSq int
	directAttackers = GetBishopAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][BISHOPS] | b.Pieces[side^1][QUEENS])
	xRayAttackers = GetBishopAttacks(king, b.Occupancy[side^1]) & (b.Pieces[side^1][BISHOPS] | b.Pieces[side^1][QUEENS]) &^ directAttackers
	for xRayAttackers > 0 {
		attackerSq = xRayAttackers.PopLS1B()
		attackMask = (GetBishopAttacks(attackerSq, b.Pieces[side][KINGS]) & GetBishopAttacks(king, SquareBitboards[attackerSq])) | SquareBitboards[attackerSq]&^b.Pieces[side][KINGS]
		pinnedPieces = attackMask & b.Occupancy[side]
		if pinnedPieces > 0 && pinnedPieces.Count() == 1 {
			pins[pinnedPieces.LS1B()] = attackMask
		}
	}

	directAttackers = GetRookAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][ROOKS] | b.Pieces[side^1][QUEENS])
	xRayAttackers = GetRookAttacks(king, b.Occupancy[side^1]) & (b.Pieces[side^1][ROOKS] | b.Pieces[side^1][QUEENS]) &^ directAttackers
	for xRayAttackers > 0 {
		attackerSq = xRayAttackers.PopLS1B()
		attackMask = (GetRookAttacks(attackerSq, b.Pieces[side][KINGS]) & GetRookAttacks(king, SquareBitboards[attackerSq])) | SquareBitboards[attackerSq]&^b.Pieces[side][KINGS]
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
	king := b.Pieces[side][KINGS].LS1B()

	attacker = PawnAttacks[side][king] & b.Pieces[side^1][PAWNS]
	if attacker != 0 {
		pawnCheck = true
		checks |= attacker
		numChecks++
	}

	attacker = GetRookAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][ROOKS] | b.Pieces[side^1][QUEENS])
	if attacker != 0 {
		checks |= (GetRookAttacks(attacker.LS1B(), b.Pieces[side][KINGS]) & GetRookAttacks(king, attacker)) | attacker&^b.Pieces[side][KINGS]
		numChecks += attacker.Count()
	}

	// A pawn can check by moving forward or capturing. Only a capture move that clears a file for a rook attack can create a double check. So only check Knight and Bishop checks if no pawn check is present
	if !pawnCheck {
		attacker = KnightAttacks[king] & b.Pieces[side^1][KNIGHTS]
		if attacker != 0 {
			checks |= attacker
			numChecks++
		}

		attacker = GetBishopAttacks(king, b.Occupancy[BOTH]) & (b.Pieces[side^1][BISHOPS] | b.Pieces[side^1][QUEENS])
		if attacker != 0 {
			checks |= (GetBishopAttacks(attacker.LS1B(), b.Pieces[side][KINGS]) & GetBishopAttacks(king, attacker)) | attacker&^b.Pieces[side][KINGS]
			numChecks++
		}
	}

	return checks, numChecks > 1
}

// Determine if a square is attacked by the opposing side.
func (b *Board) IsAttacked(sq, side int, occ BBoard) bool {
	var isAttacked bool

	if PawnAttacks[side][sq]&b.Pieces[side^1][PAWNS] != 0 {
		return true
	}

	if KnightAttacks[sq]&b.Pieces[side^1][KNIGHTS] != 0 {
		return true
	}

	if KingAttacks[sq]&b.Pieces[side^1][KINGS] != 0 {
		return true
	}

	if GetBishopAttacks(sq, occ)&(b.Pieces[side^1][BISHOPS]|b.Pieces[side^1][QUEENS]) != 0 {
		return true
	}

	if GetRookAttacks(sq, occ)&(b.Pieces[side^1][ROOKS]|b.Pieces[side^1][QUEENS]) != 0 {
		return true
	}

	return isAttacked
}

// Determine if the king for the given side is in check.
func (b *Board) IsChecked(side int) bool {
	king := b.Pieces[side][KINGS].LS1B()

	return b.IsAttacked(king, side, b.Occupancy[BOTH])
}

// Get a bitboard of all the squares attacked by the opposition.
func (b *Board) AttackedSquares(side int, occ BBoard) BBoard {
	attacked := BBoard(0)

	for sq := 0; sq < 64; sq++ {
		if b.IsAttacked(sq, side, occ) {
			attacked |= SquareBitboards[sq]
		}
	}

	return attacked
}

// Generate a function to return the board state the it's current state.
func (b *Board) GetUnmake() func() {
	cp := b.Copy()
	return func() {
		b.Hash = cp.Hash
		b.Pieces = cp.Pieces
		b.Occupancy = cp.Occupancy
		b.Side = cp.Side
		b.Phase = cp.Phase
		b.InCheck = cp.InCheck
		b.CastlingRights = cp.CastlingRights
		b.EnPassantTarget = cp.EnPassantTarget
		b.HalfMoveCounter = cp.HalfMoveCounter
		b.FullMoveCounter = cp.FullMoveCounter
	}
}

// Make a legal move in position and update board state - castling rights, en passant, move count, side to move etc. Returns a function to take back the move made.
func (b *Board) MakeMove(move Move) func() {
	umove := b.GetUnmake()
	isCapture := move.IsCapture()
	piece := int(move.Piece())
	if isCapture || piece == PAWNS {
		b.HalfMoveCounter = 0
	} else {
		b.HalfMoveCounter++
	}

	if b.EnPassantTarget > 0 {
		b.ZobristEnPassant(b.EnPassantTarget)
	}

	bitboard := b.GetBitBoard(piece)

	switch {
	case move.IsEnPassant():
		b.ZobristEPCapture(move)
		b.EnPassantTarget = -1
		direction := 8
		if b.Side == WHITE {
			direction = -8
		}
		b.RemoveCaptured(int(move.To()) - direction)
	case isCapture:
		b.EnPassantTarget = -1
		b.ZobristCapture(move, piece)
		b.RemoveCaptured(int(move.To()))
	case move.IsCastling():
		b.EnPassantTarget = -1
		b.ZobristSimpleMove(move, piece)
		b.CompleteCastling(move)
	case move.IsDouble():
		b.ZobristSimpleMove(move, piece)
		b.EnPassantTarget = (move.To() + move.From()) / 2
		b.ZobristEnPassant(b.EnPassantTarget)
	default:
		b.EnPassantTarget = -1
		b.ZobristSimpleMove(move, piece)
	}
	bitboard.Set(int(move.To()))
	bitboard.Clear(int(move.From()))

	b.Promote(move)

	for side := WHITE; side <= BLACK; side++ {
		b.Occupancy[side] = b.Pieces[side][KINGS]
		for piece := PAWNS; piece < KINGS; piece++ {
			b.Occupancy[side] |= b.Pieces[side][piece]
		}
	}
	b.Occupancy[BOTH] = b.Occupancy[WHITE] | b.Occupancy[BLACK]

	b.updateCastlingRights(move)
	if b.Side == BLACK {
		b.FullMoveCounter++
	}

	b.Phase = b.GetGamePhase()
	b.ZobristSideToMove()
	b.Side ^= 1
	b.InCheck = b.IsChecked(b.Side)
	return umove
}

// Determine the game phase as a sliding factor between opening and endgame
// https://www.chessprogramming.org/Tapered_Eval#Implementation_example
func (b *Board) GetGamePhase() int {
	phase := 24

	for color := WHITE; color <= BLACK; color++ {
		for pieceType := PAWNS; pieceType <= KINGS; pieceType++ {
			switch pieceType {
			case BISHOPS, KINGS:
				phase -= b.Pieces[color][pieceType].Count()
			case ROOKS:
				phase -= 2 * b.Pieces[color][pieceType].Count()
			case QUEENS:
				phase -= 4 * b.Pieces[color][pieceType].Count()
			}
		}
	}

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
		b.ZobristEnPassant(undo.ep)
		b.EnPassantTarget = undo.ep
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

// Return a pointer to the bitboard of the piece moved.
func (b *Board) GetBitBoard(piece int) *BBoard {
	return &b.Pieces[b.Side][piece]
}

// Remove a piece captured by a move from the opposing bitboard.
func (b *Board) RemoveCaptured(sq int) {
	b.Occupancy[b.Side^1].Clear(sq)
	for piece := PAWNS; piece <= KINGS; piece++ {
		b.Pieces[b.Side^1][piece] &= b.Occupancy[b.Side^1]
	}
}

// Make the complimentary rook move when castling.
func (b *Board) CompleteCastling(move Move) {
	bitboard := &b.Pieces[b.Side][ROOKS]
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
	b.ZobristSimpleMove(rookMove, ROOKS)
	bitboard.Set(int(rookMove.To()))
	bitboard.Clear(int(rookMove.From()))
}

// Get the piece at square as a collection of values: found, color, piece.
func (b *Board) PieceAtSquare(sq Square) int {
	for color := WHITE; color <= BLACK; color++ {
		for pieceType := PAWNS; pieceType <= KINGS; pieceType++ {
			if SquareBitboards[sq]&b.Pieces[color][pieceType] != 0 {
				return pieceType
			}
		}
	}

	return 6
}

// Replace a pawn on the 8th/1st rank with the promotion piece.
func (b *Board) Promote(move Move) {
	promotion := move.Promotion()
	if promotion == 0 {
		return
	}

	var pawnBitBoard, promotionBitBoard *BBoard
	pawnBitBoard = &b.Pieces[b.Side][PAWNS]
	promotionBitBoard = &b.Pieces[b.Side][promotion]

	pawnBitBoard.Clear(int(move.To()))
	promotionBitBoard.Set(int(move.To()))
	b.ZobristPromotion(move)
}

// Determine if the game only consists of pawns and kings.
func (b *Board) IsPawnOnly() bool {
	return b.Pieces[WHITE][PAWNS]|b.Pieces[WHITE][KINGS]|b.Pieces[BLACK][PAWNS]|b.Pieces[BLACK][KINGS] == b.Occupancy[BOTH]
}

// Determine if there is a draw by insufficient material
// This determines theoretical possibility of mate. Not KvKNN, which still can be achieved as a 'help mate'.
func (b *Board) InsufficentMaterial() bool {
	isLight := func(s int) bool {
		return ((s/8)+(s%8))%2 == 0
	}
	// If any pawn or major piece on the board can't have insufficient material
	// No game with 3 or more minors is a strict draw.
	if b.Pieces[WHITE][PAWNS] != 0 || b.Pieces[BLACK][PAWNS] != 0 ||
		b.Pieces[WHITE][QUEENS] != 0 || b.Pieces[BLACK][QUEENS] != 0 ||
		b.Pieces[WHITE][ROOKS] != 0 || b.Pieces[BLACK][ROOKS] != 0 ||
		b.Occupancy[BOTH].Count() > 4 {
		return false
	}

	// We disqualified all obvious sufficient material cases above and are left with games that have at most 2 minors
	wN, wB := b.Pieces[WHITE][KNIGHTS].Count(), b.Pieces[WHITE][BISHOPS].Count()
	bN, bB := b.Pieces[BLACK][KNIGHTS].Count(), b.Pieces[BLACK][BISHOPS].Count()
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
		return isLight(b.Pieces[WHITE][BISHOPS].LS1B()) == isLight(b.Pieces[BLACK][BISHOPS].LS1B())
	}

	return false
}

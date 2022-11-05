package board

// Contains bitboard related pre-computed values and initialiaztion code

import (
	"fmt"
	"math/rand"
)

var SquareBitboards [64]BBoard
var PawnAttacks [2][64]BBoard
var KnightAttacks [64]BBoard
var KingAttacks [64]BBoard
var KingSafetyMask [2][64]BBoard
var BishopOccBitCount [64]int
var RookOccBitCount [64]int
var BishopMagics [64]BBoard
var RookMagics [64]BBoard
var BishopAttackMasks [64]BBoard
var RookAttackMasks [64]BBoard
var BishopAttacks [64][4096]BBoard
var RookAttacks [64][4096]BBoard

var PassedPawns [2][64]BBoard
var IsolatedPawns [64]BBoard
var DoubledPawns [64]BBoard

const MajorDiag BBoard = 9314046665258451585
const MinorDiag BBoard = 4946458877011600706

func init() {
	InitSquares()
	InitOccBitCounts()
	InitPawnAttacks()
	InitKnightAttacks()
	InitKingAttacks()
	InitKingSafetyMasks()
	InitPawnStrucutreMasks()
	InitMagics()
	InitSliders()
}

func InitSquares() {
	for i := 0; i < 64; i++ {
		SquareBitboards[i] = 1 << i
	}
}

// Initialize pawn attack lookup table
func InitPawnAttacks() {
	pawnAttack := func(sq int, isWhite bool) BBoard {
		var piece, attacks BBoard

		piece.Set(sq)
		if isWhite {
			if piece&HFile == 0 {
				attacks |= piece >> 7
			}
			if piece&AFile == 0 {
				attacks |= piece >> 9
			}
		} else {
			if piece&AFile == 0 {
				attacks |= piece << 7
			}
			if piece&HFile == 0 {
				attacks |= piece << 9
			}
		}

		return attacks
	}
	for sq := 0; sq < 64; sq++ {
		PawnAttacks[WHITE][sq] = pawnAttack(sq, true)
		PawnAttacks[BLACK][sq] = pawnAttack(sq, false)
	}
}

func InitPawnStrucutreMasks() {
	isloatedMask := func(sq int) BBoard {
		file := sq % 8
		switch file {
		case 0:
			return BFile
		case 1:
			return AFile | CFile
		case 2:
			return DFile | BFile
		case 3:
			return EFile | CFile
		case 4:
			return DFile | FFile
		case 5:
			return EFile | GFile
		case 6:
			return FFile | HFile
		default:
			return GFile
		}
	}

	doubledMask := func(sq int) BBoard {
		file := sq % 8
		var fileMask BBoard
		switch file {
		case 0:
			fileMask = AFile
		case 1:
			fileMask = BFile
		case 2:
			fileMask = CFile
		case 3:
			fileMask = DFile
		case 4:
			fileMask = EFile
		case 5:
			fileMask = FFile
		case 6:
			fileMask = GFile
		case 7:
			fileMask = HFile
		}
		return fileMask &^ (1 << sq)
	}

	passedMask := func(sq, color int) BBoard {
		rank := sq / 8
		mask := IsolatedPawns[sq] | DoubledPawns[sq]
		behindMask := func(rank, color int) BBoard {
			if color == WHITE {
				switch rank {
				case 0:
					return ^BBoard(0)
				case 1:
					return Rank1 | Rank2 | Rank3 | Rank4 | Rank5 | Rank6 | Rank7
				case 2:
					return Rank1 | Rank2 | Rank3 | Rank4 | Rank5 | Rank6
				case 3:
					return Rank1 | Rank2 | Rank3 | Rank4 | Rank5
				case 4:
					return Rank1 | Rank2 | Rank3 | Rank4
				case 5:
					return Rank1 | Rank2 | Rank3
				case 6:
					return Rank1 | Rank2
				default:
					return Rank1
				}
			} else {
				switch rank {
				case 0:
					return Rank8
				case 1:
					return Rank8 | Rank7
				case 2:
					return Rank8 | Rank7 | Rank6
				case 3:
					return Rank8 | Rank7 | Rank6 | Rank5
				case 4:
					return Rank8 | Rank7 | Rank6 | Rank5 | Rank4
				case 5:
					return Rank8 | Rank7 | Rank6 | Rank5 | Rank4 | Rank3
				case 6:
					return Rank8 | Rank7 | Rank6 | Rank5 | Rank4 | Rank3 | Rank2
				default:
					return ^BBoard(0)
				}
			}
		}

		return mask & ^behindMask(rank, color)
	}

	for sq := 0; sq < 64; sq++ {
		IsolatedPawns[sq] = isloatedMask(sq)
		DoubledPawns[sq] = doubledMask(sq)
		PassedPawns[WHITE][sq] = passedMask(sq, WHITE)
		PassedPawns[BLACK][sq] = passedMask(sq, BLACK)
	}
}

// Initialize Knight attack lookup table
func InitKnightAttacks() {
	knightAttack := func(sq int) BBoard {
		var piece, attacks BBoard
		piece.Set(sq)

		if piece&HFile == 0 {
			attacks |= piece >> 15
			attacks |= piece << 17
		}

		if piece&AFile == 0 {
			attacks |= piece >> 17
			attacks |= piece << 15
		}

		if piece&AFile == 0 && piece&BFile == 0 {
			attacks |= piece >> 10
			attacks |= piece << 6
		}

		if piece&HFile == 0 && piece&GFile == 0 {
			attacks |= piece >> 6
			attacks |= piece << 10
		}

		return attacks
	}
	for sq := 0; sq < 64; sq++ {
		KnightAttacks[sq] = knightAttack(sq)
	}
}

// Initialize King move lookup table
func InitKingAttacks() {
	kingAttack := func(sq int) BBoard {
		var piece, attacks BBoard
		piece.Set(sq)

		attacks |= piece >> 8
		attacks |= piece << 8

		if piece&HFile == 0 {
			attacks |= piece >> 7
			attacks |= piece << 1
			attacks |= piece << 9
		}

		if piece&AFile == 0 {
			attacks |= piece >> 9
			attacks |= piece >> 1
			attacks |= piece << 7
		}

		return attacks
	}
	for sq := 0; sq < 64; sq++ {
		KingAttacks[sq] = kingAttack(sq)
	}
}

// King safety masks ar similar to KingAttacks but do not cover squares behind the king. Only pieces in front of the kind attribute to safety
func InitKingSafetyMasks() {

	safetyMask := func(sq int, side int) BBoard {
		var piece, attacks BBoard
		piece.Set(sq)

		if piece&HFile == 0 {
			if side == WHITE {
				attacks |= piece >> 8
				attacks |= piece << 1
				attacks |= piece >> 7
			} else {
				attacks |= piece << 8
				attacks |= piece << 1
				attacks |= piece << 9
			}
		}

		if piece&AFile == 0 {
			if side == WHITE {
				attacks |= piece >> 8
				attacks |= piece >> 9
				attacks |= piece >> 1
			} else {
				attacks |= piece << 8
				attacks |= piece >> 1
				attacks |= piece << 7
			}
		}

		return attacks
	}
	for color := WHITE; color <= BLACK; color++ {
		for sq := 0; sq < 64; sq++ {
			KingSafetyMask[color][sq] = safetyMask(sq, color)
		}
	}
}

// Generate bishop relevant occupancy look table
// Relevant occupancy are the squares where a potential blocking piece can cut a sliding piece from moving beyond it
// Relevant occupancy does not include border squares as a piece on the edge of the board can not block it moving past it
func BishopRelOcc(sq int) BBoard {
	var piece, attacks BBoard
	piece.Set(sq)
	f, r := sq%8, sq/8
	for i := 1; f+i < 7 && r+i < 7; i++ {
		attacks.Set((r+i)*8 + f + i)
	}
	for i := 1; f+i < 7 && r-i > 0; i++ {
		attacks.Set((r-i)*8 + f + i)
	}
	for i := 1; f-i > 0 && r+i < 7; i++ {
		attacks.Set((r+i)*8 + f - i)
	}
	for i := 1; f-i > 0 && r-i > 0; i++ {
		attacks.Set((r-i)*8 + f - i)
	}

	return attacks
}

// Generate rook relevant occupancy. See Bishop equivalent for details.
func RookRelOcc(sq int) BBoard {
	var piece, attacks BBoard
	piece.Set(sq)
	f, r := sq%8, sq/8
	for i := 1; f+i < 7; i++ {
		attacks.Set(r*8 + f + i)
	}
	for i := 1; r-i > 0; i++ {
		attacks.Set((r-i)*8 + f)
	}
	for i := 1; f-i > 0; i++ {
		attacks.Set(r*8 + f - i)
	}
	for i := 1; r+i < 7; i++ {
		attacks.Set((r+i)*8 + f)
	}

	return attacks
}

// Initialize Magic numbers to allow sliding piece lookup in conjunction with blocking occupancy
func InitMagics() {
	for sq := 0; sq < 64; sq++ {
		BishopMagics[sq] = FindMagicNumber(sq, BishopOccBitCount[sq], true)
		RookMagics[sq] = FindMagicNumber(sq, RookOccBitCount[sq], false)
	}
}

// Initialize sliding piece lookup tables with magic numbers
func InitSliders() {
	for sq := 0; sq < 64; sq++ {
		BishopAttackMasks[sq] = BishopRelOcc(sq)
		RookAttackMasks[sq] = RookRelOcc(sq)

		attackRook := RookAttackMasks[sq]
		attackBishop := BishopAttackMasks[sq]

		occRook := attackRook.Count()
		occBishop := attackBishop.Count()
		occIdxR := 1 << occRook
		occIdxB := 1 << occBishop

		for i := 0; i < occIdxB; i++ {
			occ := Occupancy(i, occBishop, attackBishop)
			mIdx := (occ * BishopMagics[sq]) >> (64 - occBishop)
			BishopAttacks[sq][mIdx] = BishopAttacksWithBlocker(sq, occ)
		}

		for i := 0; i < occIdxR; i++ {
			occ := Occupancy(i, occRook, attackRook)
			mIdx := (occ * RookMagics[sq]) >> (64 - occRook)
			RookAttacks[sq][mIdx] = RookAttacksWithBlocker(sq, occ)
		}
	}
}

// Generate occupancy bitboards for a given relevant occupancy bitboard.
func Occupancy(index, count int, attack BBoard) BBoard {
	occ := BBoard(0)

	for i := 0; i < count; i++ {
		sq := attack.LS1B()
		attack.Clear(sq)

		if index&(1<<i) != 0 {
			occ.Set(sq)
		}
	}

	return occ
}

// Generate Bishop sliding attacks with a blocker bitboard on the fly. Only used for finding and initializing magic numbers. Too slow to use for movegen
func BishopAttacksWithBlocker(sq int, blocker BBoard) BBoard {
	var piece, attacks BBoard
	piece.Set(sq)
	f, r := sq%8, sq/8
	var target int
	for i := 1; f+i < 8 && r+i < 8; i++ {
		target = (r+i)*8 + f + i
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}
	for i := 1; f+i < 8 && r-i >= 0; i++ {
		target = (r-i)*8 + f + i
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}
	for i := 1; f-i >= 0 && r+i < 8; i++ {
		target = (r+i)*8 + f - i
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}
	for i := 1; f-i >= 0 && r-i >= 0; i++ {
		target = (r-i)*8 + f - i
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}

	return attacks
}

// Generate Rook sliding attacks with a blocker bitboard on the fly. Only used for finding and initializing magic numbers. Too slow to use for movegen
func RookAttacksWithBlocker(sq int, blocker BBoard) BBoard {
	var piece, attacks BBoard
	piece.Set(sq)
	f, r := sq%8, sq/8
	var target int
	for i := 1; f+i < 8; i++ {
		target = r*8 + f + i
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}
	for i := 1; r-i >= 0; i++ {
		target = (r-i)*8 + f
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}
	for i := 1; f-i >= 0; i++ {
		target = r*8 + f - i
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}
	for i := 1; r+i < 8; i++ {
		target = (r+i)*8 + f
		attacks.Set(target)
		if blocker&SquareBitboards[target] != 0 {
			break
		}
	}

	return attacks
}

// Calculate and initialize the number of relevant occupancies for Bishops and Rooks for all squares
func InitOccBitCounts() {
	for sq := 0; sq < 64; sq++ {
		BishopOccBitCount[sq] = BishopRelOcc(sq).Count()
		RookOccBitCount[sq] = RookRelOcc(sq).Count()
	}
}

// Generate a random bitboard with few non-zero bits for Magic Number candidates
func GetMagicNumber() BBoard {
	return BBoard(rand.Uint64() & rand.Uint64() & rand.Uint64())
}

// Find magic numbers with a brute force approach for a square
func FindMagicNumber(sq, bitCount int, isBishop bool) BBoard {
	occ := [4096]BBoard{}
	attacks := [4096]BBoard{}
	usedAttacks := [4096]BBoard{}

	var attack BBoard
	if isBishop {
		attack = BishopRelOcc(sq)
	} else {
		attack = RookRelOcc(sq)
	}

	occIdx := 1 << bitCount

	for i := 0; i < occIdx; i++ {
		occ[i] = Occupancy(i, bitCount, attack)
		if isBishop {
			attacks[i] = BishopAttacksWithBlocker(sq, occ[i])
		} else {
			attacks[i] = RookAttacksWithBlocker(sq, occ[i])
		}
	}

	for randC := 0; randC < 1<<32; randC++ {
		magicNum := GetMagicNumber()

		if ((attack * magicNum) & 0xFF00000000000000).Count() < 6 {
			continue
		}

		usedAttacks = [4096]BBoard{}
		var i int
		var fail bool

		for i = 0; !fail && i < occIdx; i++ {
			magicIdx := int((occ[i] * magicNum) >> (64 - bitCount))
			if usedAttacks[magicIdx] == 0 {
				usedAttacks[magicIdx] = attacks[i]
			} else if usedAttacks[magicIdx] != attacks[i] {
				fail = true
			}
		}

		if !fail {
			return magicNum
		}
	}

	fmt.Println("FAIL")
	panic(1)
}

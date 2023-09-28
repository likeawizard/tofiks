package board

// Contains bitboard related pre-computed values and initialiaztion code

import (
	"fmt"
	"math/rand"
)

var (
	SquareBitboards   [64]BBoard
	PawnAttacks       [2][64]BBoard
	KnightAttacks     [64]BBoard
	KingAttacks       [64]BBoard
	KingSafetyMask    [2][64]BBoard
	BishopOccBitCount [64]int
	RookOccBitCount   [64]int
	BishopMagics      = [64]BBoard{
		0x20010400808600, 0xa008010410820000, 0x1004440082038008, 0x904040098084800,
		0x600c052000520541, 0x4002010420402022, 0x11040104400480, 0x200104104202080,
		0x1200210204080080, 0x6c18600204e20682, 0x2202004200e0, 0x100044404810840,
		0x400220211108110, 0x20002011009000c, 0xa00200a2084210, 0x202008098011000,
		0xc40002004019206, 0x116042040804c500, 0x419002080a80200a, 0x4000844000800,
		0x404b080a04800, 0x4608080482012002, 0x44040500a0880841, 0x2002100909050d00,
		0x8404004030a400, 0x90709004040080, 0x11444043040d0204, 0x8080100202020,
		0x801001181004000, 0x4140822002021000, 0x102089092009006, 0x540a042100540203,
		0x50100409482820, 0x8010880900041004, 0x230100500414, 0x200800050810,
		0x8294064010040100, 0x9010100220044404, 0x154202022004008e, 0x9420220008401,
		0x71080840110401, 0x2000a40420400201, 0x802619048001004, 0x209280a058000500,
		0x2004044810100a00, 0xa0208d000804300, 0x638a80d000684, 0x1910401000080,
		0x800420210400200, 0x4404410090100, 0x8020808400880000, 0x400081042120c21,
		0x4009001022120001, 0x4902220802082000, 0x410841000820290, 0x820020401002440,
		0x800420041084000, 0x10818c05a000, 0x301804213d000, 0x800040018208801,
		0x1b80000004104405, 0x2500214084184884, 0x1000628801050400, 0x8040229e24002080,
	}
)

var RookMagics = [64]BBoard{
	0x18010a040018000, 0x40002000401001, 0x290010a841e00100, 0x29001000050900a0,
	0x4080030400800800, 0x1200040200100801, 0x2200208200040851, 0x220000820425004c,
	0x104800740008020, 0x420400020005000, 0x844801000200480, 0x4004808008001000,
	0x4009000410080100, 0x3000400020900, 0x4804000810020104, 0x74800641800900,
	0x862818014400020, 0x40048020004480, 0x11a1010040200012, 0x20828010000800,
	0x848808004020800, 0x4522808004000200, 0x10100020004, 0x400206000092411c,
	0x818004444000a000, 0x180a000c0005002, 0xb104100200100, 0x24022202000a4010,
	0x100040080080080, 0x2010200080490, 0x180390400221098, 0x410008200010044,
	0x310400089800020, 0x8c0804009002902, 0x1004402001001504, 0x105021001000920,
	0x40080800801, 0xa02001002000804, 0x108284204005041, 0x8004082002411,
	0x2802281c0028001, 0x9044000910020, 0x200010008080, 0x40201001010008,
	0x8000080004008080, 0x3010400420080110, 0x414210040008, 0x10348400460001,
	0x80002000401040, 0x460200088400080, 0x8201822000100280, 0x600100008008280,
	0xc0800800040080, 0x24040080020080, 0x22c11a0108100c00, 0x204008114104200,
	0x8800800010290041, 0x401500228206, 0x8002a00011090041, 0x42008100101,
	0x283000800100205, 0x2008810010402, 0x490102200880104, 0x800010920940042,
}

var (
	BishopAttackMasks [64]BBoard
	RookAttackMasks   [64]BBoard
	BishopAttacks     [64][4096]BBoard
	RookAttacks       [64][4096]BBoard
)

var (
	PassedPawns   [2][64]BBoard
	IsolatedPawns [64]BBoard
	DoubledPawns  [64]BBoard
)

const (
	MajorDiag BBoard = 9314046665258451585
	MinorDiag BBoard = 4946458877011600706
)

func init() {
	InitSquares()
	InitOccBitCounts()
	InitPawnAttacks()
	InitKnightAttacks()
	InitKingAttacks()
	InitKingSafetyMasks()
	InitPawnStrucutreMasks()
	InitSliders()
}

func InitSquares() {
	for i := 0; i < 64; i++ {
		SquareBitboards[i] = 1 << i
	}
}

// Initialize pawn attack lookup table.
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

// Initialize Knight attack lookup table.
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

// Initialize King move lookup table.
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

// King safety masks ar similar to KingAttacks but do not cover squares behind the king. Only pieces in front of the kind attribute to safety.
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
// Relevant occupancy does not include border squares as a piece on the edge of the board can not block it moving past it.
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

// Initialize sliding piece lookup tables with magic numbers.
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

// Generate Bishop sliding attacks with a blocker bitboard on the fly. Only used for finding and initializing magic numbers. Too slow to use for movegen.
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

// Generate Rook sliding attacks with a blocker bitboard on the fly. Only used for finding and initializing magic numbers. Too slow to use for movegen.
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

// Calculate and initialize the number of relevant occupancies for Bishops and Rooks for all squares.
func InitOccBitCounts() {
	for sq := 0; sq < 64; sq++ {
		BishopOccBitCount[sq] = BishopRelOcc(sq).Count()
		RookOccBitCount[sq] = RookRelOcc(sq).Count()
	}
}

// Generate a random bitboard with few non-zero bits for Magic Number candidates.
func GetMagicNumber() BBoard {
	return BBoard(rand.Uint64() & rand.Uint64() & rand.Uint64())
}

// Find magic numbers with a brute force approach for a square.
func FindMagicNumber(sq, bitCount int, isBishop bool) BBoard {
	var occ, attacks, usedAttacks [4096]BBoard
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

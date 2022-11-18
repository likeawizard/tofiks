package book

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
)

const (
	black_pawn = iota
	white_pawn
	black_knight
	white_knight
	black_bishop
	white_bishop
	black_rook
	white_rook
	black_queen
	white_queen
	black_king
	white_king

	entrySize = 16
	sideHash  = 780
)

var BookMoves map[uint64][]polyEntry

type polyEntry struct {
	move   string
	weight uint16
}

type engineMove struct {
	move   board.Move
	weight uint16
}

func getPieceIdx(piece, row, file int) int {
	return 64*piece + 8*row + file
}

func convertPiece(piece, color int) int {
	color ^= 1
	switch piece {
	case board.PAWNS:
		piece = black_pawn
	case board.BISHOPS:
		piece = black_bishop
	case board.KNIGHTS:
		piece = black_knight
	case board.ROOKS:
		piece = black_rook
	case board.QUEENS:
		piece = black_queen
	case board.KINGS:
		piece = black_king
	}
	return piece + color
}

func squareToRowAndFile(sq int) (int, int) {
	return 7 - (sq / 8), sq % 8
}

// Check if current position is in book
func InBook(b *board.Board) bool {
	_, ok := BookMoves[PolyZobrist(b)]
	return ok
}

// Convert polyglot castling moves to UCI else return unchanged
func convertPolyToUCI(b *board.Board, polyMove string) string {
	switch polyMove {
	case "e1h1":
		if b.Pieces[board.WHITE][board.KINGS].LS1B() == board.E1 {
			return "e1g1"
		} else {
			return polyMove
		}
	case "e1a1":
		if b.Pieces[board.WHITE][board.KINGS].LS1B() == board.E1 {
			return "e1c1"
		} else {
			return polyMove
		}
	case "e8h8":
		if b.Pieces[board.BLACK][board.KINGS].LS1B() == board.E8 {
			return "e8g8"
		} else {
			return polyMove
		}
	case "e8a8":
		if b.Pieces[board.BLACK][board.KINGS].LS1B() == board.E8 {
			return "e8c8"
		} else {
			return polyMove
		}
	default:
		return polyMove
	}
}

// Prune book moves that are illegal in current position. Moves in book could be corrupted and castling has different notation
func pruneIllegal(b *board.Board, polyMoves []polyEntry) []engineMove {
	moves := b.MoveGenLegal()
	legal := make([]engineMove, 0)
	for _, pMove := range polyMoves {
		pMove.move = convertPolyToUCI(b, pMove.move)
		for _, move := range moves {
			if pMove.move == move.String() {
				legal = append(legal, engineMove{move: move, weight: pMove.weight})
			}
		}
	}
	return legal
}

// Get best scoring book move.
func GetBest(b *board.Board) board.Move {
	moves := getBookMoves(b)
	return moves[0].move
}

// Get weighted random
func GetWeighted(b *board.Board) board.Move {
	moves := getBookMoves(b)
	type bin struct {
		min  int
		max  int
		move board.Move
	}
	moveBins := make([]bin, len(moves))
	counter := 0
	for idx, move := range moves {
		moveBins[idx] = bin{min: counter, max: counter + int(move.weight), move: move.move}
		counter += int(move.weight)
	}
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(counter)
	for _, bin := range moveBins {
		if r > bin.min && r <= bin.max {
			return bin.move
		}
	}
	return moves[0].move
}

func getBookMoves(b *board.Board) []engineMove {
	moves := BookMoves[PolyZobrist(b)]
	return pruneIllegal(b, moves)
}

func PrintBookMoves(b *board.Board) {
	polyMoves := getBookMoves(b)

	if len(polyMoves) == 0 {
		fmt.Println("Out of book.")
	}
	totalWeight := float32(0)
	for _, pMove := range polyMoves {
		totalWeight += float32(pMove.weight)
	}
	for _, pMove := range polyMoves {
		fmt.Printf(" %s (%.1f%%)", pMove.move, 100*float32(pMove.weight)/totalWeight)
	}
	fmt.Println()

}

func LoadBook(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	BookMoves = make(map[uint64][]polyEntry)
	buffer := make([]byte, entrySize)
	reader := bufio.NewReader(file)

	lines := 0
	for numBytes, err := reader.Read(buffer); err == nil && numBytes == 16; numBytes, err = reader.Read(buffer) {
		if len(buffer) != 16 {
			continue
		}
		key, entry := decodeBookEntry(buffer)
		// if _, ok := BookMoves[key]; !ok {
		// 	BookMoves[key] = make([]polyEntry, 0)
		// }
		BookMoves[key] = append(BookMoves[key], entry)
		lines++
	}

	if err != nil {
		fmt.Println(err)
	}

	return lines
}

func decodeBookEntry(bytes []byte) (uint64, polyEntry) {
	key := binary.BigEndian.Uint64(bytes[:8])
	move := binary.BigEndian.Uint16(bytes[8:10])
	weight := binary.BigEndian.Uint16(bytes[10:12])

	return key, polyEntry{move: polyMoveToUCI(move), weight: weight}
}

func polyMoveToUCI(move uint16) string {
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	promoPiece := []string{"", "n", "b", "r", "q"}
	promo := move >> 12 & 7
	fromRow := move >> 9 & 7
	fromFile := move >> 6 & 7
	toRow := move >> 3 & 7
	toFile := move & 7
	return fmt.Sprintf("%s%d%s%d%s", files[fromFile], fromRow+1, files[toFile], toRow+1, promoPiece[promo])
}

func PolyZobrist(b *board.Board) uint64 {
	var hash uint64
	for color := board.WHITE; color <= board.BLACK; color++ {
		for piece := board.PAWNS; piece <= board.KINGS; piece++ {
			polyPiece := convertPiece(piece, color)
			pieces := b.Pieces[color][piece]
			for pieces > 0 {
				sq := pieces.PopLS1B()
				row, file := squareToRowAndFile(sq)
				// fmt.Printf("Piece: %d, Sqaure: %d (%v), Row: %d, File: %d, Polypiece: %d Offset:%d\n", piece, sq, board.Square(sq), row, file, polyPiece, getPieceIdx(polyPiece, row, file))
				hash ^= zobristHashes[getPieceIdx(polyPiece, row, file)]
			}
		}
	}

	cRs := [4]board.CastlingRights{board.WOO, board.WOOO, board.BOO, board.BOOO}
	polyCastling := [4]int{768, 769, 770, 771}

	for idx, cr := range cRs {
		if b.CastlingRights&cr != 0 {
			hash ^= zobristHashes[polyCastling[idx]]
		}
	}

	if b.EnPassantTarget > 0 && b.Pieces[b.Side][board.PAWNS]&board.PawnAttacks[b.Side^1][b.EnPassantTarget] != 0 {
		hash ^= zobristHashes[772+b.EnPassantTarget%8]
	}

	if b.Side == board.WHITE {
		hash ^= zobristHashes[sideHash]
	}

	return hash
}

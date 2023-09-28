package board

import (
	"fmt"
	"strconv"
	"strings"
)

func (b *Board) ExportFEN() string {
	fen := b.serializePosition()
	castlingRights := ""
	if b.CastlingRights != 0 {
		if b.CastlingRights&WOO != 0 {
			castlingRights += "K"
		}
		if b.CastlingRights&WOOO != 0 {
			castlingRights += "Q"
		}
		if b.CastlingRights&BOO != 0 {
			castlingRights += "k"
		}
		if b.CastlingRights&BOOO != 0 {
			castlingRights += "q"
		}
	} else {
		castlingRights = "-"
	}

	epString := "-"
	if b.EnPassantTarget != -1 {
		epString = b.EnPassantTarget.String()
	}

	sideToMove := WhiteToMove
	if b.Side == BLACK {
		sideToMove = BlackToMove
	}

	fen += fmt.Sprintf(" %c %s %s %d %d", sideToMove, castlingRights, epString, b.HalfMoveCounter, b.FullMoveCounter)
	return fen
}

func (b *Board) ImportFEN(fen string) error {
	fields := strings.Fields(fen)
	if len(fields) != 6 {
		return fmt.Errorf("FEN must contain six fields - '%s'", fen)
	}
	position := fields[0]
	sideToMove, castling, enPassant, halfMove, fullMove := fields[1], fields[2], fields[3], fields[4], fields[5]

	var err error
	b.parsePieces(position)

	if sideToMove[0] == WhiteToMove {
		b.Side = WHITE
	} else {
		b.Side = BLACK
	}
	fm, err := strconv.Atoi(fullMove)
	if err != nil {
		return err
	}
	b.FullMoveCounter = uint8(fm)

	hm, err := strconv.Atoi(halfMove)
	if err != nil {
		return err
	}
	b.HalfMoveCounter = uint8(hm)

	for _, c := range []byte(castling) {
		switch c {
		case 'K':
			b.CastlingRights |= WOO
		case 'Q':
			b.CastlingRights |= WOOO
		case 'k':
			b.CastlingRights |= BOO
		case 'q':
			b.CastlingRights |= BOOO
		}
	}

	if enPassant != "-" {
		b.EnPassantTarget = SquareFromString(enPassant)
	} else {
		b.EnPassantTarget = -1
	}

	b.Hash = b.SeedHash()

	return nil
}

func (b *Board) parsePieces(position string) {
	ranks := strings.Split(position, "/")
	for i, rankData := range ranks {
		file := 7
		for f := len(rankData) - 1; f >= 0; f-- {
			symbol := rankData[f : f+1]
			empty, err := strconv.Atoi(symbol)
			if err == nil {
				file -= empty
				continue
			}

			piece := BBoard(1 << (i*8 + file))
			switch symbol {
			case "P":
				b.Pieces[WHITE][PAWNS] |= piece
			case "B":
				b.Pieces[WHITE][BISHOPS] |= piece
			case "N":
				b.Pieces[WHITE][KNIGHTS] |= piece
			case "R":
				b.Pieces[WHITE][ROOKS] |= piece
			case "Q":
				b.Pieces[WHITE][QUEENS] |= piece
			case "K":
				b.Pieces[WHITE][KINGS] |= piece
			case "p":
				b.Pieces[BLACK][PAWNS] |= piece
			case "b":
				b.Pieces[BLACK][BISHOPS] |= piece
			case "n":
				b.Pieces[BLACK][KNIGHTS] |= piece
			case "r":
				b.Pieces[BLACK][ROOKS] |= piece
			case "q":
				b.Pieces[BLACK][QUEENS] |= piece
			case "k":
				b.Pieces[BLACK][KINGS] |= piece
			}
			file--
		}
	}
	for side := 0; side <= 1; side++ {
		for piece := 0; piece <= 5; piece++ {
			b.Occupancy[side] |= b.Pieces[side][piece]
		}
	}

	b.Occupancy[BOTH] = b.Occupancy[WHITE] | b.Occupancy[BLACK]
}

// Serialize the board into fen representation of piece placement.
func (b *Board) serializePosition() string {
	byteBoard := make([]byte, 64)
	for color := WHITE; color <= BLACK; color++ {
		for pieceType := PAWNS; pieceType <= KINGS; pieceType++ {
			pieces := b.Pieces[color][pieceType]
			var piece byte
			switch pieceType {
			case PAWNS:
				piece = 'P'
			case BISHOPS:
				piece = 'B'
			case KNIGHTS:
				piece = 'N'
			case ROOKS:
				piece = 'R'
			case QUEENS:
				piece = 'Q'
			case KINGS:
				piece = 'K'
			}
			piece += byte(color) * 32
			for pieces > 0 {
				sq := pieces.PopLS1B()
				byteBoard[sq] = piece
			}
		}
	}

	empty := 0
	fen := ""
	for i, val := range byteBoard {
		if i%8 == 0 {
			if empty > 0 {
				fen += fmt.Sprint(empty)
			}
			empty = 0
			if i != 0 {
				fen += "/"
			}
		}
		if val == 0 {
			empty++
		} else {
			if empty > 0 {
				fen += fmt.Sprint(empty)
				empty = 0
			}
			fen += string([]byte{val})
		}
	}

	if empty > 0 {
		fen += fmt.Sprint(empty)
	}

	return fen
}

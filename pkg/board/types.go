package board

type (
	// BBoard is a bitboard type.
	BBoard uint64
	// CastlingRights is a enum type for various castling rights.
	CastlingRights byte

	// Board represents the state of the chess board.
	Board struct {
		Hash            uint64
		Pieces          [2][6]BBoard
		Occupancy       [3]BBoard
		Phase           int
		EnPassantTarget Square
		HalfMoveCounter uint8
		FullMoveCounter uint8
		Side            int8
		CastlingRights  CastlingRights
		InCheck         bool
	}
)

const (
	WOO CastlingRights = 1 << iota
	WOOO
	BOO
	BOOO
	CASTLING_ALL = WOO | WOOO | BOO | BOOO
)

const (
	// StartPos is UCI shorthand for Starting Position FEN.
	StartPos = "startpos"
	// StartingFEN contains the position FEN.
	StartingFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
)

const (
	// Constants for the board Side and Occupancy.
	WHITE = 0
	BLACK = 1
	BOTH  = 2

	// Constants for the board Pieces.
	PAWNS    = 0
	BISHOPS  = 1
	KNIGHTS  = 2
	ROOKS    = 3
	QUEENS   = 4
	KINGS    = 5
	NO_PIECE = 6

	// File and Rank constants.
	AFile BBoard = 72340172838076673
	BFile BBoard = 144680345676153346
	CFile BBoard = 289360691352306692
	DFile BBoard = 578721382704613384
	EFile BBoard = 1157442765409226768
	FFile BBoard = 2314885530818453536
	GFile BBoard = 4629771061636907072
	HFile BBoard = 9259542123273814144

	Rank8 BBoard = 255
	Rank7        = Rank8 << 8
	Rank6        = Rank7 << 8
	Rank5        = Rank6 << 8
	Rank4        = Rank5 << 8
	Rank3        = Rank4 << 8
	Rank2        = Rank3 << 8
	Rank1        = Rank2 << 8

	// Constants for squares relevant to castling legality.
	F1G1 = Rank1&FFile | Rank1&GFile
	D1C1 = Rank1&DFile | Rank1&CFile
	D1B1 = Rank1&DFile | Rank1&CFile | Rank1&BFile
	F8G8 = Rank8&FFile | Rank8&GFile
	D8C8 = Rank8&DFile | Rank8&CFile
	D8B8 = Rank8&DFile | Rank8&CFile | Rank8&BFile
)

const (
	A8 int = iota
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A1
	B1
	C1
	D1
	E1
	F1
	G1
	H1
)

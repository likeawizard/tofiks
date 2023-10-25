package testsuite

type drawByThreefold struct {
	fen    string
	moves  string
	number int
}

var drawByThreefoldPositions = []drawByThreefold{
	{
		number: 1,
		fen:    "6k1/5p1p/6p1/8/1p1q1P2/1QP3P1/1P3RKP/4r3 b - - 2 35",
		moves:  "d4e4 g2h3 e4f5 h3g2 f5e4 g2h3 e4f5 h3g2 f5e4",
	},
	{
		number: 2,
		fen:    "3r1rk1/1p1b1pp1/p1p1p1qp/P1Pp4/1P1PPP2/2N3R1/2Q2RPP/6K1 b - - 8 28",
		moves:  "g6h5 g3h3 h5g4 h3g3 g4h5 g3h3 h5g4 h3g3 g4h5",
	},
	{
		number: 3,
		fen:    "8/2R5/6pk/5p1p/4p2P/6P1/1r3PK1/8 w - - 2 52",
		moves:  "g2g1 b2a2 g1f1 a2b2 f1g1 b2a2 g1f1 a2b2 f1g1",
	},
	{
		number: 4,
		fen:    "8/2BP4/2K3k1/8/1q5p/4PB2/8/8 b - - 8 52",
		moves:  "b4a4 c6d6 a4a3 d6c6 a3a4 c6d6 a4b4 d6c6 b4a4",
	},
	{
		number: 5,
		fen:    "1k6/8/1p3b1R/1P3Np1/1P2r2p/1K6/8/8 b - - 14 83",
		moves:  "f6d8 h6h8 b8c7 h8h7 c7b8 h7h8 b8c7 h8h7 c7b8 h7h8",
	},
	{
		number: 6,
		fen:    "2r1kb1r/5p1p/p1q2p2/3Npb2/8/PN6/2PQ2PP/R2R1K2 w k - 10 27",
		moves:  "b3a5 c6d6 a5b7 d6c6 b7a5 c6d6 a5b7 d6c6 b7a5",
	},
	{
		number: 7,
		fen:    "6k1/5q2/3p3p/1pnP1Pp1/3Q4/r7/2B2PK1/7R b - - 1 46",
		moves:  "f7f8 d4b4 f8a8 b4d4 a8f8 d4b4 f8a8 b4d4 a8f8",
	},
}

var forceThreeFoldPositions = []drawByThreefold{
	{
		number: 1,
		fen:    "7k/3p2p1/4Q3/8/4p3/P2q4/PP1r4/K7 w - - 0 1",
		moves:  "",
	},
}

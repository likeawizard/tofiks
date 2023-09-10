package testsuite

type matePosition struct {
	mateIn int
	fen    string
}

var matePositions = []matePosition{
	{
		mateIn: -3,
		fen:    "8/2k5/p7/5p2/8/8/4q3/6K1 w - - 5 82",
	},
	{
		mateIn: 2,
		fen:    "8/2k5/8/p7/5p2/8/4q3/7K b - - 1 84",
	},
	{
		mateIn: 3,
		fen:    "8/2k5/p7/5p2/8/8/4q3/7K b - - 6 82",
	},
	{
		mateIn: 5,
		fen:    "8/2k5/p7/5p2/8/8/7K/5q2 b - - 2 80",
	},
	{
		mateIn: 5,
		fen:    "5rk1/2p1RNbr/8/8/p2p2Q1/P1P3P1/1P4P1/6K1 w - - 12 49",
	},
	{
		mateIn: -4,
		fen:    "8/2k5/p7/5p2/8/5q2/7K/8 w - - 3 81",
	},
	{
		mateIn: -4,
		fen:    "5rk1/2p1R1br/8/4N3/p2p2Q1/P1P3P1/1P4P1/6K1 b - - 13 49",
	},
	{
		mateIn: -1,
		fen:    "3r1r1k/pp4R1/4p1Qp/4Np1P/3p4/P7/8/R5K1 b - - 0 34",
	},
	{
		mateIn: 1,
		fen:    "3r1r1k/pp4R1/4p1Qp/4Np1P/8/P2p4/8/R5K1 w - - 0 35",
	},
	{
		mateIn: 1,
		fen:    "1k6/2p5/2p5/8/1K6/2Pq4/8/r7 b - - 7 59",
	},
	{
		mateIn: -1,
		fen:    "1k6/2p5/2p5/8/8/1KPq4/8/r7 w - - 6 59",
	},
}

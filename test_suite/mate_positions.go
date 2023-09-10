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
}

package testsuite

type perftPos struct {
	position string
	fen      string
	results  []perft
}

type perft struct {
	depth int
	nodes int64
}

func (pp perftPos) getResultAtDepth(depth int) perft {
	depth = min(depth, len(pp.results))
	return pp.results[depth-1]
}

var perftResults = []perftPos{
	{
		position: "Start position",
		fen:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		results: []perft{
			{1, 20},
			{2, 400},
			{3, 8902},
			{4, 197281},
			{5, 4865609},
			{6, 119060324},
			{7, 3195901860},
			{8, 84998978956},
			{9, 2439530234167},
			{10, 69352859712417},
		},
	},
	{
		position: "Position 1",
		fen:      "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		results: []perft{
			{1, 48},
			{2, 2039},
			{3, 97862},
			{4, 4085603},
			{5, 193690690},
			{6, 8031647685},
		},
	},
	{
		position: "Position 2",
		fen:      "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		results: []perft{
			{1, 14},
			{2, 191},
			{3, 2812},
			{4, 43238},
			{5, 674624},
			{6, 11030083},
			{7, 178633661},
			{8, 3009794393},
		},
	},
	{
		position: "Position 3",
		fen:      "r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1",
		results: []perft{
			{1, 6},
			{2, 264},
			{3, 9467},
			{4, 422333},
			{5, 15833292},
			{6, 706045033},
		},
	},
	{
		position: "Position 4",
		fen:      "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		results: []perft{
			{1, 44},
			{2, 1486},
			{3, 62379},
			{4, 2103487},
			{5, 89941194},
		},
	},
	{
		position: "Position 5",
		fen:      "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		results: []perft{
			{1, 46},
			{2, 2079},
			{3, 89890},
			{4, 3894594},
			{5, 164075551},
			{6, 6923051137},
			{7, 287188994746},
			{8, 11923589843526},
			{9, 490154852788714},
		},
	},
}

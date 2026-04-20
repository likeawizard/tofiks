//go:build debug

package testsuite

// healthPos is a position used by TestSearchHealth. Edit healthPositions to
// add positions with interesting search behaviour or drop ones that no longer
// tell you anything new.
type healthPos struct {
	name string
	fen  string
}

var healthPositions = []healthPos{
	{"Start position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
	{"Kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"},
	{"Endgame KR vs Kp", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"},
	{"Promo edge", "r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1"},
	{"Mirror", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"},
	{"Middlegame", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"},
}

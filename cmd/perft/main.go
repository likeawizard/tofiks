package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/likeawizard/tofiks/internal/board"
)

type perftTest struct {
	name  string
	fen   string
	count []int
}

var tests = []perftTest{
	{
		name: "Test 1",
		fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		count: []int{
			20,
			400,
			8902,
			197281,
			4865609,
			119060324,
			3195901860,
			84998978956,
			2439530234167,
			69352859712417,
		},
	},
	{
		name: "Test 2",
		fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		count: []int{
			48,
			2039,
			97862,
			4085603,
			193690690,
			8031647685,
		},
	},
	{
		name: "Test 3",
		fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		count: []int{
			14,
			191,
			2812,
			43238,
			674624,
			11030083,
			178633661,
			3009794393,
		},
	},
	{
		name: "Test 4",
		fen:  "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		count: []int{
			6,
			264,
			9467,
			422333,
			15833292,
			706045033,
		},
	},
	{
		name: "Test 5",
		fen:  "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		count: []int{
			44,
			1486,
			62379,
			2103487,
			89941194,
		},
	},
	{
		name: "Test 6",
		fen:  "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		count: []int{
			46,
			2079,
			89890,
			3894594,
			164075551,
			6923051137,
			287188994746,
			11923589843526,
			490154852788714,
		},
	},
}

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	var depth int
	var fen string
	flag.IntVar(&depth, "d", 4, "default: 4")
	flag.StringVar(&fen, "fen", "", "Optional: Run perft debug on specific FEN")
	flag.Parse()

	start := time.Now()

	if fen == "" {
		var containsError bool

		for _, test := range tests {
			fmt.Printf("%s (fen: %s)\n", test.name, test.fen)
			for d, count := range test.count {
				if d >= depth {
					break
				}
				leafs, perf := board.Perft(test.fen, d+1)
				fmt.Printf("depth %d: %d %v %v\n", d+1, leafs, perf, leafs == count)
				if leafs != count {
					containsError = true
				}
			}
		}

		fmt.Printf("\nRun time: %v\n", time.Since(start))
		if !containsError {
			fmt.Println("All tests passed successfully.")
		} else {
			fmt.Println("Encountered errors")
		}
	} else {
		board.PerftDebug(fen, depth)
	}
}

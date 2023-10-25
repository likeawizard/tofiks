package board

import (
	"fmt"
	"time"
)

func Perft(fen string, depth int) (int64, time.Duration) {
	b := &Board{}
	err := b.ImportFEN(fen)
	if err != nil {
		panic(err)
	}
	start := time.Now()
	leafs := traverse(b, depth)
	return leafs, time.Since(start)
}

func traverse(b *Board, depth int) int64 {
	num := int64(0)
	if depth == 0 {
		return 1
	} else {
		all := b.PseudoMoveGen()
		for i := 0; i < len(all); i++ {
			umove := b.MakeMove(all[i])
			if b.IsChecked(b.Side ^ 1) {
				umove()
				continue
			}

			// Not part of the actual perft but a Zobrist and tt health check. To ensure updated hash is the same as one calculated from scratch.
			// This takes additional compute power and reduces the performance.
			// As perft is used only for testing incremental improvement and consistency, I am not concerned with getting the biggest possible number.
			if b.Hash != b.SeedHash() {
				fmt.Println(b.ExportFEN())
				umove()
				fmt.Println(b.ExportFEN(), all[i])
				panic(1)
			}
			num += traverse(b, depth-1)
			umove()
		}
		return num
	}
}

func (b *Board) PerftDebug(depth int) {
	all := b.PseudoMoveGen()
	start := time.Now()
	nodesSearched := int64(0)
	for _, move := range all {
		umove := b.MakeMove(move)
		if b.IsChecked(b.Side ^ 1) {
			umove()
			continue
		}
		nodes := traverse(b, depth-1)
		nodesSearched += nodes
		fmt.Printf("%s: %d\n", move, nodes)
		umove()
	}
	fmt.Printf("\nNodes searched: %d (nps %d)\n", nodesSearched, (1000000*nodesSearched)/time.Since(start).Microseconds())
}

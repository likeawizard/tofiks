package board

import (
	"fmt"
	"time"
)

func Perft(fen string, depth int) (int, time.Duration) {
	b := &Board{}
	b.ImportFEN(fen)
	start := time.Now()
	leafs := traverse(b, depth)
	return leafs, time.Since(start)
}

func traverse(b *Board, depth int) int {
	num := 0

	if depth == 1 {

		return len(b.MoveGen())
	} else {
		all := b.PseudoMoveGen()
		for i := 0; i < len(all); i++ {
			umove := b.MakeMove(all[i])
			if b.IsChecked(b.Side ^ 1) {
				umove()
				continue
			}
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

func PerftDebug(fen string, depth int) {
	b := &Board{}
	b.ImportFEN(fen)
	all := b.PseudoMoveGen()

	nodesSearched := 0
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
	fmt.Println("\nNodes searched: ", nodesSearched)
}

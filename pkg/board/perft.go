package board

import (
	"fmt"
	"time"
)

func Perft(fen string, depth int) (int64, time.Duration) {
	b := &Board{}
	b.ImportFEN(fen)
	start := time.Now()
	leafs := traverse(b, depth)
	return leafs, time.Since(start)
}

func traverse(b *Board, depth int) int64 {
	num := int64(0)

	if depth == 1 {
		return int64(len(b.MoveGenLegal()))
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

package board

import (
	"fmt"
	"math/rand"
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

			// Not part of the actual perft but a Zobrist and tt health check. To ensure updated hash is the same as one calculated from scratch.
			// This takes additional compute power and reduces the performance.
			// As perft is used only for testing incremental improvement and consistency, I am not concerned with getting the biggest possible number.
			if !ttDataHealthCheck(all[i]) {
				panic("invalid packaging")
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
		if b.InCheck {
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

// Simple check for ensuring hash table data packing and unpacking
// Copy & pasted from eval package to avoid cyclic import
func ttDataHealthCheck(move Move) bool {

	const (
		move_mask   = (1 << 16) - 1
		type_mask   = (1 << 8) - 1
		depth_mask  = type_mask
		score_mask  = (1 << 32) - 1
		depth_shift = 16
		type_shift  = 24
		score_shift = 32
	)

	pack := func(move Move, depth int8, eType int8, score int32) uint64 {
		return uint64(move) |
			uint64(depth)<<depth_shift |
			uint64(eType)<<type_shift |
			uint64(score)<<score_shift
	}

	unpack := func(data uint64) (Move, int8, int8, int32) {
		return Move(data & move_mask),
			int8((data >> depth_shift) & type_mask),
			int8((data >> type_shift) & type_mask),
			int32(data >> score_shift)
	}
	depth := int8(rand.Int31n(63))
	ttype := int8(rand.Int() % 3)
	score := rand.Int31n(2*90000) - 90000

	data := pack(move, depth, ttype, score)

	upmove, updeth, upttype, upscore := unpack(data)
	// fmt.Println(move, depth, ttype, score)
	// fmt.Println(unpack(data))
	return upmove == move && updeth == depth && upttype == ttype && upscore == score
}

package eval

import (
	"context"

	"github.com/likeawizard/tofiks/pkg/board"
)

// Start number of helper threads to build the hash table
func (e *EvalEngine) StartHelpers(ctx context.Context, depth, numThreads int8) context.CancelFunc {
	helperCtx, cancelFn := context.WithCancel(ctx)
	for offset := int8(1); offset <= numThreads; offset++ {
		go e.HelperIDSearch(helperCtx, e.Board.Copy(), depth, offset)
	}
	return cancelFn
}

// Helper thread Iterative Deepening Search. Offset is used as starting depth to desynchronize the threads.
func (e *EvalEngine) HelperIDSearch(ctx context.Context, b *board.Board, depth, offset int8) {
	color := int32(1)
	if b.Side != board.WHITE {
		color = -color
	}
	for d := depth + offset%2; d <= int8(63); d++ {
		select {
		case <-ctx.Done():
			return
		default:
			e.PVSHelper(ctx, b, d, 0, -Inf, Inf, true, color)
			// fmt.Printf("Helper #%d finshed depth %d\n", offset, d)
		}
	}
}

func (e *EvalEngine) PVSHelper(ctx context.Context, b *board.Board, depth, ply int8, alpha, beta int32, nmp bool, side int32) int32 {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		inCheck := b.InCheck
		// If search depth is reached and not in check enter Qsearch
		if depth <= 0 && !inCheck {
			return e.quiescenceHelper(ctx, b, alpha, beta, side)
		} else if depth <= 0 { // If depth is reached and we are in check extend
			depth++
		}

		// if ply > 0 && (b.HalfMoveCounter >= 100 || b.InsufficentMaterial() || e.IsDrawByRepetition()) {
		// 	return 0
		// }

		var pvMove board.Move
		if entry, ok := e.TTable.Probe(b.Hash); ok && ply > 0 && entry.Depth() >= depth {
			if eval, ok := entry.GetScore(depth, ply, alpha, beta); ok {
				return eval
			}
			pvMove = entry.Move()
		}

		// Null move pruning.
		// Do not prune:
		// - when in check.
		// - when less than 7 pieces on board (random heuristic) or pawn only endgame due to possible zugzwang situations
		if !inCheck && nmp && b.Occupancy[board.BOTH].Count() > 6 && !b.IsPawnOnly() {
			unull := b.MakeNullMove()
			R := 3 + depth/6
			value := -e.PVSHelper(ctx, b, depth-R-1, ply+1, -beta, -beta+1, false, -side)
			unull()
			if value >= beta {
				return beta
			}
		}

		all := b.PseudoMoveGen()
		legalMoves := 0
		e.OrderMovesPV(pvMove, &all, &all, ply)

		value := int32(0)
		entryType := TT_UPPER
		bestVal := -Inf
		var bestMove board.Move
		for i := 0; i < len(all); i++ {
			umove := b.MakeMove(all[i])
			if b.IsChecked(b.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++
			// e.IncrementHistory()
			if legalMoves == 1 {
				value = -e.PVSHelper(ctx, b, depth-1, ply+1, -beta, -alpha, true, -side)
			} else {
				value = -e.PVSHelper(ctx, b, depth-1, ply+1, -(alpha + 1), -alpha, true, -side)
				if value > alpha {
					value = -e.PVSHelper(ctx, b, depth-1, ply+1, -beta, -alpha, true, -side)
				}
			}
			umove()
			// e.DecrementHistory()

			if value > bestVal {
				bestVal = value
				bestMove = all[i]
			}

			if value >= beta {
				e.AddKillerMove(ply, all[i])
				entryType = TT_LOWER
				break
			}

			if value > alpha {
				entryType = TT_EXACT
				alpha = value
			}

		}

		if legalMoves == 0 {
			if inCheck {
				return int32(ply) - CheckmateScore
			} else {
				return 0
			}
		}
		e.TTable.Store(b.Hash, entryType, bestVal, depth, bestMove)
		return bestVal
	}
}

func (e *EvalEngine) quiescenceHelper(ctx context.Context, b *board.Board, alpha, beta, side int32) int32 {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		eval := side * int32(e.GetEvaluation(b))

		if eval >= beta {
			return beta
		}

		if eval > alpha {
			alpha = eval
		}
		var all []board.Move
		inCheck := b.InCheck
		if inCheck {
			all = b.PseudoMoveGen()
		} else {
			all = b.PseudoCaptureAndQueenPromoGen()
		}

		legalMoves := 0
		e.OrderMoves(&all)

		value := -Inf
		for i := 0; i < len(all); i++ {
			umove := b.MakeMove(all[i])
			if b.IsChecked(b.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++
			value = max(value, -e.quiescenceHelper(ctx, b, -beta, -alpha, -side))
			umove()
			alpha = max(value, alpha)
			if alpha >= beta {
				break
			}
		}
		if legalMoves == 0 {
			return eval
		}
		return value
	}
}

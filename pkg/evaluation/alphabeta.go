package eval

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
)

const (
	// Mate score to be adjusted by the ply that it is found on by subtracting the ply to favor shorter mates.
	CheckmateScore int32 = 90000
	// ply adjusted adjusted mates scores should not exceed this value and anything above this should be considered a mate instead of normal eval
	CheckmateThreshold = CheckmateScore - 1000
	Inf                = 2 * CheckmateScore
)

func (e *EvalEngine) PVS(ctx context.Context, line *[]board.Move, depth, ply int8, alpha, beta int32, nmp bool, side int32) int32 {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		inCheck := e.Board.IsChecked(e.Board.Side)
		// If search depth is reached and not in check enter Qsearch
		if depth <= 0 && !inCheck {
			return e.quiescence(ctx, alpha, beta, side)
		} else if depth <= 0 { // If depth is reached and we are in check extend
			depth++
		}

		e.Stats.nodes++

		if ply > 0 && (e.Board.HalfMoveCounter >= 100 || e.Board.InsufficentMaterial() || e.IsDrawByRepetition()) {
			return 0
		}

		var pvMove board.Move
		if entry, ok := e.TTable.Probe(e.Board.Hash); ok && ply > 0 && entry.Depth() >= depth {
			if eval, ok := entry.GetScore(depth, ply, alpha, beta); ok {
				return eval
			}
			pvMove = entry.Move()
		}

		// Null move pruning.
		// Do not prune:
		// - when in check.
		// - when less than 7 pieces on board (random heuristic) or pawn only endgame due to possible zugzwang situations
		if !inCheck && nmp && e.Board.Occupancy[board.BOTH].Count() > 6 && !e.Board.IsPawnOnly() {
			unull := e.Board.MakeNullMove()
			R := 3 + depth/6
			value := -e.PVS(ctx, &[]board.Move{}, depth-R-1, ply+1, -beta, -beta+1, false, -side)
			unull()
			if value >= beta {
				return beta
			}
		}

		all := e.Board.PseudoMoveGen()
		legalMoves := 0
		e.OrderMoves(pvMove, &all, ply)

		value := int32(0)
		entryType := TT_UPPER
		bestVal := -Inf
		var bestMove board.Move
		pv := []board.Move{}
		for i := 0; i < len(all); i++ {
			umove := e.Board.MakeMove(all[i])
			if e.Board.IsChecked(e.Board.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++
			e.IncrementHistory()
			if legalMoves == 1 {
				value = -e.PVS(ctx, &pv, depth-1, ply+1, -beta, -alpha, true, -side)
			} else {
				value = -e.PVS(ctx, &pv, depth-1, ply+1, -(alpha + 1), -alpha, true, -side)
				if value > alpha {
					value = -e.PVS(ctx, &pv, depth-1, ply+1, -beta, -alpha, true, -side)
				}
			}
			umove()
			e.DecrementHistory()

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
				*line = []board.Move{all[i]}
				*line = append(*line, pv...)
			}

		}

		if legalMoves == 0 {
			if inCheck {
				return int32(ply) - CheckmateScore
			} else {
				return 0
			}
		}
		e.TTable.Store(e.Board.Hash, entryType, bestVal, depth, bestMove)
		return bestVal
	}
}

func (e *EvalEngine) quiescence(ctx context.Context, alpha, beta, side int32) int32 {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		e.Stats.qNodes++
		eval := side * int32(e.GetEvaluation(e.Board))

		if eval >= beta {
			return beta
		}

		if eval > alpha {
			alpha = eval
		}
		var all []board.Move
		inCheck := e.Board.IsChecked(e.Board.Side)
		if inCheck {
			all = e.Board.PseudoMoveGen()
		} else {
			all = e.Board.PseudoCaptureAndQueenPromoGen()
		}

		legalMoves := 0
		e.OrderMoves(0, &all, 0)

		value := -Inf
		for i := 0; i < len(all); i++ {
			umove := e.Board.MakeMove(all[i])
			if e.Board.IsChecked(e.Board.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++
			value = Max(value, -e.quiescence(ctx, -beta, -alpha, -side))
			umove()
			alpha = Max(value, alpha)
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

// Iterative deepening search. Returns best move, ponder and ok if search succeeded.
func (e *EvalEngine) IDSearch(ctx context.Context, depth int, infinite bool) (board.Move, board.Move, bool) {
	e.MateFound = false
	var wg sync.WaitGroup
	var best, ponder board.Move
	var eval int32
	var line []board.Move
	start := time.Now()
	color := int32(1)
	alpha, beta := -Inf, Inf
	if e.Board.Side != board.WHITE {
		color = -color
	}
	done, ok := false, true
	wg.Add(1)
	go func() {
		for d := int8(1); d <= int8(depth); d++ {
			if done {
				wg.Done()
				return
			}

			e.Stats.Start()
			// stopHelpers := e.StartHelpers(ctx, d, 3)
			eval = e.PVS(ctx, &line, d, 0, alpha, beta, true, color)
			// stopHelpers()

			select {
			case <-ctx.Done():
				// Do nothing as alpha-beta was canceled and results are unreliable
				done = true
				wg.Done()
				return
			default:
				if len(line) == 0 {
					done, ok = true, false
					break
				} else {
					best = line[0]
					if len(line) > 1 {
						ponder = line[1]
					}
				}
				lineStr := ""
				for _, m := range line {
					lineStr += " " + m.String()
				}
				totalN := e.Stats.nodes + e.Stats.qNodes
				timeSince := time.Since(start)
				nps := int64(totalN)
				if timeSince.Milliseconds() != 0 {
					nps = (1000 * nps) / timeSince.Milliseconds()
				}
				fmt.Printf("info depth %d score %s nodes %d nps %d time %d hashfull %d pv%s\n", d, e.ConvertEvalToScore(eval), totalN, nps, timeSince.Milliseconds(), e.TTable.Hashfull(), lineStr)
				// fmt.Printf("hashfull %d, newwrite %d, overwrite %d, rejected %d\n", e.TTable.Hashfull(), e.TTable.newWrite, e.TTable.overWrite, e.TTable.rejectedWrite)
				if eval > CheckmateThreshold || eval < -CheckmateThreshold {
					e.MateFound = true
				}

				//found mate stop
				if !infinite && e.MateFound {
					done = true
				}
			}
		}
		wg.Done()
	}()

	wg.Wait()
	return best, ponder, ok
}

package eval

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
)

const (
	// CheckmateScore score to be adjusted by the ply that it is found on by subtracting the ply to favor shorter mates.
	CheckmateScore int16 = 8192
	// CheckmateThreshold ply adjusted mates scores should not exceed this value and anything above this should be considered a mate instead of normal eval.
	CheckmateThreshold = CheckmateScore - 1024
	// Inf should be an unachievable score treated as infinity.
	Inf = 2 * CheckmateScore
)

func (e *Engine) PVS(ctx context.Context, pvOrder []board.Move, line *[]board.Move, depth, ply int8, alpha, beta int16, nmp bool, side int16) int16 {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		isPV := beta-alpha != 1
		inCheck := e.Board.InCheck

		if e.Board.InCheck {
			depth++
		}

		// If search depth is reached and not in check enter Qsearch
		if depth <= 0 {
			return e.Quiescence(ctx, alpha, beta, side)
		}

		e.Stats.nodes++

		if ply > 0 && (e.Board.HalfMoveCounter >= 100 || e.Board.InsufficientMaterial() || e.IsDrawByRepetition()) {
			return 0
		}

		var pvMove board.Move
		if entry, ok := e.TTable.Probe(e.Board.Hash); ok && ply > 0 && entry.Depth() >= depth {
			if eval, ok := entry.GetScore(depth, ply, alpha, beta); ok {
				*line = []board.Move{entry.Move()}
				return eval
			}
			pvMove = entry.Move()
		}

		// Null move pruning.
		// Do not prune:
		// - when in check.
		// - when less than 7 pieces on board (random heuristic) or pawn only endgame due to possible zugzwang situations
		if !isPV && !inCheck && nmp && e.Board.Occupancy[board.BOTH].Count() > 6 && !e.Board.IsPawnOnly() {
			unull := e.Board.MakeNullMove()
			R := 3 + depth/6
			value := -e.PVS(ctx, pvOrder, &[]board.Move{}, depth-R-1, ply+1, -beta, -beta+1, false, -side)
			unull()
			if value >= beta {
				return beta
			}
		}

		all := e.Board.PseudoMoveGen()
		legalMoves := 0
		selectMove := e.GetMoveSelector(pvMove, all, pvOrder, ply)

		value := int16(0)
		entryType := TT_UPPER
		bestVal := -Inf
		var currMove, bestMove board.Move
		var pv []board.Move
		moveCount := len(all)
		if moveCount > 0 {
			bestMove = all[0]
		}

		for i := 0; i < moveCount; i++ {
			currMove = selectMove(i)
			umove := e.Board.MakeMove(currMove)
			if e.Board.IsChecked(e.Board.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++

			if !isPV && !inCheck && depth < ply/2 && legalMoves > 8+(int(depth))*4 && currMove.Promotion() == 0 {
				umove()
				continue
			}

			e.AddPly()
			pv = []board.Move{}
			if legalMoves == 1 {
				value = -e.PVS(ctx, pvOrder, &pv, depth-1, ply+1, -beta, -alpha, true, -side)
			} else {
				depthR := int8(0)
				if !isPV && legalMoves > 4 && !inCheck && depth > 3 &&
					currMove.Promotion() == 0 && !currMove.IsEnPassant() && board.SquareBitboards[currMove.To()]&e.Board.Occupancy[board.BOTH] == 0 {
					depthR = max(2, depth/4) + int8(legalMoves)/8
				}

				value = -e.PVS(ctx, pvOrder, &pv, depth-1-depthR, ply+1, -(alpha + 1), -alpha, true, -side)

				if value > alpha && value < beta {
					value = -e.PVS(ctx, pvOrder, &pv, depth-1, ply+1, -beta, -alpha, true, -side)
				}
			}
			umove()
			e.RemovePly()

			if value > bestVal {
				bestVal = value
				bestMove = currMove
			}

			if value >= beta {
				if !currMove.IsCapture() {
					e.AddKillerMove(ply, currMove)
					e.IncrementHistory(depth, currMove)
				}

				entryType = TT_LOWER
				break
			}
			e.DecrementHistory(currMove)

			if value > alpha {
				entryType = TT_EXACT
				bestMove = currMove
				alpha = value
				*line = []board.Move{currMove}
				*line = append(*line, pv...)
			}
		}

		if legalMoves == 0 {
			if inCheck {
				return int16(ply) - CheckmateScore
			}

			return 0
		}
		e.TTable.Store(e.Board.Hash, entryType, bestVal, depth, bestMove)
		return bestVal
	}
}

func (e *Engine) Quiescence(ctx context.Context, alpha, beta, side int16) int16 {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		e.Stats.qNodes++
		eval := side * int16(e.GetEvaluation(e.Board))

		if !e.Board.InCheck && eval >= beta {
			return beta
		}

		if !e.Board.InCheck && eval < alpha-975 {
			return alpha
		}

		if eval > alpha {
			alpha = eval
		}
		var all []board.Move
		if e.Board.InCheck {
			all = e.Board.PseudoMoveGen()
		} else {
			all = e.Board.PseudoCaptureAndQueenPromoGen()
		}

		legalMoves := 0

		selectMove := e.GetMoveSelectorQ(all)
		var currMove board.Move
		value := -Inf
		for i := 0; i < len(all); i++ {
			currMove = selectMove(i)
			umove := e.Board.MakeMove(currMove)
			if e.Board.IsChecked(e.Board.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++
			value = max(value, -e.Quiescence(ctx, -beta, -alpha, -side))
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

// Iterative deepening search. Returns best move, ponder and ok if search succeeded.
func (e *Engine) IDSearch(depth int, infinite bool) (board.Move, board.Move, bool) {
	ctx, cancel := e.Clock.GetContext(int(e.Board.FullMoveCounter), e.Board.Side)
	go func() {
		select {
		case <-e.Stop:
			cancel()
			return
		}
	}()
	defer cancel()

	e.MateFound = false
	var wg sync.WaitGroup
	var best, ponder board.Move
	var eval int16
	var line []board.Move
	start := time.Now()
	color := int16(1)
	alpha, beta := -Inf, Inf
	if e.Board.Side != board.WHITE {
		color = -color
	}
	e.TTable.age = max(0, int8(e.Board.HalfMoveCounter)/2+63-int8(e.Board.FullMoveCounter)/4)
	e.AgeHistory()
	done, ok := false, true
	wg.Add(1)
	go func() {
		for d := int8(1); d <= int8(depth); d++ {
			if done {
				wg.Done()
				return
			}

			e.Stats.Start()
			var pv []board.Move
			pv = append(pv, line...)
			eval = e.PVS(ctx, pv, &line, d, 0, alpha, beta, true, color)

			if eval <= alpha || eval >= beta {
				alpha, beta = -Inf, Inf
				eval = e.PVS(ctx, pv, &line, d, 0, alpha, beta, true, color)
			}
			alpha, beta = eval-50, eval+100

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
				if eval > CheckmateThreshold || eval < -CheckmateThreshold {
					e.MateFound = true
				}

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

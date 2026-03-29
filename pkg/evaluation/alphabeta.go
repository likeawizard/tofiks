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
			return e.Quiescence(ctx, ply, alpha, beta, side)
		}

		e.Stats.nodes++

		if ply > 0 && (e.Board.HalfMoveCounter >= 100 || e.Board.InsufficientMaterial() || e.IsDrawByRepetition()) {
			return 0
		}

		// Static eval for pruning decisions.
		var staticEval int16
		canFutility := !isPV && !inCheck && beta > -CheckmateThreshold && beta < CheckmateThreshold
		if canFutility && depth <= 6 {
			staticEval = side * int16(e.GetEvaluation(e.Board))

			// Reverse futility pruning. If static eval is well above beta at shallow depths,
			// the opponent is unlikely to improve their position enough to drop below beta.
			if depth <= 5 && staticEval-100*int16(depth) >= beta {
				return staticEval
			}
		}

		var pvMove board.Move
		if entry, ok := e.TTable.Probe(e.Board.Hash); ok && ply > 0 && entry.Depth() >= depth {
			ttMove := entry.Move()
			if eval, ok := entry.GetScore(depth, ply, alpha, beta); ok && e.Board.IsPseudoLegal(ttMove) {
				e.TTable.Stats.recordCutoff()
				*line = []board.Move{ttMove}
				return eval
			}
			e.TTable.Stats.recordMoveHit()
			pvMove = ttMove
		}

		// Internal iterative reduction. Without a hash move, move ordering is weaker,
		// so reduce depth to avoid spending too much time on poorly ordered nodes.
		if pvMove == 0 && depth > 3 {
			depth--
		}

		// Null move pruning.
		// Do not prune:
		// - when in check.
		// - when less than 7 pieces on board (random heuristic) or pawn only endgame due to possible zugzwang situations
		if !isPV && !inCheck && nmp && e.Board.Occupancy[board.BOTH].Count() > 6 && !e.Board.IsPawnOnly() {
			unull := e.Board.MakeNullMove()
			R := 3 + depth/6
			e.PrevMove[ply] = 0
			value := -e.PVS(ctx, pvOrder, &[]board.Move{}, depth-R-1, ply+1, -beta, -beta+1, false, -side)
			unull()
			if value >= beta {
				return beta
			}
		}

		all := e.Board.PseudoMoveGen()
		legalMoves := 0
		e.ScoreMoves(pvMove, all, pvOrder, ply)

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
			currMove = SelectMove(all, i)
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

			// Futility pruning.
			if canFutility && depth <= 2 && legalMoves > 1 &&
				!currMove.IsCapture() && currMove.Promotion() == 0 &&
				!e.Board.InCheck &&
				staticEval+150*int16(depth) <= alpha {
				umove()
				continue
			}

			e.AddPly()
			e.PrevMove[ply] = currMove
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
					if depthR > 0 {
						e.Stability.recordLMR(true)
					}
					value = -e.PVS(ctx, pvOrder, &pv, depth-1, ply+1, -beta, -alpha, true, -side)
				} else if depthR > 0 {
					e.Stability.recordLMR(false)
				}
			}
			umove()
			e.RemovePly()

			if value > bestVal {
				bestVal = value
				bestMove = currMove
			}

			if value >= beta {
				e.MoveOrder.recordFailHigh(legalMoves == 1)
				if !currMove.IsCapture() {
					e.AddKillerMove(ply, currMove)
					e.IncrementHistory(depth, currMove)
					if ply > 0 {
						from, to := e.PrevMove[ply-1].FromTo()
						e.CounterMoves[from][to] = currMove
					}
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
		e.TTable.Store(e.Board.Hash, entryType, bestVal, depth, ply, bestMove)
		return bestVal
	}
}

func (e *Engine) Quiescence(ctx context.Context, ply int8, alpha, beta, side int16) int16 {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		e.Stats.qNodes++

		if entry, ok := e.TTable.Probe(e.Board.Hash); ok {
			if eval, ok := entry.GetScore(0, ply, alpha, beta); ok {
				return eval
			}
		}

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
		bestVal := eval
		var bestMove board.Move
		entryType := TT_UPPER

		e.ScoreMovesQ(all)
		var currMove board.Move
		for i := 0; i < len(all); i++ {
			currMove = SelectMove(all, i)

			// SEE pruning: skip losing captures when not in check.
			if !e.Board.InCheck && currMove.IsCapture() &&
				e.SEE(currMove.From(), currMove.To()) < 0 {
				continue
			}

			umove := e.Board.MakeMove(currMove)
			if e.Board.IsChecked(e.Board.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++
			value := -e.Quiescence(ctx, ply+1, -beta, -alpha, -side)
			umove()

			if value > bestVal {
				bestVal = value
				bestMove = currMove
			}

			if value > alpha {
				entryType = TT_EXACT
				alpha = value
			}

			if alpha >= beta {
				entryType = TT_LOWER
				break
			}
		}

		if legalMoves == 0 {
			return eval
		}

		e.TTable.Store(e.Board.Hash, entryType, bestVal, 0, ply, bestMove)
		return bestVal
	}
}

// Iterative deepening search. Returns best move, ponder and ok if search succeeded.
func (e *Engine) IDSearch(ctx context.Context, depth int, infinite bool) (board.Move, board.Move, bool) {
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
	e.TTable.age = 0
	e.AgeHistory()
	e.Stability.reset()
	done, ok := false, true
	wg.Add(1)
	go func() {
		for d := int8(1); d <= int8(depth); d++ {
			if done {
				wg.Done()
				return
			}

			// Don't start a new iteration if it's predicted to not finish in time.
			if d > 1 && e.TC.ShouldStop() {
				wg.Done()
				return
			}

			e.TC.IterationStarted()
			e.TTable.age = d
			e.Stats.Start()
			e.TTable.Stats.reset()
			e.MoveOrder.reset()
			var pv []board.Move
			pv = append(pv, line...)
			eval = e.PVS(ctx, pv, &line, d, 0, alpha, beta, true, color)

			if eval <= alpha || eval >= beta {
				e.Stability.recordAspiration(true)
				e.TC.AspirationFailed()
				alpha, beta = -Inf, Inf
				eval = e.PVS(ctx, pv, &line, d, 0, alpha, beta, true, color)
			} else {
				e.Stability.recordAspiration(false)
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
				e.TC.IterationFinished()
				e.Stability.recordIteration(best, eval)
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
				if s := e.TTable.Stats.String(); s != "" {
					fmt.Printf("info string %s\n", s)
				}
				if s := e.MoveOrder.String(); s != "" {
					fmt.Printf("info string %s\n", s)
				}
				if s := e.Stability.String(); s != "" {
					fmt.Printf("info string %s\n", s)
				}
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

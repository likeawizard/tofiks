package eval

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/likeawizard/tofiks/pkg/board"
)

func (e *EvalEngine) negamax(ctx context.Context, line *[]board.Move, depth, ply int, alpha, beta int, side int) int {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		if depth == 0 && !e.Board.IsChecked(e.Board.Side) {
			return e.quiescence(ctx, alpha, beta, side)
		} else if depth == 0 {
			depth++
		}

		e.Stats.nodes++

		if ply > 0 && e.IsDrawByRepetition() {
			return 0
		}

		alphaTemp := alpha
		var pvMove board.Move

		if entry, ok := e.TTable.Probe(e.Board.Hash); ok && entry.depth >= depth && ply > 0 {
			switch entry.ttType {
			case TT_EXACT:
				*line = []board.Move{entry.move}
				return entry.eval
			case TT_LOWER:
				alpha = Max(alpha, entry.eval)
			case TT_UPPER:
				beta = Min(beta, entry.eval)
			}

			pvMove = entry.move

			if alpha >= beta {
				*line = []board.Move{entry.move}
				return entry.eval
			}
		}

		all := e.Board.PseudoMoveGen()
		legalMoves := 0
		e.OrderMoves(pvMove, &all, ply)

		value := -math.MaxInt
		pv := []board.Move{}
		for i := 0; i < len(all); i++ {
			umove := e.Board.MakeMove(all[i])
			if e.Board.IsChecked(e.Board.Side ^ 1) {
				umove()
				continue
			}
			legalMoves++
			value = Max(value, -e.negamax(ctx, &pv, depth-1, ply+1, -beta, -alpha, -side))
			umove()

			if value > alpha {
				alpha = value
				*line = []board.Move{all[i]}
				*line = append(*line, pv...)
			}

			if alpha >= beta {
				e.AddKillerMove(ply, all[i])
				break
			}

		}

		if legalMoves == 0 {
			if e.Board.IsChecked(e.Board.Side) {
				value = -math.MaxInt
			} else {
				value = 0
			}
		}

		if len(*line) > 0 {
			var entryType ttType
			if value <= alphaTemp {
				entryType = TT_UPPER
			} else if value >= beta {
				entryType = TT_LOWER
			} else {
				entryType = TT_EXACT
			}
			e.TTable.Store(e.Board.Hash, entryType, value, depth, (*line)[0])
		}
		return value
	}
}

func (e *EvalEngine) quiescence(ctx context.Context, alpha, beta int, side int) int {
	select {
	case <-ctx.Done():
		// Meaningless return. Should never trust the result after ctx is expired
		return 0
	default:
		e.Stats.qNodes++
		eval := side * e.GetEvaluation(e.Board)

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
			all = e.Board.PseudoCaptureGen()
		}

		legalMoves := 0
		e.OrderMoves(0, &all, 0)

		value := -math.MaxInt
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
func (e *EvalEngine) IDSearch(ctx context.Context, depth int, pv *[]board.Move, silent bool) (board.Move, board.Move, bool) {
	var wg sync.WaitGroup
	var best, ponder board.Move
	var eval int
	var line, bestLine []board.Move
	color := 1
	alpha, beta := -math.MaxInt, math.MaxInt
	if e.Board.Side != board.WHITE {
		color = -color
	}
	done, ok := false, true
	wg.Add(1)
	go func() {
		for d := 1; d <= depth; d++ {
			if done {
				wg.Done()
				return
			}

			e.Stats.Start()
			eval = e.negamax(ctx, &line, d, 0, alpha, beta, color)

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
					bestLine = line
				}
				if !silent {
					evalStr := ""
					switch eval {
					case math.MaxInt:
						evalStr = fmt.Sprintf("#%d", 1+len(line)/2)
					case -math.MaxInt:
						evalStr = fmt.Sprintf("#%d", 1+len(line)/2)
					default:
						evalStr = fmt.Sprintf("%2.2f", float32(color*eval)/100)
					}

					fmt.Printf("Depth: %d (%s) Move: %v (%s)\n", d, evalStr, line, e.Stats.String())
				}

				//found mate stop
				if eval == math.MaxInt || eval == -math.MaxInt {
					done = true
				}
			}
		}
		wg.Done()
	}()

	wg.Wait()
	*pv = bestLine
	return best, ponder, ok
}

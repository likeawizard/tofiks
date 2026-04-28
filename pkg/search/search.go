package search

import (
	"fmt"
	"strings"
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

func (e *Engine) PVS(pvOrder []board.Move, line *[]board.Move, depth, ply int, alpha, beta int16, nmp bool, side int16) int16 {
	if e.TC.ShouldAbort() {
		// Meaningless return. Should never trust the result after abort.
		return 0
	}

	isPV := beta-alpha != 1
	inCheck := e.Board.InCheck

	if ply > e.Stats.SelDepth {
		e.Stats.SelDepth = ply
	}

	if e.Board.InCheck {
		depth++
	}

	if ply > 0 && (e.Board.HalfMoveCounter >= 100 || e.Board.InsufficientMaterial() || e.IsDrawByRepetition()) {
		return 0
	}

	// If search depth is reached and not in check enter Qsearch
	if depth <= 0 {
		return e.Quiescence(ply, alpha, beta, side)
	}

	e.Stats.nodes++

	// Static eval for pruning decisions.
	var staticEval int16
	canPrune := !isPV && !inCheck && beta > -CheckmateThreshold && beta < CheckmateThreshold
	if canPrune {
		e.Stats.evals++
		staticEval = side * int16(e.Eval.GetEvaluation(e.Board))
		e.StaticEvals[ply] = staticEval

		// Reverse futility pruning. If static eval is well above beta at shallow depths,
		// the opponent is unlikely to improve their position enough to drop below beta.
		if depth <= 5 {
			cutoff := staticEval-90*int16(depth) >= beta
			e.Prune.recordRFP(cutoff)
			if cutoff {
				return staticEval
			}
		}
	}

	// Improving: is our static eval better than 2 plies ago?
	improving := canPrune && ply >= 2 && staticEval > e.StaticEvals[ply-2]

	var pvMove board.Move
	var ttValue int16
	var ttDepth int
	var ttBound EntryType
	ttHit := false
	if entry, ok := e.TTable.Probe(e.Board.Hash); ok {
		ttMove := entry.Move()
		ttValue = entry.Score()
		ttDepth = entry.Depth()
		ttBound = entry.Type()
		ttHit = true

		// Adjust mate scores for current ply.
		if ttValue > CheckmateThreshold {
			ttValue -= int16(ply)
		} else if ttValue < -CheckmateThreshold {
			ttValue += int16(ply)
		}

		// Skip the TT cutoff during singular extension verification at this
		// ply — the excluded move is typically the TT move, so the stored
		// score would short-circuit the verification and make SE a no-op.
		if ply > 0 && ttDepth >= depth && e.ExcludedMove[ply] == 0 {
			if eval, ok := entry.GetScore(depth, ply, alpha, beta); ok && e.Board.IsPseudoLegal(ttMove) {
				e.TTable.Stats.recordCutoff(ttBound, ttValue, alpha, beta)
				*line = []board.Move{ttMove}
				return eval
			}
		}
		if e.Board.IsPseudoLegal(ttMove) {
			e.TTable.Stats.recordMoveHit()
			pvMove = ttMove
		}
	}

	// Internal iterative reduction. Without a hash move, move ordering is weaker,
	// so reduce depth to avoid spending too much time on poorly ordered nodes.
	if pvMove == 0 && depth > 3 {
		e.Prune.recordIIR()
		depth--
	}

	// Null move pruning.
	// Do not prune:
	// - when in check.
	// - when less than 7 pieces on board (random heuristic) or pawn only endgame due to possible zugzwang situations
	if !isPV && !inCheck && nmp && e.Board.Occupancy[board.Both].Count() > 6 && !e.Board.IsPawnOnly() {
		unull := e.Board.MakeNullMove()
		R := 3 + depth/7
		e.PrevMove[ply] = 0
		value := -e.PVS(pvOrder, &[]board.Move{}, depth-R-1, ply+1, -beta, -beta+1, false, -side)
		unull()
		e.Prune.recordNMP(value >= beta)
		if value >= beta {
			return beta
		}
	}

	// Singular extension: check if the TT move is significantly better than all alternatives.
	singularExtension := 0
	if ply > 0 && depth >= 8 && pvMove != 0 && e.ExcludedMove[ply] == 0 &&
		ttHit && ttDepth >= depth-3 && (ttBound == Lower || ttBound == Exact) &&
		ttValue > -CheckmateThreshold && ttValue < CheckmateThreshold {
		singularBeta := ttValue - 2*int16(depth)
		singularDepth := depth / 2

		e.ExcludedMove[ply] = pvMove
		seValue := e.PVS(pvOrder, &[]board.Move{}, singularDepth, ply, singularBeta-1, singularBeta, true, side)
		e.ExcludedMove[ply] = 0

		applied := seValue < singularBeta
		multiCut := !applied && seValue >= beta
		e.Prune.recordSE(applied, multiCut)
		if applied {
			singularExtension = 1
		} else if multiCut {
			// Multi-cut: even without the TT move, the position fails high.
			return beta
		}
	}

	all := e.Board.PseudoMoveGen()
	legalMoves := 0
	e.ScoreMoves(pvMove, all, pvOrder, ply)

	value := int16(0)
	entryType := Upper
	bestVal := -Inf
	var currMove, bestMove board.Move
	var pv []board.Move
	moveCount := len(all)
	if moveCount > 0 {
		bestMove = all[0]
	}

	for i := range moveCount {
		currMove = SelectMove(all, i)
		// Skip the excluded move during singular extension verification search.
		if currMove == e.ExcludedMove[ply] {
			continue
		}
		umove := e.Board.MakeMove(currMove)
		if e.Board.IsChecked(e.Board.Side ^ 1) {
			umove()
			continue
		}
		legalMoves++

		// Late move pruning. At shallow depths, skip quiet moves that are ordered late.
		lmpThreshold := (5 + 3*depth*depth) / (2 - boolToInt(improving))
		if canPrune && depth >= 2 && depth <= 6 &&
			!currMove.IsCapture() && currMove.Promotion() == 0 &&
			bestVal > -CheckmateThreshold {
			prune := legalMoves > lmpThreshold
			e.Prune.recordLMP(prune)
			if prune {
				umove()
				continue
			}
		}

		// Futility pruning.
		if canPrune && depth <= 2 && legalMoves > 1 &&
			!currMove.IsCapture() && currMove.Promotion() == 0 &&
			!e.Board.InCheck {
			prune := staticEval+154*int16(depth) <= alpha
			e.Prune.recordFP(prune)
			if prune {
				umove()
				continue
			}
		}

		// Apply singular extension to the TT move.
		ext := 0
		if currMove == pvMove && singularExtension > 0 {
			ext = singularExtension
		}

		e.AddPly()
		e.PrevMove[ply] = currMove
		pv = []board.Move{}
		if legalMoves == 1 {
			value = -e.PVS(pvOrder, &pv, depth-1+ext, ply+1, -beta, -alpha, true, -side)
		} else {
			depthR := 0
			if !isPV && legalMoves > 4 && !inCheck && depth > 2 &&
				currMove.Promotion() == 0 && !currMove.IsEnPassant() && !currMove.IsCapture() {
				depthR = lmrReduction(depth, legalMoves)
			}

			value = -e.PVS(pvOrder, &pv, depth-1-depthR+ext, ply+1, -(alpha + 1), -alpha, true, -side)

			if depthR > 0 {
				e.Stability.recordLMR(value > alpha)
				// Verify reduced fail-high / fail-soft result at full depth,
				if value > alpha {
					value = -e.PVS(pvOrder, &pv, depth-1+ext, ply+1, -(alpha + 1), -alpha, true, -side)
				}
			}

			if value > alpha && value < beta {
				e.MoveOrder.recordPVSReSearch()
				value = -e.PVS(pvOrder, &pv, depth-1+ext, ply+1, -beta, -alpha, true, -side)
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

			entryType = Lower
			break
		}
		e.DecrementHistory(currMove)

		if value > alpha {
			entryType = Exact
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

func (e *Engine) Quiescence(ply int, alpha, beta, side int16) int16 {
	if e.TC.ShouldAbort() {
		// Meaningless return. Should never trust the result after abort.
		return 0
	}
	e.Stats.qNodes++

	if ply > e.Stats.SelDepth {
		e.Stats.SelDepth = ply
	}

	if entry, ok := e.TTable.Probe(e.Board.Hash); ok {
		if eval, ok := entry.GetScore(0, ply, alpha, beta); ok {
			return eval
		}
	}

	e.Stats.evals++
	eval := side * int16(e.Eval.GetEvaluation(e.Board))

	if !e.Board.InCheck && eval >= beta {
		return eval
	}

	if !e.Board.InCheck && eval < alpha-975 {
		return eval
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
	entryType := Upper

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
		value := -e.Quiescence(ply+1, -beta, -alpha, -side)
		umove()

		if value > bestVal {
			bestVal = value
			bestMove = currMove
		}

		if value > alpha {
			entryType = Exact
			alpha = value
		}

		if alpha >= beta {
			entryType = Lower
			break
		}
	}

	if legalMoves == 0 {
		return eval
	}

	e.TTable.Store(e.Board.Hash, entryType, bestVal, 0, ply, bestMove)
	return bestVal
}

// Iterative deepening search. Returns best move, ponder and ok if search succeeded.
func (e *Engine) IDSearch(depth int, infinite bool) (board.Move, board.Move, bool) {
	e.MateFound = false
	var wg sync.WaitGroup
	var best, ponder board.Move
	var eval int16
	var line []board.Move
	start := time.Now()
	color := int16(1)
	alpha, beta := -Inf, Inf
	if e.Board.Side != board.White {
		color = -color
	}
	e.TTable.age = 0
	e.AgeHistory()
	e.Stability.reset()

	// Quick fix - if search fails to return a valid move - play the first legal move.
	for _, m := range e.Board.PseudoMoveGen() {
		umove := e.Board.MakeMove(m)
		legal := !e.Board.IsChecked(e.Board.Side ^ 1)
		umove()
		if legal {
			best = m
			break
		}
	}

	done, ok := false, true
	wg.Go(func() {
		for d := 1; d <= depth; d++ {
			if done {
				return
			}

			// Don't start a new iteration if it's predicted to not finish in time.
			if d > 1 && e.TC.ShouldStop() {
				return
			}

			e.TC.IterationStarted()
			e.TTable.age = int8(d)
			e.Stats.Start()
			e.TTable.Stats.reset()
			e.Eval.PawnTable.Stats.Reset()
			e.MoveOrder.reset()
			e.Prune.reset()
			var pv []board.Move
			pv = append(pv, line...)
			eval = e.PVS(pv, &line, d, 0, alpha, beta, true, color)

			if eval <= alpha || eval >= beta {
				e.Stability.recordAspiration(true)
				e.TC.AspirationFailed()
				alpha, beta = -Inf, Inf
				nodesBefore := uint64(e.Stats.nodes + e.Stats.qNodes)
				eval = e.PVS(pv, &line, d, 0, alpha, beta, true, color)
				e.Stability.recordAspirationReSearch(uint64(e.Stats.nodes+e.Stats.qNodes) - nodesBefore)
			} else {
				e.Stability.recordAspiration(false)
			}
			alpha, beta = eval-50, eval+100

			if e.TC.ShouldAbort() {
				// Search was aborted; results unreliable.
				done = true
				return
			}
			if len(line) == 0 {
				done, ok = true, false
				continue
			}
			best = line[0]
			if len(line) > 1 {
				ponder = line[1]
			}
			e.TC.IterationFinished()
			e.TC.RecordIteration(best, eval)
			e.Stability.recordIteration(best, eval)
			var lineStr strings.Builder
			for _, m := range line {
				lineStr.WriteString(" " + m.String())
			}
			totalN := e.Stats.nodes + e.Stats.qNodes
			timeSince := time.Since(start)
			nps := int64(totalN)
			if timeSince.Milliseconds() != 0 {
				nps = (1000 * nps) / timeSince.Milliseconds()
			}
			fmt.Printf("info depth %d seldepth %d score %s nodes %d nps %d time %d hashfull %d pv%s\n", d, e.Stats.SelDepth, e.ConvertEvalToScore(eval), totalN, nps, timeSince.Milliseconds(), e.TTable.Hashfull(), lineStr.String())
			if s := e.TTable.Stats.String(); s != "" {
				fmt.Printf("info string %s\n", s)
			}
			if s := e.MoveOrder.String(); s != "" {
				fmt.Printf("info string %s\n", s)
			}
			if s := e.Prune.String(); s != "" {
				fmt.Printf("info string %s\n", s)
			}
			if s := e.Stability.String(); s != "" {
				fmt.Printf("info string %s\n", s)
			}
			if s := e.Eval.PawnTable.Stats.String(); s != "" {
				fmt.Printf("info string %s\n", s)
			}
			if eval > CheckmateThreshold || eval < -CheckmateThreshold {
				e.MateFound = true
			}
			if !infinite && e.MateFound {
				done = true
			}
		}
	})

	wg.Wait()
	return best, ponder, ok
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

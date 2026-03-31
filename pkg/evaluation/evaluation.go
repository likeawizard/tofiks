package eval

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/book"
)

// Move ordering score tiers (10-bit range 0-1023, embedded in Move bits 22-31).
const (
	scorePV      = 1023
	scoreHash    = 1022
	scoreCap     = 512 // + mvvlva (5-50)
	scoreKiller0 = 511
	scoreKiller1 = 510
	scoreCounter = 509
	scoreHistMax = 508
)

type HistoryHeuristic [2][64][64]int

type Engine struct {
	MoveOrder    MoveOrderStats
	Stability    SearchStability
	Board        *board.Board
	TTable       *TTable
	PawnTable    *PawnTable
	Stop         context.CancelFunc
	Stats        Stats
	TC           TimeControl
	History      HistoryHeuristic
	Plys         [512]uint64
	Clock        Clock
	WG           sync.WaitGroup
	Ply          int
	SearchDepth  int
	CounterMoves [64][64]board.Move
	KillerMoves  [100][2]board.Move
	PrevMove     [100]board.Move
	MateFound    bool
	OwnBook      bool
	Ponder       bool
}

var mvvlva = [7][6]int{
	{10, 9, 8, 7, 6, 5},      // pawn victim.
	{30, 29, 28, 27, 26, 25}, // bishop victim.
	{20, 19, 18, 17, 16, 15}, // knight victim.
	{40, 39, 38, 37, 36, 35}, // rook victim.
	{50, 49, 48, 47, 46, 45}, // queen victim.
	{0, 0, 0, 0, 0, 0},       // king victim.
	{0, 0, 0, 0, 0, 0},       // no piece.
}

func NewEvalEngine() *Engine {
	return &Engine{
		Board:     board.NewBoard(board.StartPos),
		TTable:    NewTTable(64),
		PawnTable: NewPawnTable(),
	}
}

// Returns the best move and best opponent response - ponder.
func (e *Engine) GetMove(ctx context.Context, depth int, infinite bool) (board.Move, board.Move) {
	var best, ponder board.Move
	if e.OwnBook && book.InBook(e.Board) {
		move := book.GetWeighted(e.Board)
		return move, 0
	}

	e.TC = e.Clock.NewTimeControl(int(e.Board.FullMoveCounter), e.Board.Side)
	best, ponder, _ = e.IDSearch(ctx, depth, infinite)

	return best, ponder
}

func (e *Engine) AddKillerMove(ply int8, move board.Move) {
	if move != e.KillerMoves[ply][0] {
		e.KillerMoves[ply][1] = e.KillerMoves[ply][0]
		e.KillerMoves[ply][0] = move
	}
}

func (e *Engine) IncrementHistory(depth int8, move board.Move) {
	d := int(depth)
	from, to := move.FromTo()
	e.History[e.Board.Side][from][to] += d * d
}

func (e *Engine) DecrementHistory(move board.Move) {
	if move.IsCapture() {
		from, to := move.FromTo()
		if e.History[e.Board.Side][from][to] > 0 {
			e.History[e.Board.Side][from][to]--
		}
	}
}

func (e *Engine) GetHistory(move board.Move) int {
	from, to := move.FromTo()
	return e.History[e.Board.Side][from][to]
}

func (e *Engine) AgeHistory() {
	for side := 0; side <= 1; side++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {
				e.History[side][from][to] /= 2
			}
		}
	}
}

func (e *Engine) AddPly() {
	e.Plys[e.Ply] = e.Board.Hash
	e.Ply++
}

func (e *Engine) RemovePly() {
	e.Ply--
}

// IsDrawByRepetition checks if the current position has been seen before.
func (e *Engine) IsDrawByRepetition() bool {
	// e.Ply is the index the next move should be stored at
	// Ply - 1 is the current position
	// So start checking at Ply - 3 skipping opponent's move
	// history depth: the halfmove counter is reset on pawn moves and captures and increased otherwise
	// no equal position can be found beyond this point.
	historyDepth := max(0, e.Ply-2-int(e.Board.HalfMoveCounter))
	for ply := e.Ply - 3; ply >= historyDepth; ply -= 2 {
		if e.Board.Hash == e.Plys[ply] {
			return true
		}
	}

	return false
}

// ScoreMoves embeds move ordering scores into the unused bits of each move.
// Order: 1. PV 2. hash move 3. Captures by MVVLVA 4. killer moves 5. countermove 6. History Heuristic.
func (e *Engine) ScoreMoves(hashMove board.Move, moves, pvOrder []board.Move, ply int8) {
	lenPV := int8(len(pvOrder))
	var counterMove board.Move
	if ply > 0 {
		from, to := e.PrevMove[ply-1].FromTo()
		counterMove = e.CounterMoves[from][to]
	}
	for i := range moves {
		switch {
		case lenPV > ply && pvOrder[ply] == moves[i]:
			moves[i] = moves[i].SetScore(scorePV)
		case moves[i] == hashMove:
			moves[i] = moves[i].SetScore(scoreHash)
		case moves[i].IsCapture():
			moves[i] = moves[i].SetScore(scoreCap + e.MvvLva(moves[i]))
		case moves[i] == e.KillerMoves[ply][0]:
			moves[i] = moves[i].SetScore(scoreKiller0)
		case moves[i] == e.KillerMoves[ply][1]:
			moves[i] = moves[i].SetScore(scoreKiller1)
		case counterMove != 0 && moves[i] == counterMove:
			moves[i] = moves[i].SetScore(scoreCounter)
		default:
			moves[i] = moves[i].SetScore(min(e.GetHistory(moves[i]), scoreHistMax))
		}
	}
}

// ScoreMovesQ embeds MVVLVA scores into capture moves for quiescence ordering.
func (e *Engine) ScoreMovesQ(moves []board.Move) {
	for i := range moves {
		moves[i] = moves[i].SetScore(e.MvvLva(moves[i]))
	}
}

// SelectMove performs one step of selection sort: finds the highest-scored move
// from index k onwards and swaps it into position k.
func SelectMove(moves []board.Move, k int) board.Move {
	maxIndex := k
	for i := k + 1; i < len(moves); i++ {
		if moves[i] > moves[maxIndex] {
			maxIndex = i
		}
	}
	moves[k], moves[maxIndex] = moves[maxIndex], moves[k]
	return moves[k].ClearScore()
}

// Estimate the potential strength of the move for move ordering.
func (e *Engine) MvvLva(move board.Move) int {
	return mvvlva[e.Board.PieceAtSquare(move.To())][move.Piece()]
}

// PlayMovesUCI plays a list of moves in UCI format and updates game history.
func (e *Engine) PlayMovesUCI(uciMoves string) bool {
	moveSlice := strings.Fields(uciMoves)
	e.Ply = 0

	for _, uciMove := range moveSlice {
		_, ok := e.Board.MoveUCI(uciMove)
		if !ok {
			return false
		}
		e.AddPly()
	}

	return true
}

func (e *Engine) ReportMove(move, ponder board.Move, allowPonder bool) {
	if !allowPonder || ponder == 0 {
		fmt.Printf("bestmove %v\n", move)
	} else {
		fmt.Printf("bestmove %v ponder %v\n", move, ponder)
	}
}

// Display centipawn score. If the eval is in the checkmate score threshold convert to mate score.
func (e *Engine) ConvertEvalToScore(eval int16) string {
	if eval < -CheckmateThreshold {
		mateDist := -CheckmateScore - eval
		mateDist = mateDist/2 + mateDist%2
		return fmt.Sprintf("mate %d", mateDist)
	}

	if eval > CheckmateThreshold {
		mateDist := CheckmateScore - eval
		mateDist = mateDist/2 + mateDist%2
		return fmt.Sprintf("mate %d", mateDist)
	}

	return fmt.Sprintf("cp %d", eval)
}

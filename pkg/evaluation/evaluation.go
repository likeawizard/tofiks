package eval

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/book"
)

type PickBookMove func(*board.Board) board.Move
type HistoryHeuristic [2][64][64]int

type EvalEngine struct {
	WG          sync.WaitGroup
	Stats       EvalStats
	Board       *board.Board
	Ponder      bool
	OwnBook     bool
	MateFound   bool
	KillerMoves [100][2]board.Move
	History     HistoryHeuristic
	Ply         int
	Plys        [512]uint64
	SearchDepth int
	TTable      *TTable
	Clock       Clock
	Stop        context.CancelFunc
}

func NewEvalEngine() (*EvalEngine, error) {
	return &EvalEngine{
		Board:  board.NewBoard("startpos"),
		TTable: NewTTable(64),
	}, nil
}

// Returns the best move and best opponent response - ponder
func (e *EvalEngine) GetMove(ctx context.Context, depth int, infinite bool) (board.Move, board.Move) {
	var best, ponder board.Move
	if e.OwnBook && book.InBook(e.Board) {
		move := book.GetWeighted(e.Board)
		return move, 0
	} else {
		best, ponder, _ = e.IDSearch(ctx, depth, infinite)
	}

	return best, ponder
}

func (e *EvalEngine) AddKillerMove(ply int8, move board.Move) {
	if move != e.KillerMoves[ply][0] {
		e.KillerMoves[ply][1] = e.KillerMoves[ply][0]
		e.KillerMoves[ply][0] = move
	}
}

func (e *EvalEngine) IncrementHistory(depth int8, move board.Move) {
	d := int(depth)
	from, to := move.FromTo()
	e.History[e.Board.Side][from][to] += d * d
}

func (e *EvalEngine) DecrementHistory(move board.Move) {
	if !e.Board.IsCapture(move) {
		from, to := move.FromTo()
		if e.History[e.Board.Side][from][to] > 0 {
			e.History[e.Board.Side][from][to]--
		}
	}
}

func (e *EvalEngine) GetHistory(move board.Move) int {
	from, to := move.FromTo()
	return e.History[e.Board.Side][from][to]
}

func (e *EvalEngine) AgeHistory() {
	for from := 0; from < 64; from++ {
		for to := 0; to < 64; to++ {
			e.History[e.Board.Side][from][to] /= 2
		}
	}
}

func (e *EvalEngine) AddPly() {
	e.Plys[e.Ply] = e.Board.Hash
	e.Ply++
}

func (e *EvalEngine) RemovePly() {
	e.Ply--
}

// Draw by 3-fold repetition.
// Detect if the current position has been encountered already twice before.
func (e *EvalEngine) IsDrawByRepetition() bool {
	// e.Ply is the index the next move should be stored at
	// Ply - 1 is the current position
	// So start checking at Ply - 3 skipping opponent's move
	// history depth: the halfmove counter is reset on pawn moves and captures and increased otherwise
	// no equal position can be found beyond this point.
	historyDepth := max(0, e.Ply-2-int(e.Board.HalfMoveCounter))
	count := 0
	for ply := e.Ply - 3; ply >= historyDepth; ply -= 2 {
		if e.Board.Hash == e.Plys[ply] {
			count++
			if count > 1 {
				return true
			}
		}
	}

	return false
}

type moveScore struct {
	move  board.Move
	score int
}

var capScore = 2048

type moveSelector func(k int) board.Move

// Move ordering 1. PV 2. hash move 3. Captures orderd by MVVLVA, 4. killer moves  5. History Heuristic
func (e *EvalEngine) GetMoveSelector(hashMove board.Move, moves, pvOrder []board.Move, ply int8) moveSelector {
	moveCount := len(moves)
	scores := make([]int, moveCount)
	lenPV := int8(len(pvOrder))
	for i := range moves {
		switch {
		case lenPV > ply && pvOrder[ply] == moves[i]:
			scores[i] = capScore + 200
		case moves[i] == hashMove:
			scores[i] = capScore + 100
		case e.Board.IsCapture(moves[i]):
			scores[i] = e.MvvLva(moves[i])
		case moves[i] == e.KillerMoves[ply][0]:
			scores[i] = capScore - 5
		case moves[i] == e.KillerMoves[ply][1]:
			scores[i] = capScore - 10
		default:
			scores[i] = e.GetHistory(moves[i])
		}
	}

	return func(k int) board.Move {
		maxIndex := k
		n := len(moves)
		for i := k; i < n; i++ {
			if scores[i] > scores[maxIndex] {
				maxIndex = i
			}
		}
		scores[k], scores[maxIndex] = scores[maxIndex], scores[k]
		moves[k], moves[maxIndex] = moves[maxIndex], moves[k]
		return moves[k]
	}
}

func (e *EvalEngine) GetMoveSelectorQ(moves []board.Move) moveSelector {
	moveCount := len(moves)
	scores := make([]int, moveCount)

	for i := range moves {
		scores[i] = e.MvvLva(moves[i])
	}

	return func(k int) board.Move {
		maxIndex := k
		n := len(moves)
		for i := k; i < n; i++ {
			if scores[i] > scores[maxIndex] {
				maxIndex = i
			}
		}
		scores[k], scores[maxIndex] = scores[maxIndex], scores[k]
		moves[k], moves[maxIndex] = moves[maxIndex], moves[k]
		return moves[k]
	}
}

var mvvlva = [7][6]int{
	{10, 9, 8, 7, 6, 5},
	{30, 29, 28, 27, 26, 25},
	{20, 19, 18, 17, 16, 15},
	{40, 39, 38, 37, 36, 35},
	{50, 49, 48, 47, 46, 45},
}

// Estimate the potential strength of the move for move ordering
func (e *EvalEngine) MvvLva(move board.Move) int {
	var victim int
	attacker := e.Board.Piece(move)
	// Note: for EP captures pieceAtSquare will fail but return 0 which is still pawn
	_, _, victim = e.Board.PieceAtSquare(move.To())
	return mvvlva[victim][attacker]
}

func (e *EvalEngine) PlayMovesUCI(uciMoves string) bool {
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

func (e *EvalEngine) ReportMove(move, ponder board.Move, allowPonder bool) {
	if !allowPonder || ponder == 0 {
		fmt.Printf("bestmove %v\n", move)
	} else {
		umove := e.Board.MakeMove(move)
		defer umove()
		// TODO: cleanup, verify ponder move if legal, has returned illegal moves
		moves := e.Board.MoveGenLegal()
		for _, m := range moves {
			if m == ponder {
				fmt.Printf("bestmove %v ponder %v\n", move, ponder)
				return
			}
		}
		fmt.Printf("bestmove %v\n", move)
	}
}

// Display centipawn score. If the eval is in the checkmate score threshold convert to mate score
func (e *EvalEngine) ConvertEvalToScore(eval int32) string {
	if eval < -CheckmateThreshold {
		return fmt.Sprintf("mate %d", max(-(eval+CheckmateScore+int32(e.Board.Side^1))/2, -1))
	}

	if eval > CheckmateThreshold {
		return fmt.Sprintf("mate %d", max(-(eval-CheckmateScore-int32(e.Board.Side^1))/2, 1))
	}

	return fmt.Sprintf("cp %d", eval)
}

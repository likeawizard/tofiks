package eval

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/book"
)

type PickBookMove func(*board.Board) board.Move

type EvalEngine struct {
	WG             sync.WaitGroup
	Stats          EvalStats
	Board          *board.Board
	Ponder         bool
	OwnBook        bool
	MateFound      bool
	KillerMoves    [100][2]board.Move
	GameHistoryPly int
	GameHistory    [512]uint64
	SearchDepth    int
	TTable         *TTable
	Clock          Clock
	Stop           context.CancelFunc
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
	if !e.Board.IsCapture(move) {
		e.KillerMoves[ply][0] = e.KillerMoves[ply][1]
		e.KillerMoves[ply][1] = move
	}
}

func (e *EvalEngine) AgeKillers() {
	for i := 1; i < len(e.KillerMoves); i++ {
		e.KillerMoves[i-1] = e.KillerMoves[i]
	}
}

func (e *EvalEngine) IncrementHistory() {
	e.GameHistory[e.GameHistoryPly] = e.Board.Hash
	e.GameHistoryPly++
}

func (e *EvalEngine) DecrementHistory() {
	e.GameHistoryPly--
}

// Draw by 3-fold repetition.
// Detect if the current position has been encountered already twice before.
func (e *EvalEngine) IsDrawByRepetition() bool {
	// e.GameHistoryPly is the index the next move should be stored at
	// GameHistoryPly - 1 is the current position
	// So start checking at GameHistoryPly - 3 skipping opponent's move
	// history depth: the halfmove counter is reset on pawn moves and captures and increased otherwise
	// no equal position can be found beyond this point.
	historyDepth := Max(0, e.GameHistoryPly-2-int(e.Board.HalfMoveCounter))
	count := 0
	for ply := e.GameHistoryPly - 3; ply >= historyDepth; ply -= 2 {
		if e.Board.Hash == e.GameHistory[ply] {
			count++
			if count > 1 {
				return true
			}
		}
	}

	return false
}

func (e *EvalEngine) OrderMovesPV(pv board.Move, moves, pvOrder *[]board.Move, ply int8) {
	lenPV := int8(len(*pvOrder))
	sort.Slice(*moves, func(i int, j int) bool {
		return (lenPV > ply && (*pvOrder)[ply] == (*moves)[i]) ||
			(*moves)[i] == pv ||
			(*moves)[i] == e.KillerMoves[ply][0] ||
			(*moves)[i] == e.KillerMoves[ply][1] ||
			e.getMoveValue((*moves)[i]) > e.getMoveValue((*moves)[j])
	})
}

func (e *EvalEngine) OrderMoves(moves *[]board.Move) {
	sort.Slice(*moves, func(i int, j int) bool {
		return e.getMoveValue((*moves)[i]) > e.getMoveValue((*moves)[j])
	})
}

// Estimate the potential strength of the move for move ordering
func (e *EvalEngine) getMoveValue(move board.Move) (value int) {
	if e.Board.IsCapture(move) {
		var victim int
		attacker := PieceWeights[e.Board.Piece(move)]
		// Note: for EP captures pieceAtSquare will fail but return 0 which is still pawn
		_, _, victim = e.Board.PieceAtSquare(move.To())

		value = PieceWeights[victim] - attacker/2
	}

	// TODO: implement SEE or MVV-LVA ordering
	// Calculate the relative value of exchange
	// from, to := move.FromTo()
	// us, them := PieceWeights[b.Coords[from]], PieceWeights[b.Coords[to]]
	// if them == 0 {
	// 	value += 0
	// } else {
	// 	value += dir * (0.5*us + them)
	// }

	// Prioritize promotions
	if move.Promotion() != 0 {
		value += 3
	}

	return
}

func (e *EvalEngine) PlayMovesUCI(uciMoves string) bool {
	moveSlice := strings.Fields(uciMoves)
	e.GameHistoryPly = 0

	for _, uciMove := range moveSlice {
		_, ok := e.Board.MoveUCI(uciMove)
		if !ok {
			return false
		}
		e.IncrementHistory()
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
		return fmt.Sprintf("mate %d", Min(-(eval+CheckmateScore+int32(e.Board.Side^1))/2, -1))
	}

	if eval > CheckmateThreshold {
		return fmt.Sprintf("mate %d", Max(-(eval-CheckmateScore-int32(e.Board.Side^1))/2, 1))
	}

	return fmt.Sprintf("cp %d", eval)
}

type Number interface {
	int | int16 | int32 | int64
}

// TODO: try branchless optimization
func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

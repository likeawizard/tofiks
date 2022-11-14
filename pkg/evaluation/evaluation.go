package eval

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/book"
)

type PickBookMove func(*board.Board) board.Move

type EvalEngine struct {
	Quit           bool
	MU             sync.Mutex
	Stats          EvalStats
	Board          *board.Board
	OwnBook        bool
	KillerMoves    [100][2]board.Move
	GameHistoryPly int
	GameHistory    [512]uint64
	SearchDepth    int
	TTable         *TTable
	Stop           context.CancelFunc
}

func NewEvalEngine() (*EvalEngine, error) {
	return &EvalEngine{
		Board:  board.NewBoard("startpos"),
		TTable: NewTTable(64),
	}, nil
}

// Returns the best move and best opponent response - ponder
func (e *EvalEngine) GetMove(ctx context.Context, depth int) (board.Move, board.Move) {
	var best, ponder board.Move
	var ok bool
	all := e.Board.MoveGen()
	if len(all) == 1 {
		best = all[0]
	} else if e.OwnBook && book.InBook(e.Board) {
		move := book.GetWeighted(e.Board)
		return move, 0
	} else {
		best, ponder, ok = e.IDSearch(ctx, depth)
		e.TTable.Hashfull()
		if !ok {
			best = all[0]
		}
	}

	return best, ponder
}

func (e *EvalEngine) AddKillerMove(ply int, move board.Move) {
	if !move.IsCapture() {
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

// Two-fold repetition detection. While the rules of chess require a three-fold repetition a two-fold repetition should logically lead to three-fold repetition assuming best moves were played to repeat the position once they will be played again.
// TODO: proper three-fold repetition seems tricky as threefold repetition detection in a search can be obsucred by transpositions table
func (e *EvalEngine) IsDrawByRepetition() bool {
	// e.GameHistoryPly is the index the next move should be stored at
	// GameHistoryPly - 1 is current position
	// So start checking at GameHistoryPly - 3 skipping opponent's move
	//history depth. the halfmove counter is reset on pawn moves and captures and increased otherwise
	// no equal position can be found beyond this point.
	historyDepth := Max(0, e.GameHistoryPly-2-int(e.Board.HalfMoveCounter))
	for ply := e.GameHistoryPly - 3; ply >= historyDepth; ply -= 2 {
		if e.Board.Hash == e.GameHistory[ply] {
			return true
		}
	}

	return false
}

func (e *EvalEngine) OrderMoves(pv board.Move, moves *[]board.Move, ply int) {
	sort.Slice(*moves, func(i int, j int) bool {
		return (*moves)[i] == pv ||
			(*moves)[i] == e.KillerMoves[ply][0] ||
			(*moves)[i] == e.KillerMoves[ply][1] ||
			e.getMoveValue((*moves)[i]) > e.getMoveValue((*moves)[j])
	})
}

// Estimate the potential strength of the move for move ordering
func (e *EvalEngine) getMoveValue(move board.Move) (value int) {

	if move.IsCapture() {
		attacker := PieceWeights[(move.Piece()-1)%6]
		_, _, victim := e.Board.PieceAtSquare(move.To())
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

// TODO: try branchless optimization
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

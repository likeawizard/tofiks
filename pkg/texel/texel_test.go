package texel

import (
	"math"
	"testing"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/search"
)

var benchPositions = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"1r2r1k1/pp3pbp/2n1p1p1/q1ppP1N1/2PP4/1P4PP/P2Q1PBK/R4R2 w - - 0 1",
	"r1bqkb1r/pppppppp/2n2n2/8/4P3/2N5/PPPP1PPP/R1BQKBNR w KQkq - 2 3",
}

// TestTraceMatchesEval verifies that the trace-based eval reproduces the engine eval.
func TestTraceMatchesEval(t *testing.T) {
	weights := InitialParams()

	for _, fen := range benchPositions {
		b := board.NewBoard(fen)
		e := search.NewEngine()
		e.Board = b
		e.Board.Phase = e.Board.GetGamePhase()

		engineEval := float64(e.Eval.GetEvaluation(e.Board))

		trace, _ := TraceEvaluate(b)
		traceEval := EvalFromTrace(&trace, &weights)

		// Allow small rounding differences due to integer vs float phase interpolation.
		// Larger PST values increase rounding error from int truncation in the engine.
		if math.Abs(engineEval-traceEval) > 3.0 {
			t.Errorf("FEN %s: engine=%v trace=%v diff=%v",
				fen, engineEval, traceEval, engineEval-traceEval)
		}
	}
}

// TestTraceSymmetry verifies that flipping the board negates the trace eval.
func TestTraceSymmetry(t *testing.T) {
	weights := InitialParams()

	for _, fen := range benchPositions {
		b := board.NewBoard(fen)
		trace, _ := TraceEvaluate(b)
		evalW := EvalFromTrace(&trace, &weights)

		b.Flip()
		traceFlip, _ := TraceEvaluate(b)
		evalB := EvalFromTrace(&traceFlip, &weights)

		if math.Abs(evalW+evalB) > 1.5 {
			t.Errorf("FEN %s: white=%v flipped=%v sum=%v", fen, evalW, evalB, evalW+evalB)
		}
	}
}

func TestSigmoid(t *testing.T) {
	K := 0.2109375
	// sigmoid(0) should be 0.5.
	if math.Abs(Sigmoid(K, 0)-0.5) > 1e-10 {
		t.Errorf("sigmoid(0) = %v, want 0.5", Sigmoid(K, 0))
	}
	// sigmoid should be monotonically increasing.
	if Sigmoid(K, 100) <= Sigmoid(K, 0) {
		t.Error("sigmoid not monotonically increasing")
	}
	if Sigmoid(K, 0) <= Sigmoid(K, -100) {
		t.Error("sigmoid not monotonically increasing")
	}
}

// BenchmarkTraceEvaluate measures trace computation speed per position.
func BenchmarkTraceEvaluate(b *testing.B) {
	for _, fen := range benchPositions {
		brd := board.NewBoard(fen)
		b.Run(fen[:min(30, len(fen))], func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				TraceEvaluate(brd)
			}
		})
	}
}

// BenchmarkEvalFromTrace measures dot product speed.
func BenchmarkEvalFromTrace(b *testing.B) {
	brd := board.NewBoard(benchPositions[0])
	trace, _ := TraceEvaluate(brd)
	weights := InitialParams()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvalFromTrace(&trace, &weights)
	}
}

// BenchmarkComputeGradient measures gradient computation for a small dataset.
func BenchmarkComputeGradient(b *testing.B) {
	entries := make([]Entry, len(benchPositions))
	for i, fen := range benchPositions {
		brd := board.NewBoard(fen)
		trace, phase := TraceEvaluate(brd)
		entries[i] = Entry{Trace: trace, Phase: phase, Result: 0.5}
	}
	weights := InitialParams()
	K := 0.2109375
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeGradient(entries, &weights, K)
	}
}

// BenchmarkMSE measures MSE computation speed.
func BenchmarkMSE(b *testing.B) {
	entries := make([]Entry, len(benchPositions))
	for i, fen := range benchPositions {
		brd := board.NewBoard(fen)
		trace, phase := TraceEvaluate(brd)
		entries[i] = Entry{Trace: trace, Phase: phase, Result: 0.5}
	}
	weights := InitialParams()
	K := 0.2109375
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MeanSquaredError(entries, &weights, K)
	}
}

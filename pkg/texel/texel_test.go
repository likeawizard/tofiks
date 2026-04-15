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
	// Asymmetric positions — different pawn counts per side and/or piece
	// imbalances — so the self-consistency test exercises every per-side
	// tunable without white/black contributions cancelling each other out.
	"4k3/8/8/8/8/8/PPP5/R3K2R w KQ - 0 1",           // white-only rooks + 3 pawns
	"r3k2r/8/8/8/8/8/8/4K3 w kq - 0 1",              // black-only rooks, no pawns
	"4k3/pppppppp/8/8/8/8/P7/4K3 w - - 0 1",         // 1 white pawn vs 8 black pawns
	"r3k3/1ppp4/8/8/8/8/PPPPPP2/R3K3 w Qq - 0 1",    // asymmetric rook + pawn counts
	"1n2k3/p1pppppp/8/8/8/8/PPPPPP2/N3K3 w - - 0 1", // knights, asymmetric pawns
	"4k3/8/8/3p4/4P3/8/8/4K3 w - - 0 1",             // king-and-pawn endgame
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
		// The eval does several independent integer `/256` divisions (main PST loop,
		// king safety/activity, knight/rook pawn bonus, one per passed pawn for
		// king-proximity), each contributing up to 1 cp of truncation error. The
		// trace computes everything in float so has none. 10 cp tolerance covers
		// the realistic worst case (8-passer endgames); TestTraceCoefficientsMatchEval
		// verifies correctness at the per-coefficient level where truncation is avoided.
		if math.Abs(engineEval-traceEval) > 10.0 {
			t.Errorf("FEN %s: engine=%v trace=%v diff=%v",
				fen, engineEval, traceEval, engineEval-traceEval)
		}
	}
}

// TestTraceCoefficientsMatchEval is a per-parameter self-consistency check
// between the eval function and the trace.
//
// For every parameter index, we bump the weight by delta and measure the
// resulting eval change, then compare it to (trace_coefficient * delta). If
// they disagree by more than a small rounding budget, the trace and eval have
// drifted out of sync. Missing coefficients (eval uses a param the trace
// doesn't emit), phantom coefficients (trace emits a param the eval doesn't
// use), wrong scale factors, and sign flips are all caught.
//
// Delta=2560 (10× the natural 256) is chosen so that genuine wiring bugs show
// up as errors >> the noise floor from integer truncation. The eval does
// several `/256` integer divisions per piece and Go truncates toward zero on
// negative intermediates, so a correctly-wired param can still drift by up to
// ~1 cp per application site. With delta=2560 a sign flip is ~5000 cp off and
// a scale error is ≥256 cp off, both dwarfed by any tolerance we pick.
//
// Whenever a new tunable parameter is added, this test verifies automatically
// that both sides (eval and trace) are wired up correctly.
func TestTraceCoefficientsMatchEval(t *testing.T) {
	const delta = 2560
	// Rounding budget: up to ~1 cp per application site, times delta/256.
	// Most params affect ≤4 sites (e.g., king safety terms hit both kings).
	// 40 cp gives ample headroom for truncation without masking real bugs,
	// which would be off by hundreds of cp or more.
	const tolerance = 40.0

	// Snapshot the baseline weights once and restore at the end so other tests
	// don't see mutated eval globals.
	baseline := InitialParams()
	defer ApplyParams(&baseline)

	for _, fen := range benchPositions {
		b := board.NewBoard(fen)
		e := search.NewEngine()
		e.Board = b
		e.Board.Phase = e.Board.GetGamePhase()

		// Measure the baseline eval.
		ApplyParams(&baseline)
		// Clear the pawn cache between param changes; pawn eval depends on
		// PawnProtected etc. and the cache would serve stale scores.
		e.Eval.PawnTable.Clear()
		baseEval := e.Eval.GetEvaluation(e.Board)

		// Get the sparse trace for this position.
		trace, _ := TraceEvaluate(b)
		traceMap := make(map[int]float64, len(trace))
		for _, c := range trace {
			traceMap[int(c.Index)] = float64(c.Value)
		}

		for i := 0; i < NumParams; i++ {
			perturbed := baseline
			perturbed[i] += float64(delta)
			ApplyParams(&perturbed)
			e.Eval.PawnTable.Clear()
			bumpedEval := e.Eval.GetEvaluation(e.Board)

			actualDelta := float64(bumpedEval - baseEval)
			predictedDelta := float64(delta) * traceMap[i]

			if math.Abs(actualDelta-predictedDelta) > tolerance {
				t.Errorf("FEN %s param[%d]: actual delta=%v predicted=%v (trace coef=%v)",
					fen, i, actualDelta, predictedDelta, traceMap[i])
			}
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

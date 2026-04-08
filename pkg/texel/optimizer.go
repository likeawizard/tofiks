package texel

import (
	"fmt"
	"log"
	"math"
	"time"
)

// Sigmoid maps an eval score to a [0, 1] probability.
func Sigmoid(k, x float64) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, -k*x/400.0))
}

// SigmoidPrime is the derivative of Sigmoid with respect to x.
func SigmoidPrime(k, x float64) float64 {
	s := Sigmoid(k, x)
	return s * (1.0 - s) * k * math.Log(10.0) / 400.0
}

// MeanSquaredError computes the average squared error over all entries.
func MeanSquaredError(entries []Entry, weights *[NumParams]float64, k float64) float64 {
	var total float64
	for i := range entries {
		e := EvalFromTrace(&entries[i].Trace, weights)
		diff := entries[i].Result - Sigmoid(k, e)
		total += diff * diff
	}
	return total / float64(len(entries))
}

// StreamMSE computes MSE by streaming through a cache file. O(1) memory.
func StreamMSE(cachePath string, n int, weights *[NumParams]float64, k float64) float64 {
	var total float64
	if err := ForEachEntry(cachePath, func(e *Entry) {
		ev := EvalFromTrace(&e.Trace, weights)
		diff := e.Result - Sigmoid(k, ev)
		total += diff * diff
	}); err != nil {
		log.Fatalf("StreamMSE: %v", err)
	}
	return total / float64(n)
}

// StreamGradientAndMSE computes both gradient and MSE in a single pass.
func StreamGradientAndMSE(cachePath string, n int, weights *[NumParams]float64, k float64) ([NumParams]float64, float64) {
	var grad [NumParams]float64
	var mseTotal float64
	nf := float64(n)

	if err := ForEachEntry(cachePath, func(e *Entry) {
		ev := EvalFromTrace(&e.Trace, weights)
		sig := Sigmoid(k, ev)
		sigPrime := SigmoidPrime(k, ev)

		diff := e.Result - sig
		mseTotal += diff * diff

		factor := 2.0 * (sig - e.Result) * sigPrime / nf
		for _, c := range e.Trace {
			grad[c.Index] += factor * float64(c.Value)
		}
	}); err != nil {
		log.Fatalf("StreamGradientAndMSE: %v", err)
	}

	return grad, mseTotal / nf
}

// OptimizeK finds the K constant that minimizes MSE on the given entries.
func OptimizeK(entries []Entry, weights *[NumParams]float64) float64 {
	lo, hi := 0.0, 10.0
	for step := 0; step < 100; step++ {
		m1 := lo + (hi-lo)/3.0
		m2 := hi - (hi-lo)/3.0
		e1 := MeanSquaredError(entries, weights, m1)
		e2 := MeanSquaredError(entries, weights, m2)
		if e1 < e2 {
			hi = m2
		} else {
			lo = m1
		}
	}
	return (lo + hi) / 2.0
}

// StreamOptimizeK finds optimal K by streaming through the cache file.
// 30 ternary search steps give precision to ~1e-9, more than enough.
func StreamOptimizeK(cachePath string, n int, weights *[NumParams]float64) float64 {
	lo, hi := 0.0, 10.0
	for step := 0; step < 30; step++ {
		m1 := lo + (hi-lo)/3.0
		m2 := hi - (hi-lo)/3.0
		e1 := StreamMSE(cachePath, n, weights, m1)
		e2 := StreamMSE(cachePath, n, weights, m2)
		if e1 < e2 {
			hi = m2
		} else {
			lo = m1
		}
	}
	return (lo + hi) / 2.0
}

// AdamConfig holds Adam optimizer hyperparameters.
type AdamConfig struct {
	LR      float64 // Learning rate.
	Beta1   float64 // Exponential decay rate for first moment.
	Beta2   float64 // Exponential decay rate for second moment.
	Epsilon float64 // Small constant for numerical stability.
}

// DefaultAdamConfig returns sensible defaults for texel tuning.
func DefaultAdamConfig() AdamConfig {
	return AdamConfig{
		LR:      1.0,
		Beta1:   0.9,
		Beta2:   0.999,
		Epsilon: 1e-8,
	}
}

// Optimize runs Adam gradient descent on the weight vector.
func Optimize(entries []Entry, weights *[NumParams]float64, k float64, iterations int, cfg AdamConfig) {
	var m, v [NumParams]float64 // First and second moment estimates.

	log.Printf("Starting optimization: %d params, %d entries, K=%.6f", NumParams, len(entries), k)
	log.Printf("Initial MSE: %.10f", MeanSquaredError(entries, weights, k))

	for iter := 1; iter <= iterations; iter++ {
		start := time.Now()

		// Compute analytical gradients.
		grad := ComputeGradient(entries, weights, k)

		// Adam update.
		t := float64(iter)
		for i := 0; i < NumParams; i++ {
			m[i] = cfg.Beta1*m[i] + (1.0-cfg.Beta1)*grad[i]
			v[i] = cfg.Beta2*v[i] + (1.0-cfg.Beta2)*grad[i]*grad[i]
			mHat := m[i] / (1.0 - math.Pow(cfg.Beta1, t))
			vHat := v[i] / (1.0 - math.Pow(cfg.Beta2, t))
			weights[i] -= cfg.LR * mHat / (math.Sqrt(vHat) + cfg.Epsilon)
		}

		mse := MeanSquaredError(entries, weights, k)
		elapsed := time.Since(start)
		log.Printf("Iter %d/%d  MSE: %.10f  (%v)", iter, iterations, mse, elapsed)
	}
}

// ComputeGradient computes the analytical gradient of MSE w.r.t. all weights.
// dMSE/dw_i = (2/N) * Σ (sigmoid(eval) - result) * sigmoid'(eval) * trace[i].
func ComputeGradient(entries []Entry, weights *[NumParams]float64, k float64) [NumParams]float64 {
	var grad [NumParams]float64
	n := float64(len(entries))

	for i := range entries {
		e := EvalFromTrace(&entries[i].Trace, weights)
		sig := Sigmoid(k, e)
		sigPrime := SigmoidPrime(k, e)
		factor := 2.0 * (sig - entries[i].Result) * sigPrime / n

		for _, c := range entries[i].Trace {
			grad[c.Index] += factor * float64(c.Value)
		}
	}

	return grad
}

// StreamOptimize runs Adam gradient descent by streaming from a cache file.
// Gradient and MSE are computed in a single pass per iteration.
// Memory usage is O(NumParams), independent of dataset size.
// Stops early if dMSE falls below earlyStopThreshold (0 = disabled).
func StreamOptimize(cachePath string, n int, weights *[NumParams]float64, k float64, iterations int, cfg AdamConfig, earlyStopThreshold float64) {
	var m, v [NumParams]float64

	log.Printf("Starting optimization: %d params, %d entries, K=%.6f", NumParams, n, k)

	prevMSE := StreamMSE(cachePath, n, weights, k)
	log.Printf("Initial MSE: %.10f", prevMSE)

	for iter := 1; iter <= iterations; iter++ {
		start := time.Now()

		// Single pass: compute gradient and MSE together.
		grad, mse := StreamGradientAndMSE(cachePath, n, weights, k)
		dMSE := prevMSE - mse

		// Adam update.
		t := float64(iter)
		for i := 0; i < NumParams; i++ {
			m[i] = cfg.Beta1*m[i] + (1.0-cfg.Beta1)*grad[i]
			v[i] = cfg.Beta2*v[i] + (1.0-cfg.Beta2)*grad[i]*grad[i]
			mHat := m[i] / (1.0 - math.Pow(cfg.Beta1, t))
			vHat := v[i] / (1.0 - math.Pow(cfg.Beta2, t))
			weights[i] -= cfg.LR * mHat / (math.Sqrt(vHat) + cfg.Epsilon)
		}

		// Re-center PSTs: shift mean into piece weights so PSTs stay positional.
		recenterPSTs(weights)

		elapsed := time.Since(start)
		log.Printf("Iter %d/%d  MSE: %.10f  dMSE: %+.2e  (%v)", iter, iterations, mse, dMSE, elapsed)

		if earlyStopThreshold > 0 && iter > 10 && math.Abs(dMSE) < earlyStopThreshold {
			log.Printf("Early stop: |dMSE| %.2e < threshold %.2e", math.Abs(dMSE), earlyStopThreshold)
			break
		}

		prevMSE = mse
	}
}

// PrintParams outputs the tuned weights in a format ready to paste into Go source.
func PrintParams(w *[NumParams]float64) {
	fmt.Println("// === Piece-Square Tables ===")
	pieceNames := []string{"pawn", "bishop", "knight", "rook", "queen", "king"}
	stageNames := []string{"", "EG"}
	for stage := 0; stage < 2; stage++ {
		for piece := 0; piece < 6; piece++ {
			fmt.Printf("var %s%sPST = [64]int{\n", pieceNames[piece], stageNames[stage])
			for sq := 0; sq < 64; sq++ {
				if sq%8 == 0 {
					fmt.Print("\t")
				}
				fmt.Printf("%d, ", int(math.Round(w[pstIndex(stage, piece, sq)])))
				if sq%8 == 7 {
					fmt.Println()
				}
			}
			fmt.Println("}")
			fmt.Println()
		}
	}

	fmt.Println("// === Piece Weights ===")
	fmt.Printf("PieceWeights = [6]int{%d, %d, %d, %d, %d, 10000}\n\n",
		int(math.Round(w[pieceWeightStart+0])),
		int(math.Round(w[pieceWeightStart+1])),
		int(math.Round(w[pieceWeightStart+2])),
		int(math.Round(w[pieceWeightStart+3])),
		int(math.Round(w[pieceWeightStart+4])))

	fmt.Println("// === Mobility ===")
	fmt.Printf("MOVE_QUEEN  = %d\n", int(math.Round(w[mobilityStart+0])))
	fmt.Printf("MOVE_ROOK   = %d\n", int(math.Round(w[mobilityStart+1])))
	fmt.Printf("MOVE_BISHOP = %d\n", int(math.Round(w[mobilityStart+2])))
	fmt.Printf("MOVE_KNIGHT = %d\n", int(math.Round(w[mobilityStart+3])))
	fmt.Printf("MOVE_KING   = %d\n", int(math.Round(w[mobilityStart+4])))
	fmt.Printf("W_CAPTURE   = %d\n\n", int(math.Round(w[captureStart])))

	fmt.Println("// === Threats ===")
	fmt.Printf("QUEEN_THREAT  = %d\n", int(math.Round(w[threatStart+0])))
	fmt.Printf("ROOK_THREAT   = %d\n", int(math.Round(w[threatStart+1])))
	fmt.Printf("BISHOP_THREAT = %d\n", int(math.Round(w[threatStart+2])))
	fmt.Printf("KNIGHT_THREAT = %d\n\n", int(math.Round(w[threatStart+3])))

	fmt.Println("// === Pawn Structure ===")
	fmt.Printf("W_P_PROTECTED      = %d\n", int(math.Round(w[pawnStructStart+0])))
	fmt.Printf("W_P_DOUBLED        = %d\n", int(math.Round(w[pawnStructStart+1])))
	fmt.Printf("W_P_ISOLATED       = %d\n", int(math.Round(w[pawnStructStart+2])))
	fmt.Printf("W_P_BACKWARD       = %d\n", int(math.Round(w[pawnStructStart+3])))
	fmt.Printf("W_P_BLOCKED        = %d\n", int(math.Round(w[pawnStructStart+4])))
	fmt.Printf("W_P_CONNECTED_PASS = %d\n", int(math.Round(w[pawnStructStart+5])))
	fmt.Printf("W_P_CANDIDATE      = %d\n\n", int(math.Round(w[pawnStructStart+6])))

	fmt.Println("// === Passed Pawn Bonus ===")
	fmt.Printf("PassedPawnBonus = [8]int{0, %d, %d, %d, %d, %d, %d, 0}\n\n",
		int(math.Round(w[passedPawnStart+0])),
		int(math.Round(w[passedPawnStart+1])),
		int(math.Round(w[passedPawnStart+2])),
		int(math.Round(w[passedPawnStart+3])),
		int(math.Round(w[passedPawnStart+4])),
		int(math.Round(w[passedPawnStart+5])))

	fmt.Println("// === Rook Files ===")
	fmt.Printf("W_ROOK_OPEN_FILE      = %d\n", int(math.Round(w[rookFileStart+0])))
	fmt.Printf("W_ROOK_SEMI_OPEN_FILE = %d\n\n", int(math.Round(w[rookFileStart+1])))

	fmt.Println("// === Bishop Pair ===")
	fmt.Printf("BishopPairBonus = %d\n\n", int(math.Round(w[bishopPairStart])))

	fmt.Println("// === King Safety (MG) ===")
	fmt.Printf("KS_DIST_CENTER = %d\n", int(math.Round(w[kingSafetyStart+0])))
	fmt.Printf("KS_PAWN_SHIELD = %d\n", int(math.Round(w[kingSafetyStart+1])))
	fmt.Printf("KS_FRIENDLY    = %d\n", int(math.Round(w[kingSafetyStart+2])))
	fmt.Printf("KS_MOBILITY    = %d\n\n", int(math.Round(w[kingSafetyStart+3])))

	fmt.Println("// === King Activity (EG) ===")
	fmt.Printf("KA_DIST_CENTER  = %d\n", int(math.Round(w[kingActivityStart+0])))
	fmt.Printf("KA_DIST_SQUARES = %d\n", int(math.Round(w[kingActivityStart+1])))
	fmt.Printf("KA_MOBILITY     = %d\n\n", int(math.Round(w[kingActivityStart+2])))

	fmt.Println("// === Outposts ===")
	for _, name := range []string{"knight", "bishop"} {
		offset := outpostStart
		if name == "bishop" {
			offset += 64
		}
		fmt.Printf("var %sOutposts = [64]int{\n", name)
		for sq := 0; sq < 64; sq++ {
			if sq%8 == 0 {
				fmt.Print("\t")
			}
			fmt.Printf("%d, ", int(math.Round(w[offset+sq])))
			if sq%8 == 7 {
				fmt.Println()
			}
		}
		fmt.Println("}")
		fmt.Println()
	}
}

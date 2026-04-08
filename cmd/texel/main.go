package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/likeawizard/tofiks/pkg/texel"
)

func main() {
	var (
		file         string
		limit        int
		iterations   int
		workers      int
		earlyStopStr string
		earlyStop    float64
	)
	flag.StringVar(&file, "f", "texel_data.txt", "Training data file (format: result fen)")
	flag.IntVar(&limit, "lim", 0, "Max positions to load (0 = all)")
	flag.IntVar(&iterations, "i", 200, "Max optimization iterations")
	flag.IntVar(&workers, "c", runtime.NumCPU(), "Worker goroutines for cache building")
	flag.StringVar(&earlyStopStr, "es", "1e-7", "Early stop threshold for dMSE (0 = disabled)")
	flag.Parse()

	if _, err := fmt.Sscanf(earlyStopStr, "%e", &earlyStop); err != nil {
		log.Fatalf("Invalid early stop threshold %q: %v", earlyStopStr, err)
	}

	// Build binary cache from text data (or reuse existing).
	cachePath := strings.TrimSuffix(file, ".txt") + ".bin"
	if _, err := os.Stat(cachePath); err != nil {
		log.Printf("Building cache from %s (limit=%d, workers=%d)", file, limit, workers)
		n, err := texel.BuildCache(file, cachePath, limit, workers)
		if err != nil {
			log.Fatalf("Failed to build cache: %v", err)
		}
		log.Printf("Cached %d positions to %s", n, cachePath)
	} else {
		log.Printf("Using existing cache: %s", cachePath)
	}

	n, err := texel.CountEntries(cachePath)
	if err != nil {
		log.Fatalf("Failed to read cache: %v", err)
	}
	log.Printf("Cache contains %d positions", n)

	weights := texel.InitialParams()

	log.Println("Optimizing K...")
	K := texel.StreamOptimizeK(cachePath, n, &weights)
	log.Printf("Optimal K: %.6f", K)

	cfg := texel.DefaultAdamConfig()
	texel.StreamOptimize(cachePath, n, &weights, K, iterations, cfg, earlyStop)

	log.Println("=== Tuned Parameters ===")
	texel.PrintParams(&weights)
}

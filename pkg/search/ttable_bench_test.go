package search

import (
	"math/rand/v2"
	"testing"
)

const benchTTSize = 16

func generateHashes(n int) []uint64 {
	r := rand.New(rand.NewPCG(1, 2))
	hashes := make([]uint64, n)
	for i := range hashes {
		hashes[i] = r.Uint64()
	}
	return hashes
}

func BenchmarkTTStore(b *testing.B) {
	tt := NewTTable(benchTTSize)
	hashes := generateHashes(b.N)

	b.ResetTimer()
	for i := range b.N {
		tt.Store(hashes[i], TT_EXACT, 100, 10, 0, 0)
	}
}

func BenchmarkTTProbeHit(b *testing.B) {
	tt := NewTTable(benchTTSize)
	hashes := generateHashes(b.N)

	for i := range b.N {
		tt.Store(hashes[i], TT_EXACT, 100, 10, 0, 0)
	}

	b.ResetTimer()
	for i := range b.N {
		tt.Probe(hashes[i])
	}
}

func BenchmarkTTProbeMiss(b *testing.B) {
	tt := NewTTable(benchTTSize)
	// Store one set of hashes, probe with a different set.
	stored := generateHashes(b.N)
	for i := range b.N {
		tt.Store(stored[i], TT_EXACT, 100, 10, 0, 0)
	}

	r := rand.New(rand.NewPCG(3, 4))
	probes := make([]uint64, b.N)
	for i := range probes {
		probes[i] = r.Uint64()
	}

	b.ResetTimer()
	for i := range b.N {
		tt.Probe(probes[i])
	}
}

func BenchmarkTTStoreMixed(b *testing.B) {
	tt := NewTTable(benchTTSize)
	hashes := generateHashes(b.N)
	depths := []int{0, 2, 5, 8, 12}
	types := []EntryType{TT_UPPER, TT_LOWER, TT_EXACT}

	b.ResetTimer()
	for i := range b.N {
		tt.Store(hashes[i], types[i%len(types)], int16(i%500), depths[i%len(depths)], 0, 0)
	}
}

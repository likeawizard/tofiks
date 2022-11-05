package eval

import (
	"fmt"

	"github.com/likeawizard/tofiks/internal/board"
)

type ttType uint8

const (
	TT_UPPER ttType = iota
	TT_LOWER
	TT_EXACT
)

type TTable struct {
	entries []SearchEntry
	size    uint64
}

type SearchEntry struct {
	hash   uint64
	ttType ttType
	eval   int
	depth  int
	move   board.Move
}

func NewTTable(sizeInMb int) *TTable {
	size := (1024 * 1024 * sizeInMb) / 40
	fmt.Println("TT Size:", size)
	return &TTable{
		entries: make([]SearchEntry, size),
		size:    uint64(size),
	}
}

func (tt *TTable) Probe(hash uint64) (*SearchEntry, bool) {
	idx := hash % tt.size
	if hash == tt.entries[idx].hash {
		return &tt.entries[idx], true
	}
	return nil, false
}

func (tt *TTable) Store(hash uint64, entryType ttType, eval, depth int, move board.Move) {
	idx := hash % tt.size
	tt.entries[idx] = SearchEntry{
		hash:   hash,
		ttType: entryType,
		eval:   eval,
		depth:  depth,
		move:   move,
	}
}

func (tt *TTable) Clear() {
	tt.entries = make([]SearchEntry, tt.size)
}

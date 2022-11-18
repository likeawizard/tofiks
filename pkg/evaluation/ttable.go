package eval

import (
	"github.com/likeawizard/tofiks/pkg/board"
)

type ttType uint8

const (
	TT_UPPER ttType = iota
	TT_LOWER
	TT_EXACT
)

type TTable struct {
	entries  []SearchEntry
	hashfull uint64
	size     uint64
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

func (tt *TTable) Hashfull() uint64 {
	tt.hashfull = 0
	for _, e := range tt.entries {
		if e.hash != 0 {
			tt.hashfull++
		}
	}
	tt.hashfull = (tt.hashfull * 1000) / tt.size
	return tt.hashfull
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
	tt.hashfull = 0
	tt.entries = make([]SearchEntry, tt.size)
}

func (se *SearchEntry) GetScore(depth, ply, alpha, beta int) (int, bool) {
	eval := se.eval

	if eval > CheckmateThreshold {
		eval -= ply
	}

	if eval < -CheckmateThreshold {
		eval += ply
	}

	switch {
	case se.ttType == TT_EXACT:
		return eval, true
	case se.ttType == TT_UPPER && eval <= alpha:
		return eval, true
	case se.ttType == TT_LOWER && eval >= beta:
		return eval, true
	}

	return eval, false
}

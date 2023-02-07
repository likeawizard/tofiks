package eval

import (
	"unsafe"

	"github.com/likeawizard/tofiks/pkg/board"
)

type ttType uint8

const (
	TT_UPPER ttType = iota
	TT_LOWER
	TT_EXACT
)

func (ttt ttType) String() string {
	switch ttt {
	case TT_EXACT:
		return "exact"
	case TT_UPPER:
		return "upper"
	case TT_LOWER:
		return "lower"
	default:
		return "none"
	}
}

type TTable struct {
	entries       []SearchEntry
	newWrite      uint64
	overWrite     uint64
	rejectedWrite uint64
	size          uint64
}

// LSB 0..15 move, 16..23 depth 24..31 type 32..63 score MSB
type EntryData uint64

const (
	move_mask   = (1 << 16) - 1
	type_mask   = (1 << 8) - 1
	depth_mask  = type_mask
	score_mask  = (1 << 32) - 1
	depth_shift = 16
	type_shift  = 24
	score_shift = 32
)

func NewEntry(move board.Move, depth int8, eType ttType, score int32) EntryData {
	return EntryData(move) |
		EntryData(depth)<<depth_shift |
		EntryData(eType)<<type_shift |
		EntryData(score)<<score_shift
}

func (ed EntryData) Get() (board.Move, int8, ttType, int32) {
	return board.Move(ed & move_mask),
		int8((ed >> depth_shift) & type_mask),
		ttType((ed >> type_shift) & type_mask),
		int32(ed >> score_shift)
}

func (ed EntryData) Depth() int8 {
	return int8((ed >> depth_shift) & depth_mask)
}

func (ed EntryData) Move() board.Move {
	return board.Move(ed & move_mask)
}

func (ed EntryData) Score() int32 {
	return int32(ed >> score_shift)
}

func (ed EntryData) Type() ttType {
	return ttType((ed >> type_shift) & type_mask)
}

type SearchEntry struct {
	key  uint64
	data EntryData
}

func NewTTable(sizeInMb int) *TTable {
	eSize := int(unsafe.Sizeof(SearchEntry{}))
	size := (1024 * 1024 * sizeInMb) / eSize
	return &TTable{
		entries: make([]SearchEntry, size),
		size:    uint64(size),
	}
}

func (tt *TTable) Probe(hash uint64) (*EntryData, bool) {
	idx := hash % tt.size
	entry := tt.entries[idx]
	if entry.key^uint64(entry.data) == hash {
		return &entry.data, true
	}
	return nil, false
}

func (tt *TTable) Hashfull() uint64 {
	return (tt.newWrite * 1000) / tt.size
}

func (tt *TTable) Store(hash uint64, entryType ttType, eval int32, depth int8, move board.Move) {
	idx := hash % tt.size
	data := NewEntry(move, depth, entryType, eval)
	if tt.entries[idx].data == 0 {
		tt.entries[idx] = SearchEntry{
			key:  hash ^ uint64(data),
			data: data,
		}
		tt.newWrite++
		return
	} else if entryType == TT_EXACT || hash^uint64(data) != tt.entries[idx].key || tt.entries[idx].data.Depth() < depth {
		// Replace entry for new position or same position with greater depth
		tt.entries[idx] = SearchEntry{
			key:  hash ^ uint64(data),
			data: data,
		}
		tt.overWrite++
		return
	}

	tt.rejectedWrite++
}

func (tt *TTable) Clear() {
	tt.newWrite = 0
	tt.overWrite = 0
	tt.rejectedWrite = 0
	for i := 0; i < int(tt.size); i++ {
		tt.entries[i].key = 0
		tt.entries[i].data = 0
	}
}

func (ed *EntryData) GetScore(depth, ply int8, alpha, beta int32) (int32, bool) {
	ttType, eval := ed.Type(), ed.Score()

	if eval > CheckmateThreshold {
		eval -= int32(ply)
	}

	if eval < -CheckmateThreshold {
		eval += int32(ply)
	}

	switch {
	case ttType == TT_EXACT:
		return eval, true
	case ttType == TT_UPPER && eval <= alpha:
		return alpha, true
	case ttType == TT_LOWER && eval >= beta:
		return beta, true
	}

	return eval, false
}

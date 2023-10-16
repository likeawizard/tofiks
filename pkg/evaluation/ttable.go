package eval

import (
	"unsafe"

	"github.com/likeawizard/tofiks/pkg/board"
)

type EntryType uint8

const (
	TT_UPPER EntryType = iota
	TT_LOWER
	TT_EXACT
)

func (et EntryType) String() string {
	switch et {
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
	age           int8
	newWrite      uint64
	overWrite     uint64
	rejectedWrite uint64
	size          uint64
}

// LSB 0..15 move, 16..22 depth 23..24 type 25..31 unused 32..63 score MSB.
type EntryData uint64

const (
	move_mask   = (1 << 16) - 1
	type_mask   = (1 << 2) - 1
	depth_mask  = (1 << 7) - 1
	age_mask    = (1 << 7) - 1
	score_mask  = (1 << 32) - 1
	depth_shift = 16
	type_shift  = 23
	age_shift   = 25
	score_shift = 32
)

func NewEntry(move board.Move, depth int8, eType EntryType, age int8, score int32) EntryData {
	return EntryData(move) |
		EntryData(depth)<<depth_shift |
		EntryData(eType)<<type_shift |
		EntryData(age)<<age_shift |
		EntryData(score)<<score_shift
}

func (ed EntryData) Get() (board.Move, int8, EntryType, int32) {
	return board.Move(ed & move_mask),
		int8((ed >> depth_shift) & depth_mask),
		EntryType((ed >> type_shift) & type_mask),
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

func (ed EntryData) Type() EntryType {
	return EntryType((ed >> type_shift) & type_mask)
}

func (ed EntryData) Age() int8 {
	return int8((ed >> age_shift) & age_mask)
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

func (tt *TTable) Store(hash uint64, entryType EntryType, eval int32, depth int8, move board.Move) {
	idx := hash % tt.size
	entry := tt.entries[idx].data
	if entry == 0 {
		data := NewEntry(move, depth, entryType, tt.age, eval)
		tt.entries[idx] = SearchEntry{
			key:  hash ^ uint64(data),
			data: data,
		}
		tt.newWrite++
		return
	} else if entryType == TT_EXACT || entry.Depth()-entry.Age() < depth-tt.age {
		data := NewEntry(move, depth, entryType, tt.age, eval)
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

package search

import (
	"unsafe"

	"github.com/likeawizard/tofiks/pkg/board"
)

type (
	// EntryType is a enum type for the type of entry in the transposition table- exact, upper or lower bound.
	EntryType uint8

	// EntryData holds bit encoded data for transposition table entry.
	// LSB 0..31 move, 32..38 depth 39..40 type 41..47 age 48..63 score MSB.
	EntryData uint64

	// tableEntry is a struct for storing the key and data in the transposition table.
	tableEntry struct {
		key  uint64
		data EntryData
	}

	// TTable is a transposition table used for storing search results.
	TTable struct {
		entries       []tableEntry
		Stats         TTStats
		age           int8
		newWrite      uint64
		overWrite     uint64
		rejectedWrite uint64
		size          uint64
	}
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

const (
	// Enum values for EntryType.
	TT_UPPER EntryType = iota
	TT_LOWER
	TT_EXACT

	// Mask and shift values for EntryData.
	move_mask   = board.MoveDataMask
	type_mask   = (1 << 2) - 1
	depth_mask  = (1 << 7) - 1
	age_mask    = (1 << 7) - 1
	score_mask  = (1 << 16) - 1
	depth_shift = 32
	type_shift  = 39
	age_shift   = 41
	score_shift = 48
)

func NewEntry(move board.Move, depth int, eType EntryType, age int8, score int16) EntryData {
	// depth is masked to 7 bits to prevent spillage into the type field.
	return EntryData(move) |
		EntryData(depth&depth_mask)<<depth_shift |
		EntryData(eType)<<type_shift |
		EntryData(age)<<age_shift |
		EntryData(score)<<score_shift
}

func (ed EntryData) GetScore(depth, ply int, alpha, beta int16) (int16, bool) {
	if ed.Depth() < depth {
		return 0, false
	}

	ttType, eval := ed.Type(), ed.Score()

	if eval > CheckmateThreshold {
		eval -= int16(ply)
	}

	if eval < -CheckmateThreshold {
		eval += int16(ply)
	}

	switch {
	case ttType == TT_EXACT:
		return eval, true
	case ttType == TT_UPPER && eval <= alpha:
		return eval, true
	case ttType == TT_LOWER && eval >= beta:
		return eval, true
	}

	return eval, false
}

func (ed EntryData) Depth() int {
	return int((ed >> depth_shift) & depth_mask)
}

func (ed EntryData) Move() board.Move {
	return board.Move(ed & move_mask)
}

func (ed EntryData) Score() int16 {
	return int16(ed >> score_shift)
}

func (ed EntryData) Type() EntryType {
	return EntryType((ed >> type_shift) & type_mask)
}

func (ed EntryData) Age() int8 {
	return int8((ed >> age_shift) & age_mask)
}

const bucketsPerEntry = 2

func NewTTable(sizeInMb int) *TTable {
	eSize := int(unsafe.Sizeof(tableEntry{}))
	totalEntries := (1024 * 1024 * sizeInMb) / eSize
	// Round down to power of 2 for bucket count so we can use bitwise AND.
	buckets := uint64(1)
	for buckets*2 <= uint64(totalEntries/bucketsPerEntry) {
		buckets *= 2
	}
	return &TTable{
		entries: make([]tableEntry, buckets*bucketsPerEntry),
		size:    buckets,
	}
}

func (tt *TTable) Probe(hash uint64) (*EntryData, bool) {
	tt.Stats.recordProbe()
	base := (hash & (tt.size - 1)) * bucketsPerEntry

	for i := uint64(0); i < bucketsPerEntry; i++ {
		e := &tt.entries[base+i]
		if e.key^uint64(e.data) == hash {
			tt.Stats.recordHit(e.data.Depth())
			return &e.data, true
		}
	}

	return nil, false
}

func (tt *TTable) Hashfull() uint64 {
	return (tt.newWrite * 1000) / (tt.size * bucketsPerEntry)
}

func (tt *TTable) Store(hash uint64, entryType EntryType, eval int16, depth, ply int, move board.Move) {
	// Normalize mate scores to position-relative distances for correct retrieval at any ply.
	if eval > CheckmateThreshold {
		eval += int16(ply)
	} else if eval < -CheckmateThreshold {
		eval -= int16(ply)
	}

	data := NewEntry(move, depth, entryType, tt.age, eval)
	newEntry := tableEntry{
		key:  hash ^ uint64(data),
		data: data,
	}

	base := (hash & (tt.size - 1)) * bucketsPerEntry

	// Check for empty or same position in existing buckets.
	for i := uint64(0); i < bucketsPerEntry; i++ {
		d := tt.entries[base+i].data
		if d == 0 {
			tt.entries[base+i] = newEntry
			tt.newWrite++
			tt.Stats.recordNewWrite()
			return
		}
		if tt.entries[base+i].key^uint64(d) == hash {
			tt.entries[base+i] = newEntry
			tt.overWrite++
			tt.Stats.recordOverWrite()
			return
		}
	}

	// All buckets occupied by different positions. Evict the weakest.
	newScore := depth + int(tt.age)
	weakestScore := tt.entries[base].data.Depth() + int(tt.entries[base].data.Age())
	weakestIdx := base

	for i := uint64(1); i < bucketsPerEntry; i++ {
		s := tt.entries[base+i].data.Depth() + int(tt.entries[base+i].data.Age())
		if s < weakestScore {
			weakestScore = s
			weakestIdx = base + i
		}
	}

	if entryType == TT_EXACT || weakestScore < newScore {
		tt.entries[weakestIdx] = newEntry
		tt.overWrite++
		tt.Stats.recordOverWrite()
	} else {
		tt.rejectedWrite++
		tt.Stats.recordRejected()
	}
}

func (tt *TTable) Clear() {
	tt.newWrite = 0
	tt.overWrite = 0
	tt.rejectedWrite = 0
	tt.Stats.reset()
	clear(tt.entries)
}

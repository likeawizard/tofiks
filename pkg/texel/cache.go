package texel

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/likeawizard/tofiks/pkg/board"
)

// Binary format per entry:
//
//	uint16  count of non-zero coefficients
//	count × (uint16 index + float32 value)  [6 bytes each, packed]
//	int16   phase
//	float32 result

// BuildCache reads a texel data file, computes traces in parallel,
// and writes a compact binary cache. Returns the number of entries written.
func BuildCache(dataPath, cachePath string, limit int, workers int) (int, error) {
	f, err := os.Open(dataPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	out, err := os.Create(cachePath)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	bw := bufio.NewWriterSize(out, 1<<20)

	type rawEntry struct {
		fen    string
		result float64
	}

	const batchSize = 4096
	type processedEntry struct {
		entry Entry
		valid bool
	}

	rawCh := make(chan []rawEntry, workers)
	resultCh := make(chan []processedEntry, workers)

	// Writer goroutine — serializes entries to disk.
	var writeWg sync.WaitGroup
	var count int
	writeWg.Go(func() {
		for batch := range resultCh {
			for _, pe := range batch {
				if !pe.valid {
					continue
				}
				if err := writeEntry(bw, &pe.entry); err != nil {
					log.Fatalf("Failed to write cache entry: %v", err)
				}
				count++
			}
		}
		bw.Flush()
	})

	// Worker goroutines — compute traces.
	var workerWg sync.WaitGroup
	for range workers {
		workerWg.Go(func() {
			for raws := range rawCh {
				processed := make([]processedEntry, len(raws))
				for i, r := range raws {
					b := board.NewBoard(r.fen)
					if b.InCheck {
						continue
					}
					trace, phase := TraceEvaluate(b)
					processed[i] = processedEntry{
						entry: Entry{Trace: trace, Phase: phase, Result: r.result},
						valid: true,
					}
				}
				resultCh <- processed
			}
		})
	}

	// Reader — scan file and dispatch batches.
	s := bufio.NewScanner(f)
	var batch []rawEntry
	totalRead := 0
	for s.Scan() {
		line := s.Text()
		scoreStr, fen, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		result, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			continue
		}
		batch = append(batch, rawEntry{fen: fen, result: result})
		totalRead++

		if len(batch) >= batchSize {
			rawCh <- batch
			batch = nil
		}
		if limit > 0 && totalRead >= limit {
			break
		}
	}
	if len(batch) > 0 {
		rawCh <- batch
	}
	close(rawCh)
	workerWg.Wait()
	close(resultCh)
	writeWg.Wait()

	return count, nil
}

// writeEntry writes one entry as raw little-endian bytes (no reflection).
func writeEntry(w io.Writer, e *Entry) error {
	var buf [6]byte // reused for each coeff

	count := uint16(len(e.Trace))
	binary.LittleEndian.PutUint16(buf[:2], count)
	if _, err := w.Write(buf[:2]); err != nil {
		return err
	}

	for _, c := range e.Trace {
		binary.LittleEndian.PutUint16(buf[:2], c.Index)
		binary.LittleEndian.PutUint32(buf[2:6], math.Float32bits(c.Value))
		if _, err := w.Write(buf[:6]); err != nil {
			return err
		}
	}

	binary.LittleEndian.PutUint16(buf[:2], uint16(int16(e.Phase)))
	if _, err := w.Write(buf[:2]); err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(buf[:4], math.Float32bits(float32(e.Result)))
	_, err := w.Write(buf[:4])
	return err
}

// cacheReader wraps a bufio.Reader with bulk byte reads for fast decoding.
type cacheReader struct {
	br  *bufio.Reader
	buf []byte // scratch buffer, grown as needed
}

func newCacheReader(r io.Reader) *cacheReader {
	return &cacheReader{
		br:  bufio.NewReaderSize(r, 1<<20),
		buf: make([]byte, 1024),
	}
}

// readEntry decodes one entry using raw byte reads (no reflection).
func (cr *cacheReader) readEntry(e *Entry) error {
	// Read count (2 bytes).
	if _, err := io.ReadFull(cr.br, cr.buf[:2]); err != nil {
		return err
	}
	count := int(binary.LittleEndian.Uint16(cr.buf[:2]))

	// Read all coefficients in one bulk read: count × 6 bytes.
	need := count*6 + 6 // +6 for phase(2) + result(4)
	if len(cr.buf) < need {
		cr.buf = make([]byte, need)
	}
	if _, err := io.ReadFull(cr.br, cr.buf[:need]); err != nil {
		return err
	}

	// Decode coefficients.
	if cap(e.Trace) >= count {
		e.Trace = e.Trace[:count]
	} else {
		e.Trace = make(Trace, count)
	}
	off := 0
	for i := range count {
		e.Trace[i].Index = binary.LittleEndian.Uint16(cr.buf[off : off+2])
		e.Trace[i].Value = math.Float32frombits(binary.LittleEndian.Uint32(cr.buf[off+2 : off+6]))
		off += 6
	}

	// Decode phase and result.
	e.Phase = int(int16(binary.LittleEndian.Uint16(cr.buf[off : off+2])))
	off += 2
	e.Result = float64(math.Float32frombits(binary.LittleEndian.Uint32(cr.buf[off : off+4])))

	return nil
}

// ForEachEntry streams through a cache file, calling fn for each entry.
// Memory usage is O(1) — only one entry is in memory at a time.
func ForEachEntry(cachePath string, fn func(*Entry)) error {
	f, err := os.Open(cachePath)
	if err != nil {
		return err
	}
	defer f.Close()
	cr := newCacheReader(f)

	var e Entry
	for {
		err := cr.readEntry(&e)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fn(&e)
	}
}

// CountEntries returns the number of entries in a cache file.
func CountEntries(cachePath string) (int, error) {
	var n int
	err := ForEachEntry(cachePath, func(_ *Entry) { n++ })
	return n, err
}

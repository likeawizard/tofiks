//go:build debug

package testsuite

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/likeawizard/tofiks/pkg/board"
	"github.com/likeawizard/tofiks/pkg/search"
)

const healthDepth = 16

func TestSearchHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("search health harness is a local dev tool, not a CI test")
	}
	fmt.Printf("\n=== Search Health Report — depth %d ===\n", healthDepth)

	totalStart := time.Now()
	var agg iterStats
	for _, p := range healthPositions {
		agg.add(runHealthOnce(t, p, healthDepth))
	}
	totalElapsed := time.Since(totalStart).Round(time.Millisecond)

	fmt.Printf("\n----- TOTAL (%s across %d positions) -----\n", totalElapsed, len(healthPositions))
	fmt.Println(agg.format())
	fmt.Printf("\n=== Total elapsed: %s ===\n", totalElapsed)
}

func runHealthOnce(t *testing.T, p healthPos, depth int) iterStats {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	e := search.NewEngine()
	e.Board = board.NewBoard(p.fen)

	start := time.Now()
	e.IDSearch(depth, false)
	elapsed := time.Since(start)

	_ = w.Close()
	os.Stdout = oldStdout
	output := <-done

	block := lastIterationBlock(output)
	if len(block) == 0 {
		t.Errorf("%s: no info output from IDSearch", p.name)
		return iterStats{}
	}

	fmt.Printf("\n----- %s (%s) -----\n", p.name, elapsed.Round(time.Millisecond))
	fmt.Printf("FEN: %s\n", p.fen)
	for _, line := range block {
		fmt.Println(line)
	}
	return parseIter(block)
}

// lastIterationBlock returns the `info depth` line plus trailing `info string`
// lines from the deepest completed iteration in the captured engine output.
func lastIterationBlock(output string) []string {
	var cur, last []string
	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.HasPrefix(line, "info depth "):
			if len(cur) > 0 {
				last = cur
			}
			cur = []string{line}
		case strings.HasPrefix(line, "info string ") && len(cur) > 0:
			cur = append(cur, line)
		}
	}
	if len(cur) > 0 {
		last = cur
	}
	return last
}

// iterStats holds the counts parsed from a single iteration's info strings.
// Fields are summed across positions to produce an aggregate report.
type iterStats struct {
	nodes, timeMs                                int
	fh, fhfN, fhfD, pvsRe                        int
	nmpN, nmpD, rfpN, rfpD, fpN, fpD, lmpN, lmpD int
	seApp, seAtt, mc, iir                        int
	pvChanges, aspFailN, aspFailD, aspReNodes    int
	lmrReN, lmrReD                               int
	ttProbes, ttCutE, ttCutL, ttCutU, ttMate     int
}

func (s *iterStats) add(o iterStats) {
	s.nodes += o.nodes
	s.timeMs += o.timeMs
	s.fh += o.fh
	s.fhfN += o.fhfN
	s.fhfD += o.fhfD
	s.pvsRe += o.pvsRe
	s.nmpN += o.nmpN
	s.nmpD += o.nmpD
	s.rfpN += o.rfpN
	s.rfpD += o.rfpD
	s.fpN += o.fpN
	s.fpD += o.fpD
	s.lmpN += o.lmpN
	s.lmpD += o.lmpD
	s.seApp += o.seApp
	s.seAtt += o.seAtt
	s.mc += o.mc
	s.iir += o.iir
	s.pvChanges += o.pvChanges
	s.aspFailN += o.aspFailN
	s.aspFailD += o.aspFailD
	s.aspReNodes += o.aspReNodes
	s.lmrReN += o.lmrReN
	s.lmrReD += o.lmrReD
	s.ttProbes += o.ttProbes
	s.ttCutE += o.ttCutE
	s.ttCutL += o.ttCutL
	s.ttCutU += o.ttCutU
	s.ttMate += o.ttMate
}

func (s iterStats) format() string {
	nps := 0
	if s.timeMs > 0 {
		nps = 1000 * s.nodes / s.timeMs
	}
	ttCuts := s.ttCutE + s.ttCutL + s.ttCutU
	var b strings.Builder
	fmt.Fprintf(&b, "nodes %d nps %d time %dms\n", s.nodes, nps, s.timeMs)
	fmt.Fprintf(&b, "ordering: fh %d (%s of nodes) fhf %s pvs_re %d (%s of fh)\n",
		s.fh, pctOnly(s.fh, s.nodes),
		pctStr(s.fhfN, s.fhfD),
		s.pvsRe, pctOnly(s.pvsRe, s.fh))
	fmt.Fprintf(&b, "prune: nmp %s rfp %s fp %s lmp %s\n",
		pctStr(s.nmpN, s.nmpD), pctStr(s.rfpN, s.rfpD),
		pctStr(s.fpN, s.fpD), pctStr(s.lmpN, s.lmpD))
	fmt.Fprintf(&b, "       se %s mc %d (%s of se_att) iir %d (%s of nodes)\n",
		pctStr(s.seApp, s.seAtt),
		s.mc, pctOnly(s.mc, s.seAtt),
		s.iir, pctOnly(s.iir, s.nodes))
	fmt.Fprintf(&b, "stability: pv_changes %d asp_fail %s asp_re_nodes %d (%s of nodes) lmr_re %s\n",
		s.pvChanges, pctStr(s.aspFailN, s.aspFailD),
		s.aspReNodes, pctOnly(s.aspReNodes, s.nodes),
		pctStr(s.lmrReN, s.lmrReD))
	fmt.Fprintf(&b, "tt: probes %d (%s of nodes) cuts %d (%s of probes) E:%d L:%d U:%d mate:%d",
		s.ttProbes, pctOnly(s.ttProbes, s.nodes),
		ttCuts, pctOnly(ttCuts, s.ttProbes),
		s.ttCutE, s.ttCutL, s.ttCutU, s.ttMate)
	return b.String()
}

func pctStr(num, denom int) string {
	if denom == 0 {
		return "-"
	}
	return fmt.Sprintf("%s(%d/%d)", pctOnly(num, denom), num, denom)
}

// pctOnly returns a percentage string with adaptive precision so small rates
// like 0.04% aren't truncated to 0% by integer division.
func pctOnly(num, denom int) string {
	if denom == 0 {
		return "-"
	}
	p := float64(num) * 100 / float64(denom)
	switch {
	case p == 0 || p >= 10:
		return fmt.Sprintf("%d%%", int(p))
	case p >= 1:
		return fmt.Sprintf("%.1f%%", p)
	default:
		return fmt.Sprintf("%.2f%%", p)
	}
}

var (
	reDepth    = regexp.MustCompile(`info depth \d+ seldepth \d+ score \S+ \S+ nodes (\d+) nps \d+ time (\d+)`)
	reOrdering = regexp.MustCompile(`ordering: fh (\d+) fhf \d+% \((\d+)/(\d+)\) pvs_re (\d+)`)
	rePruneSE  = regexp.MustCompile(`se (\d+)/(\d+) mc (\d+) iir (\d+)`)
	rePruneInd = regexp.MustCompile(`(nmp|rfp|fp|lmp) \d+%\((\d+)/(\d+)\)`)
	reStab     = regexp.MustCompile(`stability: pv_changes (\d+) avg_delta \d+cp asp_fail \d+% \((\d+)/(\d+)\) asp_re_nodes (\d+) .+ lmr_re \d+% \((\d+)/(\d+)\)`)
	reTT       = regexp.MustCompile(`tt: probes (\d+) hit \d+% cut \d+% \(E:(\d+) L:(\d+) U:(\d+) mate:(\d+)\)`)
)

func parseIter(block []string) iterStats {
	var s iterStats
	for _, line := range block {
		switch {
		case strings.HasPrefix(line, "info depth "):
			if m := reDepth.FindStringSubmatch(line); m != nil {
				s.nodes = atoi(m[1])
				s.timeMs = atoi(m[2])
			}
		case strings.Contains(line, "ordering:"):
			if m := reOrdering.FindStringSubmatch(line); m != nil {
				s.fh = atoi(m[1])
				s.fhfN = atoi(m[2])
				s.fhfD = atoi(m[3])
				s.pvsRe = atoi(m[4])
			}
		case strings.Contains(line, "prune:"):
			for _, m := range rePruneInd.FindAllStringSubmatch(line, -1) {
				n, d := atoi(m[2]), atoi(m[3])
				switch m[1] {
				case "nmp":
					s.nmpN, s.nmpD = n, d
				case "rfp":
					s.rfpN, s.rfpD = n, d
				case "fp":
					s.fpN, s.fpD = n, d
				case "lmp":
					s.lmpN, s.lmpD = n, d
				}
			}
			if m := rePruneSE.FindStringSubmatch(line); m != nil {
				s.seApp = atoi(m[1])
				s.seAtt = atoi(m[2])
				s.mc = atoi(m[3])
				s.iir = atoi(m[4])
			}
		case strings.Contains(line, "stability:"):
			if m := reStab.FindStringSubmatch(line); m != nil {
				s.pvChanges = atoi(m[1])
				s.aspFailN = atoi(m[2])
				s.aspFailD = atoi(m[3])
				s.aspReNodes = atoi(m[4])
				s.lmrReN = atoi(m[5])
				s.lmrReD = atoi(m[6])
			}
		case strings.Contains(line, "tt: probes"):
			if m := reTT.FindStringSubmatch(line); m != nil {
				s.ttProbes = atoi(m[1])
				s.ttCutE = atoi(m[2])
				s.ttCutL = atoi(m[3])
				s.ttCutU = atoi(m[4])
				s.ttMate = atoi(m[5])
			}
		}
	}
	return s
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

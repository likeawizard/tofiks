package eval

import (
	"fmt"
	"time"
)

type Stats struct {
	start  time.Time
	nodes  int
	qNodes int
	evals  int
}

func (es *Stats) Start() {
	es.Clear()
	es.start = time.Now()
}

func (es *Stats) Clear() {
	es.nodes = 0
	es.qNodes = 0
	es.evals = 0
}

func (es *Stats) String() string {
	total := es.nodes + es.qNodes
	duration := int(time.Since(es.start).Microseconds())
	if total == 0 || duration == 0 {
		return "-"
	}
	nps := (1000000 * total) / duration
	return fmt.Sprintf("%snps, total: %s (%s %s), QN: %d%%, evals: %d%%", printNodeCount(nps), printNodeCount(total), printNodeCount(es.nodes), printNodeCount(es.qNodes), (100*es.qNodes)/total, 100*es.evals/total)
}

func printNodeCount(nodes int) string {
	postfix := []string{"", "k", "M", "G"}
	idx := 0
	display := float64(nodes)
	for display > 1000 && idx < len(postfix) {
		display /= 1000
		idx++
	}
	return fmt.Sprintf("%.1f%s", display, postfix[idx])
}

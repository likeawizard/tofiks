package eval

import (
	"fmt"
	"time"
)

type EvalStats struct {
	start  time.Time
	nodes  int
	qNodes int
	evals  int
}

func (es *EvalStats) Start() {
	es.Clear()
	es.start = time.Now()
}

func (es *EvalStats) Clear() {
	es.nodes = 0
	es.qNodes = 0
	es.evals = 0
}

func (es *EvalStats) String() string {
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

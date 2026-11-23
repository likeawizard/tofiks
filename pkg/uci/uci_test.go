package uci

import (
	"os"
	"testing"
	"time"

	"github.com/likeawizard/tofiks/pkg/search"
)

// TestParseGoClockPresence checks which go forms carry a clock.
func TestParseGoClockPresence(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"go", false},
		{"go infinite", false},
		{"go depth 5", false},
		{"go movetime 100", false},
		{"go wtime 0 btime 0", true},
		{"go wtime 1000 btime 1000 winc 10 binc 10", true},
		{"go wtime 1000 btime 1000 movestogo 1", true},
	}
	for _, c := range cases {
		g, ok := ParseUCI(c.cmd).(*Go)
		if !ok {
			t.Fatalf("%q did not parse as go", c.cmd)
		}
		if g.hasClock != c.want {
			t.Errorf("%q hasClock = %v, want %v", c.cmd, g.hasClock, c.want)
		}
	}
}

// TestPonderhitThinksOnOwnClock drives the engine the way python-chess (and so
// lichess-bot) does on a correct ponder prediction: go ponder with the opponent
// on 600s against our 20s, then ponderhit. The reply has to arrive within our
// own budget. The previous handler slept for a third of a budget computed from
// the live board mid-search, which half the time meant the opponent's clock:
// an 18s think with 20s left, i.e. a time forfeit.
func TestPonderhitThinksOnOwnClock(t *testing.T) {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer devnull.Close()
	stdout := os.Stdout
	os.Stdout = devnull
	defer func() { os.Stdout = stdout }()

	e := search.NewEngine()
	Dispatch(e, ParseUCI("position startpos moves e2e4 e7e5 g1f3 b8c6 f1b5 a7a6 b5a4 g8f6 e1g1 f8e7"))

	start := time.Now()
	Dispatch(e, ParseUCI("go ponder wtime 20000 btime 600000 winc 0 binc 0"))
	time.Sleep(300 * time.Millisecond)
	Dispatch(e, ParseUCI("ponderhit"))
	e.WG.Wait()

	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("bestmove took %v after ponderhit; own budget allows ~0.6s, the opponent's clock would give ~17s", elapsed)
	}
}

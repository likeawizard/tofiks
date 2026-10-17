package uci

import "testing"

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

package board

import "testing"

// TestGetGamePhaseClamped checks the phase stays within the taper range at
// both ends of the material scale.
func TestGetGamePhaseClamped(t *testing.T) {
	cases := []struct {
		fen  string
		want int
	}{
		{"4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1", 256},
		{"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPQ/RNBQKBNR w KQkq - 0 1", 0},
		{StartPos, 0},
	}
	for _, c := range cases {
		if got := NewBoard(c.fen).GetGamePhase(); got != c.want {
			t.Errorf("%s: phase = %d, want %d", c.fen, got, c.want)
		}
	}
}

package texel

// Entry holds a pre-computed sparse trace and the game result for one position.
type Entry struct {
	Trace  Trace
	Phase  int
	Result float64 // 1.0 = white win, 0.5 = draw, 0.0 = black win
}

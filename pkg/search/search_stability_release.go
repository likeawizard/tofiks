//go:build !debug

package search

import "github.com/likeawizard/tofiks/pkg/board"

// Stability is a no-op in release builds.
type Stability struct{}

func (s *Stability) recordIteration(_ board.Move, _ int16) {}
func (s *Stability) recordAspiration(_ bool)               {}
func (s *Stability) recordLMR(_ bool)                      {}
func (s *Stability) reset()                                {}
func (s *Stability) String() string                        { return "" }

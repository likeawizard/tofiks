//go:build !debug

package eval

import "github.com/likeawizard/tofiks/pkg/board"

// SearchStability is a no-op in release builds.
type SearchStability struct{}

func (s *SearchStability) recordIteration(_ board.Move, _ int16) {}
func (s *SearchStability) recordAspiration(_ bool)               {}
func (s *SearchStability) recordLMR(_ bool)                      {}
func (s *SearchStability) reset()                                {}
func (s *SearchStability) String() string                        { return "" }

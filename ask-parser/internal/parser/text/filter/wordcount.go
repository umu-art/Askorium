package filter

import "ask-parser/internal/parser/text"

var _ text.BlockFilter = (*MinWordCountFilter)(nil)

// MinWordCountFilter is a no-op stub for V1 (always passes).
// V2: block.WordCount >= threshold.
type MinWordCountFilter struct{}

func (f *MinWordCountFilter) Keep(_ text.RawBlock) bool { return true }

package filter

import "ask-parser/internal/parser/text"

var _ text.BlockFilter = (*MinWordCountFilter)(nil)

const DefaultMinWordCount = 3

// MinWordCountFilter drops blocks with fewer than threshold words.
type MinWordCountFilter struct {
	min int
}

func NewMinWordCountFilter(min int) *MinWordCountFilter {
	return &MinWordCountFilter{min: min}
}

func (f *MinWordCountFilter) Keep(block text.RawBlock) bool {
	return block.WordCount >= f.min
}

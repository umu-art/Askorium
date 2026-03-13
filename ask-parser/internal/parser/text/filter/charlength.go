package filter

import "ask-parser/internal/parser/text"

var _ text.BlockFilter = (*MinCharLengthFilter)(nil)

const DefaultMinCharLength = 10

type MinCharLengthFilter struct {
	min int
}

func NewMinCharLengthFilter(min int) text.BlockFilter {
	return &MinCharLengthFilter{min: min}
}

func (f *MinCharLengthFilter) Keep(b text.RawBlock) bool {
	return len([]rune(b.Text)) >= f.min
}

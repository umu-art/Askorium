package filter

import "ask-parser/internal/parser/text"

var _ text.BlockFilter = (*MaxLinkDensityFilter)(nil)

// MaxLinkDensityFilter is a no-op stub for V1 (always passes).
// V2: block.LinkDensity <= threshold.
type MaxLinkDensityFilter struct{}

func (f *MaxLinkDensityFilter) Keep(_ text.RawBlock) bool { return true }

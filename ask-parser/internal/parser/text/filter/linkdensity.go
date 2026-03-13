package filter

import "ask-parser/internal/parser/text"

var _ text.BlockFilter = (*MaxLinkDensityFilter)(nil)

const DefaultMaxLinkDensity = 0.5

// MaxLinkDensityFilter drops blocks where link density exceeds the threshold.
// Removes navigation lists, "See also" sections, and similar link-heavy noise.
type MaxLinkDensityFilter struct {
	threshold float64
}

func NewMaxLinkDensityFilter(threshold float64) *MaxLinkDensityFilter {
	return &MaxLinkDensityFilter{threshold: threshold}
}

func (f *MaxLinkDensityFilter) Keep(block text.RawBlock) bool {
	return block.LinkDensity <= f.threshold
}

package filter

import "ask-parser/internal/parser/text"

var _ text.BlockFilter = (*MaxLinkDensityFilter)(nil)

const DefaultMaxLinkDensity = 0.5

// contentLinkDensityThreshold is the relaxed threshold for blocks inside
// content containers (article, main). These blocks are more likely to be
// real content with legitimate links (e.g. program catalogs, reference lists).
const contentLinkDensityThreshold = 0.75

// MaxLinkDensityFilter drops blocks where link density exceeds the threshold.
// Removes navigation lists, "See also" sections, and similar link-heavy noise.
//
// Table blocks are always kept — tables serialize mixed text and links as
// content, so link density is not a useful noise signal for them.
// Blocks inside article/main get a relaxed threshold (0.75) because
// link-heavy content (catalogs, program lists) is legitimate there.
type MaxLinkDensityFilter struct {
	threshold float64
}

func NewMaxLinkDensityFilter(threshold float64) *MaxLinkDensityFilter {
	return &MaxLinkDensityFilter{threshold: threshold}
}

func (f *MaxLinkDensityFilter) Keep(block text.RawBlock) bool {
	// Table blocks are exempt: tables legitimately contain links as content.
	if block.TagName == "table" {
		return true
	}

	threshold := f.threshold
	if isInsideContentArea(block.AncestorTags) {
		threshold = contentLinkDensityThreshold
	}
	return block.LinkDensity <= threshold
}

// isInsideContentArea returns true if the block's DOM ancestors include
// a content landmark element (article or main).
func isInsideContentArea(ancestors []string) bool {
	for _, tag := range ancestors {
		if tag == "article" || tag == "main" {
			return true
		}
	}
	return false
}

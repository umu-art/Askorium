package filter

import (
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

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
	if block.Type == scrappermodel.CONTENTBLOCKTYPE_HEADING {
		return true // headings are structural markers, never drop by word count
	}
	return block.WordCount >= f.min
}

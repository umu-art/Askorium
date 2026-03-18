package filter

import (
	"slices"
	"strings"

	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

var _ text.BlockFilter = (*BoilerplateContextFilter)(nil)

var defaultTerminalHeadings = []string{
	"references", "see also", "external links", "further reading",
	"bibliography", "notes", "footnotes", "citations",
	"ссылки", "внешние ссылки", "примечания", "литература",
	"см. также", "источники", "сноски", "библиография", "навигация",
}

// BoilerplateContextFilter drops blocks inside terminal sections
// ("References", "See also", etc.) using backward PrevBlock walk.
// Handles nested headings: h3 under a terminal h2 is still terminal.
type BoilerplateContextFilter struct {
	terminals []string
}

func NewBoilerplateContextFilter() *BoilerplateContextFilter {
	return &BoilerplateContextFilter{terminals: defaultTerminalHeadings}
}

func (f *BoilerplateContextFilter) Keep(block text.RawBlock) bool {
	if block.Type == scrappermodel.CONTENTBLOCKTYPE_HEADING && f.isTerminal(block.Text) {
		return false
	}

	// Walk backward, only considering each next-higher heading level.
	// If the first heading at a given level is terminal — the block is inside it.
	minLevel := int32(7)
	for prev := block.PrevBlock; prev != nil; prev = prev.PrevBlock {
		if prev.Type != scrappermodel.CONTENTBLOCKTYPE_HEADING || prev.HeadingLevel >= minLevel {
			continue
		}
		minLevel = prev.HeadingLevel
		if f.isTerminal(prev.Text) {
			return false
		}
	}
	return true
}

func (f *BoilerplateContextFilter) isTerminal(heading string) bool {
	lower := strings.ToLower(strings.TrimSpace(heading))
	// Strip bracket suffixes like Wikipedia's "[edit]" / "[редактировать]"
	if i := strings.IndexByte(lower, '['); i > 0 {
		lower = strings.TrimSpace(lower[:i])
	}
	return slices.Contains(f.terminals, lower)
}

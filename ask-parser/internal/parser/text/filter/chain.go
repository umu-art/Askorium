package filter

import (
	"fmt"
	"log/slog"
	"strings"

	"ask-parser/internal/parser/text"
)

var _ text.BlockCleaner = (*ChainBlockCleaner)(nil)

type ChainBlockCleaner struct {
	filters []text.BlockFilter
}

func NewChainBlockCleaner(filters ...text.BlockFilter) text.BlockCleaner {
	return &ChainBlockCleaner{filters: filters}
}

func (c *ChainBlockCleaner) Clean(blocks []text.RawBlock) []text.RawBlock {
	debug := slog.Default().Enabled(nil, slog.LevelDebug)
	out := make([]text.RawBlock, 0, len(blocks))
	for _, b := range blocks {
		keep := true
		for _, f := range c.filters {
			if !f.Keep(b) {
				if debug {
					preview := strings.ReplaceAll(b.Text, "\n", " ")
					if len([]rune(preview)) > 60 {
						preview = string([]rune(preview)[:60]) + "..."
					}
					slog.Debug(fmt.Sprintf("filter DROP [%d] <%s> by %T: %s",
						b.Idx, b.TagName, f, preview))
				}
				keep = false
				break
			}
		}
		if keep {
			out = append(out, b)
		}
	}
	return out
}

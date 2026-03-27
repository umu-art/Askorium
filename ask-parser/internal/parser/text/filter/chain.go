package filter

import (
	"fmt"
	"log/slog"
	"strings"

	"ask-parser/internal/parser/text"
)

var _ text.BlockCleaner = (*ChainBlockCleaner)(nil)
var _ text.BlockCleaner = (*SequentialCleaner)(nil)

type ChainBlockCleaner struct {
	filters []text.BlockFilter
}

func NewChainBlockCleaner(filters ...text.BlockFilter) text.BlockCleaner {
	return &ChainBlockCleaner{filters: filters}
}

// SequentialCleaner applies multiple BlockCleaners in order.
type SequentialCleaner struct {
	cleaners []text.BlockCleaner
}

func NewSequentialCleaner(cleaners ...text.BlockCleaner) text.BlockCleaner {
	return &SequentialCleaner{cleaners: cleaners}
}

func (c *SequentialCleaner) Clean(blocks []text.RawBlock) []text.RawBlock {
	for _, cl := range c.cleaners {
		blocks = cl.Clean(blocks)
	}
	return blocks
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

package filter

import "ask-parser/internal/parser/text"

var _ text.BlockCleaner = (*ChainBlockCleaner)(nil)

type ChainBlockCleaner struct {
	filters []text.BlockFilter
}

func NewChainBlockCleaner(filters ...text.BlockFilter) text.BlockCleaner {
	return &ChainBlockCleaner{filters: filters}
}

func (c *ChainBlockCleaner) Clean(blocks []text.RawBlock) []text.RawBlock {
	out := make([]text.RawBlock, 0, len(blocks))
	for _, b := range blocks {
		keep := true
		for _, f := range c.filters {
			if !f.Keep(b) {
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

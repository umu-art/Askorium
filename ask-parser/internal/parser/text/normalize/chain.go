package normalize

import "ask-parser/internal/parser/text"

var _ text.TextNormalizer = (*ChainTextNormalizer)(nil)

type ChainTextNormalizer struct {
	transforms []text.TextTransform
}

func NewChainTextNormalizer(transforms ...text.TextTransform) text.TextNormalizer {
	return &ChainTextNormalizer{transforms: transforms}
}

func (n *ChainTextNormalizer) Normalize(blocks []text.RawBlock) []text.RawBlock {
	for i := range blocks {
		s := blocks[i].Text
		for _, t := range n.transforms {
			s = t.Transform(s)
		}
		blocks[i].Text = s
	}
	return blocks
}

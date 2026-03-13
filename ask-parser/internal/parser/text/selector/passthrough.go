package selector

import "ask-parser/internal/parser/text"

var _ text.BlockSelector = (*PassThroughSelector)(nil)

type PassThroughSelector struct{}

func (s *PassThroughSelector) Select(blocks []text.RawBlock) []text.RawBlock {
	return blocks
}

package selector

import "ask-parser/internal/parser/text"

var _ text.BlockSelector = (*ScoredBlockSelector)(nil)

// ScoredBlockSelector is a stub for V2.
// Will apply ScoringStrategy chains and select blocks above a threshold.
type ScoredBlockSelector struct{}

func (s *ScoredBlockSelector) Select(blocks []text.RawBlock) []text.RawBlock {
	return blocks
}

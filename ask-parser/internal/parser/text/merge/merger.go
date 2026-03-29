package merge

import "ask-parser/internal/parser/text"

var _ text.BlockMerger = (*ChainBlockMerger)(nil)

// MergeConfig controls block merging thresholds.
type MergeConfig struct {
	MinBlockWords    int // blocks shorter than this are candidates for greedy merge (default 30)
	MaxBlockWords    int // merged blocks will not exceed this word count (default 300)
	MaxQuestionWords int // max words for a block to be treated as a question (default 40)
}

func DefaultMergeConfig() MergeConfig {
	return MergeConfig{
		MinBlockWords:    30,
		MaxBlockWords:    300,
		MaxQuestionWords: 40,
	}
}

// MergeRule transforms a block slice by merging adjacent blocks according to
// a specific strategy. Rules are applied in sequence.
type MergeRule interface {
	Apply(blocks []text.RawBlock, cfg MergeConfig) []text.RawBlock
}

// ChainBlockMerger applies a sequence of MergeRules.
type ChainBlockMerger struct {
	rules []MergeRule
	cfg   MergeConfig
}

func NewChainBlockMerger(cfg MergeConfig, rules ...MergeRule) *ChainBlockMerger {
	return &ChainBlockMerger{rules: rules, cfg: cfg}
}

func (m *ChainBlockMerger) Merge(blocks []text.RawBlock) []text.RawBlock {
	for _, r := range m.rules {
		blocks = r.Apply(blocks, m.cfg)
	}
	return blocks
}

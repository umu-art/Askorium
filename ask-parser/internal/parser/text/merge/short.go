package merge

import (
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

// ShortBlockMergeRule greedily merges consecutive short blocks (< MinBlockWords)
// until reaching MinBlockWords or MaxBlockWords. Does not merge across heading boundaries.
type ShortBlockMergeRule struct{}

func (ShortBlockMergeRule) Apply(blocks []text.RawBlock, cfg MergeConfig) []text.RawBlock {
	var out []text.RawBlock
	n := len(blocks)

	for i := 0; i < n; i++ {
		b := blocks[i]

		if b.WordCount >= cfg.MinBlockWords {
			out = append(out, b)
			continue
		}

		// Greedy merge forward
		for i+1 < n &&
			b.WordCount < cfg.MinBlockWords &&
			b.WordCount+blocks[i+1].WordCount <= cfg.MaxBlockWords &&
			blocks[i+1].Type != scrappermodel.CONTENTBLOCKTYPE_HEADING {
			i++
			b.Text += "\n\n" + blocks[i].Text
			b.WordCount = wordCount(b.Text)
			// If the merged block changes type, promote to paragraph
			if blocks[i].Type != b.Type {
				b.Type = scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH
			}
		}

		out = append(out, b)
	}
	return out
}

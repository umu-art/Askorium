package merge

import (
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

// ListCoalesceRule merges consecutive LIST_ITEM blocks into a single block,
// up to MaxBlockWords.
type ListCoalesceRule struct{}

func (ListCoalesceRule) Apply(blocks []text.RawBlock, cfg MergeConfig) []text.RawBlock {
	var out []text.RawBlock
	n := len(blocks)

	for i := 0; i < n; i++ {
		b := blocks[i]

		if b.Type != scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM {
			out = append(out, b)
			continue
		}

		// Accumulate consecutive list items
		for i+1 < n &&
			blocks[i+1].Type == scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM &&
			b.WordCount+blocks[i+1].WordCount <= cfg.MaxBlockWords {
			i++
			b.Text += "\n" + blocks[i].Text
			b.WordCount = wordCount(b.Text)
		}

		out = append(out, b)
	}
	return out
}

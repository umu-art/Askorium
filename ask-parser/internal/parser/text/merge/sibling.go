package merge

import (
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

// maxSiblingGroupFrac: skip groups owning more than 40% of all blocks
// (they represent page-level containers like <body>, <article>, <main>).
const maxSiblingGroupFrac = 0.4

// SiblingMergeRule merges consecutive blocks that share the same GroupIdx
// (nearest common DOM ancestor with ≥2 extracted blocks).
//
// Guards:
//   - Groups owning > 40% of all blocks are skipped (page-level containers).
//   - Heading blocks act as boundaries — never merge across them.
//   - Merged blocks respect MaxBlockWords.
//
// Must run BEFORE HeadingAbsorptionRule so headings are still visible as boundaries.
type SiblingMergeRule struct{}

func (SiblingMergeRule) Apply(blocks []text.RawBlock, cfg MergeConfig) []text.RawBlock {
	if len(blocks) == 0 {
		return blocks
	}

	// Count blocks per GroupIdx
	groupCount := make(map[int]int)
	for _, b := range blocks {
		if b.GroupIdx >= 0 {
			groupCount[b.GroupIdx]++
		}
	}

	// Skip groups that own the majority of blocks (page-level containers)
	total := len(blocks)
	skipGroup := make(map[int]bool)
	for gid, count := range groupCount {
		if float64(count)/float64(total) > maxSiblingGroupFrac {
			skipGroup[gid] = true
		}
	}

	var out []text.RawBlock
	n := len(blocks)

	for i := 0; i < n; i++ {
		b := blocks[i]

		if b.GroupIdx < 0 || skipGroup[b.GroupIdx] || b.Type == scrappermodel.CONTENTBLOCKTYPE_HEADING {
			out = append(out, b)
			continue
		}

		// Merge consecutive same-group non-heading blocks
		gid := b.GroupIdx
		for i+1 < n &&
			blocks[i+1].GroupIdx == gid &&
			blocks[i+1].Type != scrappermodel.CONTENTBLOCKTYPE_HEADING &&
			b.WordCount+blocks[i+1].WordCount <= cfg.MaxBlockWords {
			i++
			b.Text += "\n" + blocks[i].Text
			b.WordCount = wordCount(b.Text)
		}

		out = append(out, b)
	}

	return out
}

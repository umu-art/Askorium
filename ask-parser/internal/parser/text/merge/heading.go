package merge

import (
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"

	"ask-parser/internal/parser/text"
)

// HeadingAbsorptionRule merges heading blocks into the first following content block.
// Consecutive headings (h2 then h3) are chained together as a multi-line prefix.
// Trailing headings with no following content are dropped.
type HeadingAbsorptionRule struct{}

func (HeadingAbsorptionRule) Apply(blocks []text.RawBlock, _ MergeConfig) []text.RawBlock {
	var out []text.RawBlock
	var pendingHeadings []text.RawBlock

	for _, b := range blocks {
		if b.Type == scrappermodel.CONTENTBLOCKTYPE_HEADING {
			pendingHeadings = append(pendingHeadings, b)
			continue
		}

		// Attach accumulated headings to this content block
		if len(pendingHeadings) > 0 {
			prefix := ""
			for _, h := range pendingHeadings {
				if prefix != "" {
					prefix += "\n"
				}
				prefix += h.Text
			}
			b.Text = prefix + "\n" + b.Text
			b.WordCount = wordCount(b.Text)
			// Preserve the highest heading level for metadata
			b.HeadingLevel = pendingHeadings[0].HeadingLevel
			pendingHeadings = pendingHeadings[:0]
		}

		out = append(out, b)
	}

	// Trailing headings with no content are dropped (no append)
	return out
}

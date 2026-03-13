package text

import (
	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

var _ TextPipeline = (*DefaultTextPipeline)(nil)

type DefaultTextPipeline struct {
	pruner     DOMPruner
	segmenter  BlockSegmenter
	cleaner    BlockCleaner
	selector   BlockSelector
	normalizer TextNormalizer
}

func NewDefaultTextPipeline(
	pruner DOMPruner,
	segmenter BlockSegmenter,
	cleaner BlockCleaner,
	selector BlockSelector,
	normalizer TextNormalizer,
) TextPipeline {
	return &DefaultTextPipeline{
		pruner:     pruner,
		segmenter:  segmenter,
		cleaner:    cleaner,
		selector:   selector,
		normalizer: normalizer,
	}
}

func (p *DefaultTextPipeline) Run(doc *goquery.Selection, pageURL string) []scrappermodel.ContentBlock {
	p.pruner.Prune(doc)
	blocks := p.segmenter.Segment(doc, pageURL)
	blocks = p.cleaner.Clean(blocks)
	blocks = p.selector.Select(blocks)
	blocks = p.normalizer.Normalize(blocks)
	return toContentBlocks(blocks)
}

func toContentBlocks(raw []RawBlock) []scrappermodel.ContentBlock {
	out := make([]scrappermodel.ContentBlock, 0, len(raw))
	for _, b := range raw {
		block := *scrappermodel.NewContentBlock(b.Type, b.Text)
		block.SetHtmlId(b.HTMLId)
		if b.HeadingLevel > 0 {
			block.SetHeadingLevel(b.HeadingLevel)
		}
		out = append(out, block)
	}
	return out
}

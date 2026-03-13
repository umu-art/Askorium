package text

import (
	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

// DOMPruner removes noise nodes from a cloned document in-place.
type DOMPruner interface {
	Prune(doc *goquery.Selection)
}

// BlockSegmenter traverses the pruned DOM and produces a flat list of RawBlocks.
type BlockSegmenter interface {
	Segment(doc *goquery.Selection, pageURL string) []RawBlock
}

// BlockFilter is a single predicate applied to each block independently.
type BlockFilter interface {
	Keep(block RawBlock) bool
}

// BlockCleaner applies a chain of BlockFilters to a block slice.
type BlockCleaner interface {
	Clean(blocks []RawBlock) []RawBlock
}

// ScoringStrategy assigns a Score to each block.
// It returns a new slice and must not mutate the input.
type ScoringStrategy interface {
	Score(blocks []RawBlock) []RawBlock
}

// BlockSelector chooses a subset of blocks from the full cleaned set.
// Unlike BlockFilter, it sees the entire set at once — useful for
// "top-30% by score" or "cluster around the density peak".
type BlockSelector interface {
	Select(blocks []RawBlock) []RawBlock
}

// TextTransform is a single string-level text cleaning step.
type TextTransform interface {
	Transform(text string) string
}

// TextNormalizer applies a chain of TextTransforms to every block's text.
type TextNormalizer interface {
	Normalize(blocks []RawBlock) []RawBlock
}

// TextPipeline is the complete internal pipeline used by TextExtractor.
type TextPipeline interface {
	Run(doc *goquery.Selection, pageURL string) []scrappermodel.ContentBlock
}

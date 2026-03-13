package extract

import (
	"ask-parser/internal/parser/core"
	"ask-parser/internal/parser/text"
)

var _ core.Extractor = (*TextExtractor)(nil)

type TextExtractor struct {
	pipeline text.TextPipeline
}

func NewTextExtractor(pipeline text.TextPipeline) core.Extractor {
	return &TextExtractor{pipeline: pipeline}
}

func (e *TextExtractor) Extract(ctx core.ExtractionContext) core.ExtractionResult {
	blocks := e.pipeline.Run(ctx.FullDoc.Clone(), ctx.PageURL)
	return core.ExtractionResult{Blocks: blocks}
}

package text

import scrappermodel "github.com/omo-ri/askorium/go-parser-api"

// RawBlock is the internal unit of the TextPipeline.
// It carries all signals needed by BlockCleaner and BlockSelector.
// It does not leave the text/ package — TextExtractor converts it to scrappermodel.ContentBlock.
type RawBlock struct {
	HTMLId       string
	TagName      string
	Text         string  // raw text before normalisation
	WordCount    int
	LinkDensity  float64 // fraction of words inside <a>
	DOMDepth     int
	Position     float64 // [0, 1] relative position in document
	PrevBlock    *RawBlock
	NextBlock    *RawBlock
	Type         scrappermodel.ContentBlockType
	HeadingLevel int32
	Score        float64 // set by BlockSelector
}

package scoring

import (
	"strings"

	"ask-parser/internal/parser/text"
)

var _ text.ScoringStrategy = (*ReadabilityScoringStrategy)(nil)

// Tag-level score weights.
var tagWeights = map[string]float64{
	"p":          1.0,
	"blockquote": 1.0,
	"pre":        1.0,
	"li":         0.5,
	"td":         -0.5,
	"h1":         -1.0,
	"h2":         -1.0,
	"h3":         -1.0,
	"h4":         -1.0,
	"h5":         -1.0,
	"h6":         -1.0,
}

const (
	commaWeight     = 1.0   // bonus per comma in text
	charLengthCap   = 3.0   // max bonus from char count
	charLengthScale = 100.0 // charCount / scale, capped at charLengthCap
	propagationRate = 0.5   // fraction of score propagated to neighbors
)

// ReadabilityScoringStrategy scores blocks using heuristics from the
// Readability/Arc90 algorithm: tag type, comma count, text length,
// and score propagation to neighboring blocks.
type ReadabilityScoringStrategy struct{}

func NewReadabilityScoringStrategy() *ReadabilityScoringStrategy {
	return &ReadabilityScoringStrategy{}
}

func (s *ReadabilityScoringStrategy) Score(blocks []text.RawBlock) []text.RawBlock {
	// Phase 1: base score per block
	for i := range blocks {
		blocks[i].Score += baseScore(blocks[i])
	}

	// Phase 2: propagate to neighbors via PrevBlock/NextBlock
	propagate(blocks)

	return blocks
}

func baseScore(b text.RawBlock) float64 {
	score := 0.0

	// Tag weight
	if w, ok := tagWeights[b.TagName]; ok {
		score += w
	}

	// Comma count — more commas suggests prose (ASCII + fullwidth + ideographic)
	commas := strings.Count(b.Text, ",") +
		strings.Count(b.Text, "\uFF0C") +
		strings.Count(b.Text, "\u3001")
	score += float64(commas) * commaWeight

	// Length bonus — longer blocks are more likely body text (char-based, matches Readability.js)
	lengthBonus := float64(len(b.Text)) / charLengthScale
	if lengthBonus > charLengthCap {
		lengthBonus = charLengthCap
	}
	score += lengthBonus

	return score
}

// propagate spreads a fraction of each block's score to its immediate neighbors.
func propagate(blocks []text.RawBlock) {
	if len(blocks) < 2 {
		return
	}

	// Snapshot scores before propagation to avoid feedback loops.
	scores := make([]float64, len(blocks))
	for i := range blocks {
		scores[i] = blocks[i].Score
	}

	for i := range blocks {
		if i > 0 {
			blocks[i].Score += scores[i-1] * propagationRate
		}
		if i < len(blocks)-1 {
			blocks[i].Score += scores[i+1] * propagationRate
		}
	}
}

package scoring

import (
	"strings"
	"unicode"

	"ask-parser/internal/parser/text"
)

var _ text.ScoringStrategy = (*TextDensityStrategy)(nil)

const (
	charsPerLine   = 80
	stopwordsLowTh = 0.10 // below this → not natural language
	noisePenalty_  = -2.0 // active penalty for non-natural-language blocks
	neighborBoost  = 0.5  // boost fraction from high-density neighbor (Boilerpipe)
)

// TextDensityStrategy scores blocks using Boilerpipe TDQ with two extensions:
//   - JusText stopword density gating (blocks below threshold get negative score)
//   - Boilerpipe neighbor context: a block adjacent to a high-density block
//     gets a boost, capturing transition sentences that would otherwise score low
//
// References:
//   - Kohlschütter et al., "Boilerplate Detection using Shallow Text Features", WSDM 2010
//   - Pomikálek, "Removing Boilerplate and Duplicate Content from Web Corpora", PhD thesis, 2011
type TextDensityStrategy struct {
	stopwords map[string]struct{}
}

func NewTextDensityStrategy() *TextDensityStrategy {
	return &TextDensityStrategy{
		stopwords: getStopwords(),
	}
}

func (s *TextDensityStrategy) Score(blocks []text.RawBlock) []text.RawBlock {
	// Phase 1: compute per-block TDQ
	densities := make([]float64, len(blocks))
	for i := range blocks {
		densities[i] = s.tdq(&blocks[i])
	}

	// Phase 2: apply neighbor context (Boilerpipe's key insight)
	// A low-density block adjacent to a high-density block is likely
	// a transition sentence — boost it.
	for i := range blocks {
		bonus := 0.0
		if blocks[i].PrevBlock != nil && i > 0 && densities[i-1] > 0 {
			bonus = densities[i-1] * neighborBoost
		}
		if blocks[i].NextBlock != nil && i < len(blocks)-1 && densities[i+1] > 0 {
			next := densities[i+1] * neighborBoost
			if next > bonus {
				bonus = next
			}
		}
		blocks[i].Score += densities[i] + bonus
	}

	return blocks
}

// tdq computes text density quality for a single block.
// Returns negative score for non-natural-language blocks (active penalty).
func (s *TextDensityStrategy) tdq(b *text.RawBlock) float64 {
	if b.WordCount == 0 {
		return noisePenalty_
	}

	words := tokenize(b.Text)
	if len(words) == 0 {
		return noisePenalty_
	}

	// Stopword density gate
	swCount := 0
	for _, w := range words {
		if _, ok := s.stopwords[w]; ok {
			swCount++
		}
	}
	swDensity := float64(swCount) / float64(len(words))

	if swDensity < stopwordsLowTh {
		return noisePenalty_
	}

	// TDQ for blocks that pass the gate
	numLines := float64(len(b.Text))/charsPerLine + 1
	density := float64(len(words)) / numLines
	return density * (1 - b.LinkDensity)
}

// tokenize splits text into lowercase letter-bearing tokens with
// leading/trailing punctuation stripped. Pure digit groups, lone
// punctuation, and symbols are excluded.
func tokenize(s string) []string {
	raw := strings.Fields(s)
	out := make([]string, 0, len(raw))
	for _, w := range raw {
		w = trimPunct(w)
		if w == "" || !hasLetter(w) {
			continue
		}
		out = append(out, strings.ToLower(w))
	}
	return out
}

func trimPunct(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r)
	})
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

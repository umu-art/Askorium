package scoring

import (
	"strings"

	"ask-parser/internal/parser/text"
)

var _ text.ScoringStrategy = (*SemanticMarkupStrategy)(nil)

const (
	mainArticleBonus = 3.0
	sectionBonus     = 1.0
	positionPenalty  = -1.5 // penalty for blocks in the document's head/tail regions
	idBonus          = 2.0  // bonus for content-indicating HTMLId
	idPenalty        = -2.0 // penalty for noise-indicating HTMLId
)

// boostTags gives a positive signal for blocks inside content containers.
// nav/aside/footer/header are NOT penalized here — SemanticNoisePruner
// already removes them from the DOM before segmentation.
var boostTags = map[string]float64{
	"main":    mainArticleBonus,
	"article": mainArticleBonus,
	"section": sectionBonus,
}

// Readability.js-inspired id attribute patterns.
var (
	positiveIdPatterns = []string{
		"article", "body", "content", "entry", "main",
		"page", "post", "text", "blog", "story",
	}
	negativeIdPatterns = []string{
		"comment", "combx", "contact", "masthead", "media",
		"meta", "outbrain", "promo", "related", "scroll",
		"shoutbox", "sidebar", "sponsor", "shopping", "tags",
		"tool", "widget", "social", "share", "banner",
		"menu", "advert",
	}
)

// SemanticMarkupStrategy scores blocks based on DOM signals that survive pruning:
//   - Ancestor boost tags (article, main, section) — pruner doesn't touch these
//   - HTMLId attribute pattern matching (Readability.js heuristic)
//   - Document position (first 5% and last 15% penalized)
type SemanticMarkupStrategy struct{}

func NewSemanticMarkupStrategy() *SemanticMarkupStrategy {
	return &SemanticMarkupStrategy{}
}

func (s *SemanticMarkupStrategy) Score(blocks []text.RawBlock) []text.RawBlock {
	for i := range blocks {
		blocks[i].Score += scoreAncestors(blocks[i].AncestorTags)
		blocks[i].Score += scoreHTMLId(blocks[i].HTMLId)
		blocks[i].Score += scorePosition(blocks[i].Position)
	}
	return blocks
}

func scoreAncestors(ancestors []string) float64 {
	bestBoost := 0.0
	for _, tag := range ancestors {
		if b, ok := boostTags[tag]; ok && b > bestBoost {
			bestBoost = b
		}
	}
	return bestBoost
}

// scoreHTMLId checks the block's id attribute against Readability.js patterns.
// Negative checked first — a block with id="sidebar-content" is noise, not content.
func scoreHTMLId(htmlId string) float64 {
	if htmlId == "" {
		return 0
	}
	lower := strings.ToLower(htmlId)

	for _, p := range negativeIdPatterns {
		if strings.Contains(lower, p) {
			return idPenalty
		}
	}
	for _, p := range positiveIdPatterns {
		if strings.Contains(lower, p) {
			return idBonus
		}
	}
	return 0
}

// scorePosition penalizes blocks in the first 5% and last 15% of the document.
// These regions disproportionately contain boilerplate (headers, footers, copyright).
// Note: SemanticNoisePruner removes nav/header/footer elements, but many sites
// use generic divs without semantic tags — position catches those.
func scorePosition(pos float64) float64 {
	if pos < 0.05 || pos > 0.85 {
		return positionPenalty
	}
	return 0
}

package prune

import (
	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*VisibilityPruner)(nil)

// VisibilityPruner is a no-op stub for V1.
// V2: remove display:none, visibility:hidden, aria-hidden="true" elements.
type VisibilityPruner struct{}

func (p *VisibilityPruner) Prune(_ *goquery.Selection) {}

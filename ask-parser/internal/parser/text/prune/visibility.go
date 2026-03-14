package prune

import (
	"github.com/PuerkitoBio/goquery"
	"strings"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*VisibilityPruner)(nil)

// VisibilityPruner is a no-op stub for V1.
// V2: remove display:none, visibility:hidden, aria-hidden="true" elements.
type VisibilityPruner struct{}

var visibilitySelectors = []string{
	`[aria-hidden="true"]`,
	`[hidden]`,
	`[style*="display:none"]`,
	`[style*="display: none"]`,
	`[style*="visibility:hidden"]`,
	`[style*="visibility: hidden"]`,
	`[style*="opacity:0"]`,
	`[style*="opacity: 0"]`,
	`[style*="clip:rect(0"]`,
	`[style*="clip: rect(0"]`,
	`[style*="height:0"]`,
	`[style*="height: 0"]`,
	`[style*="width:0"]`,
	`[style*="width: 0"]`,
	`[style*="overflow:hidden"][style*="height:0"]`,
	`[style*="position:absolute"][style*="left:-"]`,
	`[style*="position: absolute"][style*="left: -"]`,
	`[role="presentation"][aria-hidden]`,
}

func (p *VisibilityPruner) Prune(doc *goquery.Selection) {
	// Static selectors that can be matched directly via CSS
	combined := strings.Join(visibilitySelectors, ", ")
	doc.Find(combined).Remove()
}

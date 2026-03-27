package prune

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*VisibilityPruner)(nil)

// VisibilityPruner removes elements hidden via inline styles, attributes,
// or ARIA roles. Elements with display:none/visibility:hidden that contain
// block-level content (p, table, h1-h6, ul, ol) are kept — these are
// typically accordion/foldable panels whose content is revealed by JS.
type VisibilityPruner struct{}

// safeHiddenSelector matches display:none and visibility:hidden — these
// need the content check before removal.
var safeHiddenSelectors = []string{
	`[style*="display:none"]`,
	`[style*="display: none"]`,
	`[style*="visibility:hidden"]`,
	`[style*="visibility: hidden"]`,
}

// aggressiveSelectors match hiding techniques that never wrap real content.
var aggressiveSelectors = []string{
	`[aria-hidden="true"]`,
	`[hidden]`,
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

// substantialContentSelector detects content patterns typical of
// accordion/foldable panels: tables, or multiple block elements.
const substantialContentSelector = "table"

// minBlockElements is the number of block-level content elements
// required to consider a display:none container as an accordion panel.
const minBlockElements = 2

const blockElementSelector = "p, h1, h2, h3, h4, h5, h6, ul, ol, blockquote, dl, li"

func (p *VisibilityPruner) Prune(doc *goquery.Selection) {
	// Aggressive selectors: always remove.
	aggressive := strings.Join(aggressiveSelectors, ", ")
	doc.Find(aggressive).Remove()

	// Safe hidden selectors: only remove if no substantial content inside.
	// A table always qualifies as substantial. Otherwise, require 3+
	// block elements — a single <p> (e.g. cookie banner) does not qualify.
	safe := strings.Join(safeHiddenSelectors, ", ")
	doc.Find(safe).Each(func(_ int, s *goquery.Selection) {
		if s.Find(substantialContentSelector).Length() > 0 {
			return // contains a table — likely data panel
		}
		if s.Find(blockElementSelector).Length() >= minBlockElements {
			return // substantial block content — likely accordion
		}
		s.Remove()
	})
}

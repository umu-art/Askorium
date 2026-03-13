package prune

import (
	"github.com/PuerkitoBio/goquery"

	"ask-parser/internal/parser/text"
)

var _ text.DOMPruner = (*SemanticNoisePruner)(nil)

// noiseSelectors targets structural noise elements.
// "body > header" and "body > div > header" removes only page-level headers;
// headers inside <article> may contain author/date metadata and are kept.
var noiseSelectors = []string{
	// HTML5 semantic elements
	"nav",
	"footer",
	"aside",
	"body > header",
	"body > div > header",

	// ARIA role equivalents — same semantics as above but via <div role="...">.
	// Sites often use generic elements with roles instead of semantic HTML5 tags.
	"[role=\"navigation\"]",
	"[role=\"complementary\"]", // aside equivalent
	"[role=\"contentinfo\"]",   // footer equivalent
	"[role=\"banner\"]",        // page-level header equivalent
	"[role=\"search\"]",        // search widgets
	"[role=\"dialog\"]",        // popups, cookie banners
}

type SemanticNoisePruner struct{}

func (p *SemanticNoisePruner) Prune(doc *goquery.Selection) {
	for _, sel := range noiseSelectors {
		doc.Find(sel).Remove()
	}
}

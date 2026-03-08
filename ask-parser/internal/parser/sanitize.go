package parser

import "github.com/PuerkitoBio/goquery"

var noiseSelectors = []string{
	"nav",
	"aside",
	"footer",
	"header",
	"noscript",
	"script",
	"style",
	"svg",
	"iframe",
	"[hidden]",
	`[aria-hidden="true"]`,
	`[style*="display:none"]`,
	`[style*="display: none"]`,
}

type domSanitizer struct{}

func (s *domSanitizer) Sanitize(root *goquery.Selection) {
	for _, sel := range noiseSelectors {
		root.Find(sel).Remove()
	}
}

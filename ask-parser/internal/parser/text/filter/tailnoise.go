package filter

import (
	"strings"

	"ask-parser/internal/parser/text"
)

var _ text.BlockFilter = (*TailNoiseFilter)(nil)

// tailNoisePatterns are substrings that identify boilerplate footer blocks.
// A block whose text contains any of these patterns is dropped.
var tailNoisePatterns = []string{
	// Copyright / legal
	"условия использования материалов",
	"политика конфиденциальности",
	"© ниу вшэ",

	// Cookie / GDPR banners
	"файлы cookies",
	"файлов cookies",
	"cookie",

	// Footer contact blocks
	"мы в соцсетях",
	"карта сайта",

	// Government portal links often leaking from footer
	"minobrnauki.gov.ru",
	"edu.gov.ru",

	// Common footer boilerplate
	"все права защищены",
	"terms of use",
	"privacy policy",
	"all rights reserved",

	// HSE-specific footer
	"шрифты hse sans",
	"разработаны в школе дизайна",

	// Related-articles teasers
	"вам также может быть интересно",
}

// TailNoiseFilter drops blocks that match known boilerplate footer phrases.
// Works as a supplement to DOM-based pruning for sites where the footer
// is not wrapped in a <footer> tag or marked with ARIA roles.
type TailNoiseFilter struct{}

func (f *TailNoiseFilter) Keep(block text.RawBlock) bool {
	// Normalize non-breaking spaces before matching, because at this pipeline
	// stage WhitespaceCollapseTransform has not run yet and HTML &nbsp; / \xa0
	// would prevent patterns like "© ниу вшэ" from matching.
	lower := strings.ToLower(strings.ReplaceAll(block.Text, "\u00a0", " "))
	for _, pattern := range tailNoisePatterns {
		if strings.Contains(lower, pattern) {
			return false
		}
	}
	return true
}

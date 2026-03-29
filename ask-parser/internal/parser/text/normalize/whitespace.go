package normalize

import (
	"regexp"
	"strings"
	"unicode"

	"ask-parser/internal/parser/text"
)

var _ text.TextTransform = (*WhitespaceCollapseTransform)(nil)

// unicodeSpaceRe matches runs of ASCII whitespace plus Unicode space separators
// (category Zs: U+00A0, U+2000–U+200A, U+202F, U+205F, U+3000, etc.)
// that Go's \s does not cover.
var unicodeSpaceRe = regexp.MustCompile(`\s+`)

type WhitespaceCollapseTransform struct{}

func (t *WhitespaceCollapseTransform) Transform(s string) string {
	// First, replace all Unicode space separators (Zs) with regular space.
	// Go regexp \s only matches ASCII whitespace, so Zs chars like U+00A0
	// would otherwise survive the collapse step.
	s = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Zs, r) {
			return ' '
		}
		return r
	}, s)
	return unicodeSpaceRe.ReplaceAllString(s, " ")
}

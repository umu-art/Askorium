package normalize

import (
	"strings"
	"unicode"

	"ask-parser/internal/parser/text"
)

var _ text.TextTransform = (*ControlCharStripTransform)(nil)

// ControlCharStripTransform strips Unicode control characters (Cc, Cf)
// except common whitespace (\t, \n, \r) which is handled by WhitespaceCollapseTransform.
type ControlCharStripTransform struct{}

func (t *ControlCharStripTransform) Transform(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return r
		}
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}

package normalize

import (
	"ask-parser/internal/parser/text"
	"strings"
)

var _ text.TextTransform = (*ValidUTF8Transform)(nil)

// ValidUTF8Transform drops invalid UTF-8 byte sequences from the text.
// Uses strings.ToValidUTF8 from the standard library — no extra dependencies.
type ValidUTF8Transform struct{}

func (t *ValidUTF8Transform) Transform(s string) string {
	return strings.ToValidUTF8(s, "")
}

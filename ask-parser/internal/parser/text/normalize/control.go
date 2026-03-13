package normalize

import "ask-parser/internal/parser/text"

var _ text.TextTransform = (*ControlCharStripTransform)(nil)

// ControlCharStripTransform is a no-op stub for V1.
// V2: strip Unicode control characters (categories Cc, Cf).
type ControlCharStripTransform struct{}

func (t *ControlCharStripTransform) Transform(s string) string { return s }

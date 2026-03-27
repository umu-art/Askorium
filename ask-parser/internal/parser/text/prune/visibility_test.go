package prune

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	return doc
}

func TestVisibilityPruner_Prune(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantText string // expected remaining visible text (trimmed)
		goneText string // text that must NOT appear after pruning
	}{
		{
			name:     "aria-hidden=true removed",
			html:     `<div><p>visible</p><span aria-hidden="true">hidden</span></div>`,
			wantText: "visible",
			goneText: "hidden",
		},
		{
			name:     "hidden attribute removed",
			html:     `<div><p>visible</p><p hidden>secret</p></div>`,
			wantText: "visible",
			goneText: "secret",
		},
		{
			name:     "display:none removed",
			html:     `<div><p>visible</p><p style="display:none">invisible</p></div>`,
			wantText: "visible",
			goneText: "invisible",
		},
		{
			name:     "display: none with space removed",
			html:     `<div><p>visible</p><p style="display: none">invisible</p></div>`,
			wantText: "visible",
			goneText: "invisible",
		},
		{
			name:     "visibility:hidden removed",
			html:     `<div><p>visible</p><p style="visibility:hidden">ghost</p></div>`,
			wantText: "visible",
			goneText: "ghost",
		},
		{
			name:     "opacity:0 removed",
			html:     `<div><p>visible</p><span style="opacity:0">transparent</span></div>`,
			wantText: "visible",
			goneText: "transparent",
		},
		{
			name:     "position:absolute left:-9999px removed",
			html:     `<div><p>visible</p><span style="position:absolute;left:-9999px">offscreen</span></div>`,
			wantText: "visible",
			goneText: "offscreen",
		},
		{
			name:     "nested hidden children also removed",
			html:     `<div><div aria-hidden="true"><p>deep</p><span>nested</span></div><p>keep</p></div>`,
			wantText: "keep",
			goneText: "deep",
		},
		{
			name:     "no false positive on normal elements",
			html:     `<div><p>hello</p><p style="color:red">world</p></div>`,
			wantText: "hello",
			goneText: "",
		},
		{
			name:     "mixed hidden and visible",
			html:     `<div><p>aaa</p><p style="display:none">bbb</p><p hidden>ccc</p><p aria-hidden="true">ddd</p><p>eee</p></div>`,
			wantText: "aaa",
			goneText: "bbb",
		},
	}

	pruner := &VisibilityPruner{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			pruner.Prune(doc.Selection)

			result := doc.Text()

			if !strings.Contains(result, tt.wantText) {
				t.Errorf("expected visible text %q to remain, got: %q", tt.wantText, result)
			}

			if tt.goneText != "" && strings.Contains(result, tt.goneText) {
				t.Errorf("expected hidden text %q to be removed, got: %q", tt.goneText, result)
			}
		})
	}
}

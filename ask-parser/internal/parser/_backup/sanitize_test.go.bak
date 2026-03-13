package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func docFromHTML(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	return doc
}

func TestSanitize_RemovesNav(t *testing.T) {
	doc := docFromHTML(t, `<body><nav>menu</nav><p>content</p></body>`)
	s := &domSanitizer{}
	s.Sanitize(doc.Find("body"))

	if doc.Find("nav").Length() != 0 {
		t.Error("nav should be removed")
	}
	if doc.Find("p").Text() != "content" {
		t.Error("p should remain")
	}
}

func TestSanitize_RemovesFooterHeaderAside(t *testing.T) {
	doc := docFromHTML(t, `<body>
		<header>header</header>
		<aside>sidebar</aside>
		<main><p>content</p></main>
		<footer>footer</footer>
	</body>`)
	s := &domSanitizer{}
	s.Sanitize(doc.Find("body"))

	for _, tag := range []string{"header", "aside", "footer"} {
		if doc.Find(tag).Length() != 0 {
			t.Errorf("%s should be removed", tag)
		}
	}
	if doc.Find("p").Text() != "content" {
		t.Error("content in main should remain")
	}
}

func TestSanitize_RemovesHiddenElements(t *testing.T) {
	doc := docFromHTML(t, `<body>
		<div hidden>hidden div</div>
		<div aria-hidden="true">aria hidden</div>
		<div style="display:none">display none</div>
		<div style="display: none">display none spaced</div>
		<p>visible</p>
	</body>`)
	s := &domSanitizer{}
	s.Sanitize(doc.Find("body"))

	divs := doc.Find("div")
	if divs.Length() != 0 {
		t.Errorf("expected 0 divs, got %d", divs.Length())
	}
}

func TestSanitize_RemovesScriptStyleNoscript(t *testing.T) {
	doc := docFromHTML(t, `<body>
		<script>alert(1)</script>
		<style>.x{}</style>
		<noscript>enable js</noscript>
		<p>text</p>
	</body>`)
	s := &domSanitizer{}
	s.Sanitize(doc.Find("body"))

	for _, tag := range []string{"script", "style", "noscript"} {
		if doc.Find(tag).Length() != 0 {
			t.Errorf("%s should be removed", tag)
		}
	}
}

func TestSanitize_PreservesMainContent(t *testing.T) {
	doc := docFromHTML(t, `<body>
		<nav>nav</nav>
		<main>
			<h1>Title</h1>
			<p>Paragraph</p>
			<ul><li>Item</li></ul>
		</main>
		<footer>foot</footer>
	</body>`)
	s := &domSanitizer{}
	s.Sanitize(doc.Find("body"))

	if doc.Find("h1").Length() != 1 {
		t.Error("h1 should remain")
	}
	if doc.Find("p").Length() != 1 {
		t.Error("p should remain")
	}
	if doc.Find("li").Length() != 1 {
		t.Error("li should remain")
	}
}

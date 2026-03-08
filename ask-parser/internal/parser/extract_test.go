package parser

import (
	"testing"

	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

func TestSelectContentRoot_Main(t *testing.T) {
	doc := docFromHTML(t, `<body><nav>nav</nav><main><p>content</p></main></body>`)
	root := selectContentRoot(doc)
	if root.Find("p").Text() != "content" {
		t.Error("should select main")
	}
	if root.Find("nav").Length() != 0 {
		t.Error("should not include nav")
	}
}

func TestSelectContentRoot_Article(t *testing.T) {
	doc := docFromHTML(t, `<body><article><p>article</p></article></body>`)
	root := selectContentRoot(doc)
	if root.Find("p").Text() != "article" {
		t.Error("should select article when no main")
	}
}

func TestSelectContentRoot_Body(t *testing.T) {
	doc := docFromHTML(t, `<body><div><p>fallback</p></div></body>`)
	root := selectContentRoot(doc)
	if root.Find("p").Text() != "fallback" {
		t.Error("should fallback to body")
	}
}

func TestClassifyElement(t *testing.T) {
	tests := []struct {
		html          string
		selector      string
		expectedType  scrappermodel.ContentBlockType
		expectedLevel int32
	}{
		{`<h1>T</h1>`, "h1", scrappermodel.CONTENTBLOCKTYPE_HEADING, 1},
		{`<h3>T</h3>`, "h3", scrappermodel.CONTENTBLOCKTYPE_HEADING, 3},
		{`<h6>T</h6>`, "h6", scrappermodel.CONTENTBLOCKTYPE_HEADING, 6},
		{`<p>T</p>`, "p", scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, 0},
		{`<li>T</li>`, "li", scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, 0},
	}

	for _, tt := range tests {
		doc := docFromHTML(t, tt.html)
		sel := doc.Find(tt.selector)
		blockType, level := classifyElement(sel)
		if blockType != tt.expectedType {
			t.Errorf("%s: expected type %v, got %v", tt.selector, tt.expectedType, blockType)
		}
		if level != tt.expectedLevel {
			t.Errorf("%s: expected level %d, got %d", tt.selector, tt.expectedLevel, level)
		}
	}
}

func TestCalcPosition_RuneBased(t *testing.T) {
	text := "Привет мир hello"
	pos := calcPosition(text, "hello")
	// "Привет мир " = 11 runes
	if pos != 11 {
		t.Errorf("expected rune position 11, got %d", pos)
	}
}

func TestMakeSnippet_RuneSafe(t *testing.T) {
	text := "Привет мир, это тестовый текст для сниппета"
	runes := []rune(text)
	pos := int32(12) // "это"
	snippet := makeSnippet(text, pos, 5)

	snippetRunes := []rune(snippet)
	if len(snippetRunes) > 10 {
		t.Errorf("snippet too long: %d runes", len(snippetRunes))
	}
	// Should not contain broken UTF-8
	for _, r := range snippetRunes {
		if r == '\ufffd' {
			t.Error("snippet contains replacement character — broken UTF-8")
		}
	}
	_ = runes // suppress unused
}

func TestExtractMetadata(t *testing.T) {
	doc := docFromHTML(t, `<html lang="ru">
		<head>
			<title>Test Title</title>
			<meta name="description" content="Test desc">
			<meta property="og:image" content="https://example.com/img.jpg">
			<link rel="icon" href="/favicon.ico">
		</head>
		<body></body>
	</html>`)

	meta := extractMetadata(doc)
	if meta.Title != "Test Title" {
		t.Errorf("expected 'Test Title', got '%s'", meta.Title)
	}
	if meta.Description == nil || *meta.Description != "Test desc" {
		t.Errorf("expected 'Test desc', got %v", meta.Description)
	}
	if meta.PreviewURL == nil || *meta.PreviewURL != "https://example.com/img.jpg" {
		t.Error("expected previewURL")
	}
	if meta.IconURL == nil || *meta.IconURL != "/favicon.ico" {
		t.Error("expected iconURL")
	}
	if meta.Language == nil || *meta.Language != "ru" {
		t.Error("expected language=ru")
	}
}

func TestExtractMetadata_Fallbacks(t *testing.T) {
	doc := docFromHTML(t, `<html>
		<head>
			<meta property="og:title" content="OG Title">
			<meta property="og:description" content="OG desc">
		</head>
		<body></body>
	</html>`)

	meta := extractMetadata(doc)
	if meta.Title != "OG Title" {
		t.Errorf("expected 'OG Title', got '%s'", meta.Title)
	}
	if meta.Description == nil || *meta.Description != "OG desc" {
		t.Error("expected og:description fallback")
	}
}

func TestExtractMetadata_Untitled(t *testing.T) {
	doc := docFromHTML(t, `<html><head></head><body></body></html>`)
	meta := extractMetadata(doc)
	if meta.Title != "Untitled" {
		t.Errorf("expected 'Untitled', got '%s'", meta.Title)
	}
}

func TestContentExtractor_FullPage(t *testing.T) {
	html := `<html><head><title>Test</title></head><body>
		<main>
			<h1>Title</h1>
			<p>Hello <a href="/about">about</a> world</p>
			<ul><li>Item <a href="https://other.com">link</a></li></ul>
			<p>Doc <a href="/file.pdf">download</a></p>
			<p><img src="/photo.jpg" alt="photo"></p>
		</main>
	</body></html>`

	ext := &contentExtractor{}
	doc := docFromHTML(t, html)
	result := ext.Extract(doc, "https://example.com")

	// blocks: h1, p(Hello about world), li(Item link), p(Doc download) = 4
	// p with only <img> is skipped (empty text after TrimSpace)
	if len(result.Blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(result.Blocks))
	}
	if result.Blocks[0].GetType() != scrappermodel.CONTENTBLOCKTYPE_HEADING {
		t.Error("first block should be heading")
	}

	// links
	if len(result.Links) != 2 {
		t.Fatalf("expected 2 links (about + other.com), got %d", len(result.Links))
	}

	// check link classification
	foundInternal := false
	foundExternal := false
	for _, l := range result.Links {
		if l.GetType() == scrappermodel.LINKTYPE_INTERNAL {
			foundInternal = true
		}
		if l.GetType() == scrappermodel.LINKTYPE_EXTERNAL {
			foundExternal = true
		}
	}
	if !foundInternal || !foundExternal {
		t.Errorf("expected both internal and external links, internal=%v external=%v", foundInternal, foundExternal)
	}

	// documents: only pdf (img is in an empty-text <p>, which is skipped)
	if len(result.Documents) != 1 {
		t.Fatalf("expected 1 document (pdf), got %d", len(result.Documents))
	}
}

func TestContentExtractor_FiltersJunkHrefs(t *testing.T) {
	html := `<body><p>
		<a href="#">anchor</a>
		<a href="javascript:void(0)">js</a>
		<a href="mailto:x@x.com">email</a>
		<a href="tel:+1234">phone</a>
		<a href="/real">real link</a>
	</p></body>`

	ext := &contentExtractor{}
	doc := docFromHTML(t, html)
	result := ext.Extract(doc, "https://example.com")

	if len(result.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(result.Links))
	}
	if result.Links[0].GetHref() != "https://example.com/real" {
		t.Errorf("expected resolved URL, got %s", result.Links[0].GetHref())
	}
}

func TestContentExtractor_EmptyTextSkipped(t *testing.T) {
	html := `<body><p>   </p><p>real content</p></body>`
	ext := &contentExtractor{}
	doc := docFromHTML(t, html)
	result := ext.Extract(doc, "https://example.com")

	if len(result.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(result.Blocks))
	}
}

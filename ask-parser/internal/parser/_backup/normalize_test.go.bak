package parser

import (
	"testing"

	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

func TestFilterShortBlocks(t *testing.T) {
	blocks := []scrappermodel.ContentBlock{
		*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, "A"),
		*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, "Hello world"),
		*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, ""),
		*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, "OK"),
	}

	result := filterShortBlocks(blocks)
	if len(result) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(result))
	}
	if result[0].GetText() != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", result[0].GetText())
	}
	if result[1].GetText() != "OK" {
		t.Errorf("expected 'OK', got '%s'", result[1].GetText())
	}
}

func TestCollapseWhitespace(t *testing.T) {
	blocks := []scrappermodel.ContentBlock{
		*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, "hello   world"),
		*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, "line1\n\n\tline2"),
		*scrappermodel.NewContentBlock(scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, "  leading trailing  "),
	}

	result := collapseWhitespace(blocks)
	expected := []string{"hello world", "line1 line2", "leading trailing"}
	for i, exp := range expected {
		if result[i].GetText() != exp {
			t.Errorf("block %d: expected '%s', got '%s'", i, exp, result[i].GetText())
		}
	}
}

func TestNormalizeURL_RemovesUTM(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/page?utm_source=twitter&utm_medium=social", "https://example.com/page"},
		{"https://example.com/page?q=test&utm_campaign=spring", "https://example.com/page?q=test"},
		{"https://example.com/page#section", "https://example.com/page"},
		{"https://example.com/page?q=test#section", "https://example.com/page?q=test"},
		{"https://example.com/page", "https://example.com/page"},
	}

	for _, tt := range tests {
		got := normalizeURL(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDeduplicateLinks(t *testing.T) {
	links := []scrappermodel.Link{
		*scrappermodel.NewLink("https://a.com", scrappermodel.LINKTYPE_INTERNAL, 0),
		*scrappermodel.NewLink("https://b.com", scrappermodel.LINKTYPE_EXTERNAL, 10),
		*scrappermodel.NewLink("https://a.com", scrappermodel.LINKTYPE_INTERNAL, 20),
		*scrappermodel.NewLink("https://c.com", scrappermodel.LINKTYPE_EXTERNAL, 30),
		*scrappermodel.NewLink("https://b.com", scrappermodel.LINKTYPE_EXTERNAL, 40),
	}

	result := deduplicateLinks(links)
	if len(result) != 3 {
		t.Fatalf("expected 3 links, got %d", len(result))
	}
	hrefs := []string{result[0].GetHref(), result[1].GetHref(), result[2].GetHref()}
	expected := []string{"https://a.com", "https://b.com", "https://c.com"}
	for i, exp := range expected {
		if hrefs[i] != exp {
			t.Errorf("link %d: expected %s, got %s", i, exp, hrefs[i])
		}
	}
}

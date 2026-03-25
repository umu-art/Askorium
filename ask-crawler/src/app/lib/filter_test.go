package applib

import "testing"

func TestExtensionFilter(t *testing.T) {
	f := NewExtensionFilter()
	cases := []struct {
		url   string
		allow bool
	}{
		{"https://example.com/page", true},
		{"https://example.com/docs/api", true},
		{"https://example.com/image.jpg", false},
		{"https://example.com/image.PNG", false}, // case-insensitive
		{"https://example.com/video.mp4", false},
		{"https://example.com/archive.zip", false},
		{"https://example.com/style.css", false},
		{"https://example.com/app.js", false},
		// query string should not confuse extension detection
		{"https://example.com/page?img=test.jpg", true},
		// documents (pdf) are NOT blocked by ExtensionFilter — handled via enqueueDocuments
		{"https://example.com/doc.pdf", true},
	}
	for _, tc := range cases {
		got := f.Allow(URLCandidate{URL: tc.url})
		if got != tc.allow {
			t.Errorf("ExtensionFilter.Allow(%q) = %v, want %v", tc.url, got, tc.allow)
		}
	}
}

func TestTrapFilter(t *testing.T) {
	f := NewTrapFilter()
	cases := []struct {
		url   string
		allow bool
	}{
		{"https://example.com/docs", true},
		{"https://example.com/page/5", true},
		{"https://example.com/page/51", false},       // pagination > 50
		{"https://example.com/?page=100", false},     // pagination > 50
		{"https://example.com/?page=10", true},       // ok
		{"https://example.com/cdn-cgi/trace", false}, // cdn-cgi
		{"https://example.com/calendar/2024", false}, // calendar trap
		{"https://example.com/events/2024", false},   // calendar trap (events)
		{"https://example.com/?sort=date", false},    // sort trap
		{"https://example.com/?order=asc", false},    // order trap
		{"https://example.com/?filter=new", false},   // filter trap
		{"https://example.com/?q=search", true},      // normal query param
	}
	for _, tc := range cases {
		got := f.Allow(URLCandidate{URL: tc.url})
		if got != tc.allow {
			t.Errorf("TrapFilter.Allow(%q) = %v, want %v", tc.url, got, tc.allow)
		}
	}
}

func TestQueryParamFilter(t *testing.T) {
	f := NewQueryParamFilter(3)
	cases := []struct {
		url   string
		allow bool
	}{
		{"https://example.com/page", true},
		{"https://example.com/page?a=1&b=2&c=3", true},  // exactly 3
		{"https://example.com/page?a=1&b=2&c=3&d=4", false}, // 4 > 3
	}
	for _, tc := range cases {
		got := f.Allow(URLCandidate{URL: tc.url})
		if got != tc.allow {
			t.Errorf("QueryParamFilter.Allow(%q) = %v, want %v", tc.url, got, tc.allow)
		}
	}
}

func TestContentFilter(t *testing.T) {
	f := NewContentFilter()
	cases := []struct {
		url   string
		allow bool
	}{
		{"https://example.com/docs", true},
		{"https://example.com/login", false},
		{"https://example.com/signin", false},
		{"https://example.com/signup", false},
		{"https://example.com/logout", false},
		{"https://example.com/cart", false},
		{"https://example.com/checkout", false},
		{"https://example.com/print/article", false},
		{"https://example.com/feed.xml", false},
		{"https://example.com/rss.xml", false},
		{"https://example.com/sitemap.xml", false},
		{"https://example.com/robots.txt", false},
		// should not block paths that contain keyword as substring of segment
		{"https://example.com/blog/login-guide", true},
	}
	for _, tc := range cases {
		got := f.Allow(URLCandidate{URL: tc.url})
		if got != tc.allow {
			t.Errorf("ContentFilter.Allow(%q) = %v, want %v", tc.url, got, tc.allow)
		}
	}
}

func TestChainFilter_AllPass(t *testing.T) {
	chain := ChainFilter{NewExtensionFilter(), NewContentFilter()}
	c := URLCandidate{URL: "https://example.com/docs/api"}
	if !chain.Allow(c) {
		t.Error("ChainFilter should allow clean URL")
	}
}

func TestChainFilter_FirstBlocks(t *testing.T) {
	chain := ChainFilter{NewExtensionFilter(), NewContentFilter()}
	c := URLCandidate{URL: "https://example.com/image.jpg"}
	if chain.Allow(c) {
		t.Error("ChainFilter should block .jpg URL")
	}
}

func TestChainFilter_SecondBlocks(t *testing.T) {
	chain := ChainFilter{NewExtensionFilter(), NewContentFilter()}
	c := URLCandidate{URL: "https://example.com/login"}
	if chain.Allow(c) {
		t.Error("ChainFilter should block /login URL")
	}
}

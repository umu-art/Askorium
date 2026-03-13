package parser

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

const snippetWindow = 50

type contentExtractor struct{}

func (e *contentExtractor) Extract(doc *goquery.Document, pageURL string) *extractResult {
	meta := extractMetadata(doc)
	root := selectContentRoot(doc)

	pageHost := ""
	if u, err := url.Parse(pageURL); err == nil {
		pageHost = u.Host
	}

	var blocks []scrappermodel.ContentBlock
	var links []scrappermodel.Link
	var documents []scrappermodel.Document

	blockIdx := 0
	root.Find("h1, h2, h3, h4, h5, h6, p, li").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}

		blockType, headingLevel := classifyElement(s)
		htmlId := getOrGenerateID(s, blockIdx)

		block := *scrappermodel.NewContentBlock(blockType, text)
		block.SetHtmlId(htmlId)
		if headingLevel > 0 {
			block.SetHeadingLevel(headingLevel)
		}
		blocks = append(blocks, block)

		s.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if !exists {
				return
			}
			href = strings.TrimSpace(href)
			if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
				return
			}

			resolved := resolveURL(href, pageURL)

			if isDocumentURL(resolved) {
				doc := newDocumentFromAnchor(resolved, a.Text())
				documents = append(documents, doc)
				return
			}

			anchorText := strings.TrimSpace(a.Text())
			position := calcPosition(text, anchorText)
			linkType := classifyLink(resolved, pageHost)

			link := *scrappermodel.NewLink(resolved, linkType, position)
			link.SetContextBlockId(htmlId)
			if anchorText != "" {
				link.SetAnchorText(anchorText)
			}
			snippet := makeSnippet(text, position, snippetWindow)
			if snippet != "" {
				link.SetSnippet(snippet)
			}
			links = append(links, link)
		})

		s.Find("img[src]").Each(func(_ int, img *goquery.Selection) {
			src, exists := img.Attr("src")
			if !exists {
				return
			}
			src = strings.TrimSpace(src)
			if src == "" {
				return
			}
			resolved := resolveURL(src, pageURL)
			alt := img.AttrOr("alt", "")
			imgDoc := newDocumentFromImg(resolved, alt)
			documents = append(documents, imgDoc)
		})

		blockIdx++
	})

	return &extractResult{
		Meta:      meta,
		Blocks:    blocks,
		Links:     links,
		Documents: documents,
	}
}

// --- content root selection ---

func selectContentRoot(doc *goquery.Document) *goquery.Selection {
	if main := doc.Find("main"); main.Length() > 0 {
		return main.First()
	}
	if article := doc.Find("article"); article.Length() > 0 {
		return article.First()
	}
	return doc.Find("body")
}

// --- element classification ---

func classifyElement(s *goquery.Selection) (scrappermodel.ContentBlockType, int32) {
	tag := goquery.NodeName(s)
	switch tag {
	case "h1":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 1
	case "h2":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 2
	case "h3":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 3
	case "h4":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 4
	case "h5":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 5
	case "h6":
		return scrappermodel.CONTENTBLOCKTYPE_HEADING, 6
	case "li":
		return scrappermodel.CONTENTBLOCKTYPE_LIST_ITEM, 0
	default:
		return scrappermodel.CONTENTBLOCKTYPE_PARAGRAPH, 0
	}
}

func getOrGenerateID(s *goquery.Selection, idx int) string {
	if id, exists := s.Attr("id"); exists && id != "" {
		return id
	}
	return fmt.Sprintf("block-%d", idx)
}

// --- metadata extraction ---

func extractMetadata(doc *goquery.Document) metadata {
	return metadata{
		Title:        getTitle(doc),
		Description:  getDescription(doc),
		PreviewURL:   getPreviewURL(doc),
		IconURL:      getIconURL(doc),
		Language:     getLanguage(doc),
		LastModified: getLastModified(doc),
	}
}

func getTitle(doc *goquery.Document) string {
	title := strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find(`meta[property="og:title"]`).AttrOr("content", ""))
	}
	if title == "" {
		title = "Untitled"
	}
	return title
}

func getDescription(doc *goquery.Document) *string {
	desc := strings.TrimSpace(doc.Find(`meta[name="description"]`).AttrOr("content", ""))
	if desc == "" {
		desc = strings.TrimSpace(doc.Find(`meta[property="og:description"]`).AttrOr("content", ""))
	}
	return strPtr(desc)
}

func getPreviewURL(doc *goquery.Document) *string {
	return strPtr(strings.TrimSpace(doc.Find(`meta[property="og:image"]`).AttrOr("content", "")))
}

func getIconURL(doc *goquery.Document) *string {
	iconURL := strings.TrimSpace(doc.Find(`link[rel="icon"]`).AttrOr("href", ""))
	if iconURL == "" {
		iconURL = strings.TrimSpace(doc.Find(`link[rel="shortcut icon"]`).AttrOr("href", ""))
	}
	return strPtr(iconURL)
}

func getLanguage(doc *goquery.Document) *string {
	return strPtr(strings.TrimSpace(doc.Find("html").AttrOr("lang", "")))
}

func getLastModified(doc *goquery.Document) *time.Time {
	s := strings.TrimSpace(doc.Find(`meta[property="article:modified_time"]`).AttrOr("content", ""))
	if s == "" {
		s = strings.TrimSpace(doc.Find(`meta[http-equiv="last-modified"]`).AttrOr("content", ""))
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// --- URL helpers ---

func resolveURL(href string, baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

func classifyLink(href string, pageHost string) scrappermodel.LinkType {
	u, err := url.Parse(href)
	if err != nil || u.Host == "" {
		return scrappermodel.LINKTYPE_INTERNAL
	}
	if strings.EqualFold(u.Host, pageHost) {
		return scrappermodel.LINKTYPE_INTERNAL
	}
	return scrappermodel.LINKTYPE_EXTERNAL
}

func calcPosition(blockText string, anchorText string) int32 {
	idx := strings.Index(blockText, anchorText)
	if idx < 0 {
		return 0
	}
	return int32(len([]rune(blockText[:idx])))
}

func makeSnippet(blockText string, position int32, window int) string {
	runes := []rune(blockText)
	pos := int(position)
	start := max(pos-window, 0)
	end := min(pos+window, len(runes))
	return string(runes[start:end])
}

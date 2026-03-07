package impl

import (
	"ask-parser/internal/parser"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

type contentExtractor struct{}

func NewContentExtractor() parser.ContentExtractor {
	return &contentExtractor{}
}

func (e *contentExtractor) Extract(doc *goquery.Document, pageURL string) ([]scrappermodel.ContentBlock, []scrappermodel.Link, []scrappermodel.Document) {
	root := selectContentRoot(doc)

	var blocks []scrappermodel.ContentBlock
	var links []scrappermodel.Link
	var documents []scrappermodel.Document

	pageHost := ""
	if u, err := url.Parse(pageURL); err == nil {
		pageHost = u.Host
	}

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
			if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
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

	return blocks, links, documents
}

func selectContentRoot(doc *goquery.Document) *goquery.Selection {
	if main := doc.Find("main"); main.Length() > 0 {
		return main.First()
	}
	if article := doc.Find("article"); article.Length() > 0 {
		return article.First()
	}
	return doc.Find("body")
}

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

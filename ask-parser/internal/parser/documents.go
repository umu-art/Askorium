package parser

import (
	"path"
	"strings"

	scrappermodel "github.com/omo-ri/askorium/go-parser-api"
)

var documentExtensions = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".zip":  "application/zip",
	".rar":  "application/vnd.rar",
}

var imageExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
	".ico":  "image/x-icon",
}

func isDocumentURL(href string) bool {
	ext := strings.ToLower(path.Ext(href))
	_, ok := documentExtensions[ext]
	return ok
}

func mimeFromExtension(ext string) string {
	ext = strings.ToLower(ext)
	if mime, ok := documentExtensions[ext]; ok {
		return mime
	}
	if mime, ok := imageExtensions[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

func newDocumentFromAnchor(href string, anchorText string) scrappermodel.Document {
	ext := strings.ToLower(path.Ext(href))
	doc := *scrappermodel.NewDocument(href, mimeFromExtension(ext))
	text := strings.TrimSpace(anchorText)
	if text != "" {
		doc.SetDescription(text)
		doc.SetDescriptionSource(scrappermodel.DOCUMENTDESCRIPTIONSOURCETYPE_PARAGRAPH)
	}
	return doc
}

func newDocumentFromImg(src string, alt string) scrappermodel.Document {
	ext := strings.ToLower(path.Ext(src))
	mime := mimeFromExtension(ext)
	if mime == "application/octet-stream" {
		mime = "image/unknown"
	}
	doc := *scrappermodel.NewDocument(src, mime)
	alt = strings.TrimSpace(alt)
	if alt != "" {
		doc.SetDescription(alt)
		doc.SetDescriptionSource(scrappermodel.DOCUMENTDESCRIPTIONSOURCETYPE_PARAGRAPH)
	}
	return doc
}

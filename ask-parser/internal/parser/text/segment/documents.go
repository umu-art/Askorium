package segment

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

func IsDocumentURL(href string) bool {
	ext := strings.ToLower(path.Ext(href))
	_, ok := documentExtensions[ext]
	return ok
}

func IsImageURL(href string) bool {
	ext := strings.ToLower(path.Ext(href))
	_, ok := imageExtensions[ext]
	return ok
}

func MimeFromExtension(ext string) string {
	ext = strings.ToLower(ext)
	if mime, ok := documentExtensions[ext]; ok {
		return mime
	}
	if mime, ok := imageExtensions[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

func NewDocumentFromAnchor(href, anchorText string) scrappermodel.Document {
	ext := strings.ToLower(path.Ext(href))
	doc := *scrappermodel.NewDocument(href, MimeFromExtension(ext))
	text := strings.TrimSpace(anchorText)
	switch {
	case text != "":
		doc.SetDescription(text)
		doc.SetDescriptionSource(scrappermodel.DOCUMENTDESCRIPTIONSOURCETYPE_PARAGRAPH)
	default:
		name := strings.TrimSuffix(path.Base(href), ext)
		if name != "" && name != "." {
			doc.SetDescription(name)
			doc.SetDescriptionSource(scrappermodel.DOCUMENTDESCRIPTIONSOURCETYPE_FILENAME)
		}
	}
	return doc
}

func NewDocumentFromImg(src, alt string) scrappermodel.Document {
	ext := strings.ToLower(path.Ext(src))
	mime := MimeFromExtension(ext)
	if mime == "application/octet-stream" {
		mime = "image/unknown"
	}
	doc := *scrappermodel.NewDocument(src, mime)
	alt = strings.TrimSpace(alt)
	if alt != "" {
		doc.SetDescription(alt)
		doc.SetDescriptionSource(scrappermodel.DOCUMENTDESCRIPTIONSOURCETYPE_ALT_TEXT)
	}
	return doc
}

package parser

import (
	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
)

type Parser interface {
	Parse(rawHTML string, pageURL string) (*scrappermodel.ScrappedPage, error)
}

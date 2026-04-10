package scraper

import "errors"

var (
	ErrScrape = errors.New("scrape failed")
	ErrParse  = errors.New("failed to parse events page")
)

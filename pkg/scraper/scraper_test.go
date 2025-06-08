package scraper

import (
	"testing"
	"time"
)

func TestScrapeEvents(t *testing.T) {
	_, err := ScrapeEvents()
	if err != nil {
		t.Fatalf("failed to scrape events, possible changes to aac website: %s", err)
	}
}

func TestDateVerification(t *testing.T) {
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)

	if !isToday(today) {
		t.Fatalf("Failed to properly parse time, did not identify today")
	}
	if isToday(tomorrow) {
		t.Fatalf("Failed to properly parse time, identified tomorrow as today ")
	}
}

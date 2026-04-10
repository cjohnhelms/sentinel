//go:build integration

package scraper

import (
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/event"
)

var blackhole []event.Event

func TestScrape_integration(t *testing.T) {
	events, err := Scrape(NewHTTPFetcher(), time.Now())
	if err != nil {
		t.Fatalf("failed to scrape events: %v", err)
	}
	if len(events) == 0 {
		t.Log("scrape succeeded but no events found today")
	}
}

func BenchmarkScrapeEvents(b *testing.B) {
	b.ReportAllocs()
	f := NewHTTPFetcher()
	target := time.Date(2025, 7, 22, 0, 0, 0, 0, time.Local)
	for i := 0; i < b.N; i++ {
		events, err := Scrape(f, target)
		if err != nil {
			b.Fatalf("failed to scrape events: %v", err)
		}
		blackhole = events
	}
}

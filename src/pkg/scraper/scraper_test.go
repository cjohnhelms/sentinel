package scraper

import (
	"github.com/cjohnhelms/sentinel/pkg/event"
	"testing"
	"time"
)

var blackhole []event.Event

func TestScrapeEvents(t *testing.T) {
	events, err := scrape(time.Now())
	if err != nil {
		t.Fatalf("failed to scrape events, possible changes to aac website: %s", err)
	}
	if len(events) == 0 {
		t.Logf("scrape succeeded but no events found")
	}
}

func BenchmarkScrapeEvents(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		events, err := scrape(time.Date(2025, 7, 22, 0, 0, 0, 0, time.Local))
		if err != nil {
			b.Fatalf("failed to scrape events: %s", err)
		}
		blackhole = events
	}
}

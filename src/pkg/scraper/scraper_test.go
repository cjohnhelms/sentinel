package scraper

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/event"
)

type fakeFetcher struct {
	body string
	err  error
}

func (f *fakeFetcher) Fetch(_ string) (io.ReadCloser, error) {
	if f.err != nil {
		return nil, f.err
	}
	return io.NopCloser(strings.NewReader(f.body)), nil
}

var target = time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)

const fixtureWithEvent = `<!DOCTYPE html>
<html><body>
<div class="info clearfix">
  <div class="date">Apr 9, 2026 - 7:30PM</div>
  <h3><a href="/events/detail/test">Dallas Stars vs. Colorado Avalanche</a></h3>
</div>
<div class="info clearfix">
  <div class="date">Apr 10, 2026 - 7PM</div>
  <h3><a href="/events/detail/test2">Mavericks vs. Warriors</a></h3>
</div>
</body></html>`

const fixtureNoEvents = `<!DOCTYPE html>
<html><body>
<div class="info clearfix">
  <div class="date">Apr 10, 2026 - 7PM</div>
  <h3><a href="/events/detail/test">Mavericks vs. Warriors</a></h3>
</div>
</body></html>`

const fixtureMalformedDate = `<!DOCTYPE html>
<html><body>
<div class="info clearfix">
  <div class="date">not a valid date</div>
  <h3><a href="/events/detail/test">Bad Event</a></h3>
</div>
<div class="info clearfix">
  <div class="date">Apr 9, 2026 - 7:30PM</div>
  <h3><a href="/events/detail/test2">Valid Event</a></h3>
</div>
</body></html>`

func TestScrape_returnsTodayEvents(t *testing.T) {
	f := &fakeFetcher{body: fixtureWithEvent}
	got, err := Scrape(f, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	if got[0].Title != "Dallas Stars vs. Colorado Avalanche" {
		t.Errorf("title = %q", got[0].Title)
	}
}

func TestScrape_fetchError(t *testing.T) {
	f := &fakeFetcher{err: errors.New("connection refused")}
	_, err := Scrape(f, target)
	if !errors.Is(err, ErrScrape) {
		t.Errorf("expected ErrScrape, got %v", err)
	}
}

func FuzzParseEvents(f *testing.F) {
	f.Add(fixtureWithEvent)
	f.Add(fixtureNoEvents)
	f.Add(fixtureMalformedDate)

	f.Fuzz(func(t *testing.T, html string) {
		_, _ = parseEvents(strings.NewReader(html), target)
	})
}

func TestParseEvents(t *testing.T) {
	tests := []struct {
		name string
		html string
		want []event.Event
	}{
		{
			name: "returns event matching target date",
			html: fixtureWithEvent,
			want: []event.Event{
				{Title: "Dallas Stars vs. Colorado Avalanche", When: "2026-04-09 7:30PM"},
			},
		},
		{
			name: "returns nil when no events on target date",
			html: fixtureNoEvents,
			want: nil,
		},
		{
			name: "skips malformed date, returns valid event",
			html: fixtureMalformedDate,
			want: []event.Event{
				{Title: "Valid Event", When: "2026-04-09 7:30PM"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEvents(strings.NewReader(tt.html), target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d events, want %d: %v", len(got), len(tt.want), got)
			}
			for i, e := range got {
				if e != tt.want[i] {
					t.Errorf("event[%d] = %+v, want %+v", i, e, tt.want[i])
				}
			}
		})
	}
}

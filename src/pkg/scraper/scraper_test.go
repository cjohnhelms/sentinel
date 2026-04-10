package scraper

import (
	"strings"
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/event"
)

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

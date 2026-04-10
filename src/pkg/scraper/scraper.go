package scraper

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cjohnhelms/sentinel/pkg/event"
)

// Scrape fetches the AAC events page and returns events matching target date.
// Signature will gain a Fetcher parameter in Task 3; for now uses net/http directly.
func Scrape(target time.Time) ([]event.Event, error) {
	resp, err := http.Get("https://www.americanairlinescenter.com/events")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScrape, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrScrape, resp.StatusCode)
	}
	return parseEvents(resp.Body, target)
}

func parseEvents(r io.Reader, target time.Time) ([]event.Event, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}

	layouts := []string{
		"Jan 2, 2006 3:04PM",
		"Jan 2, 2006 3PM",
	}

	var events []event.Event
	doc.Find("div.info.clearfix").Each(func(_ int, s *goquery.Selection) {
		dt := strings.TrimSpace(s.Find("div.date").Text())
		title := strings.TrimSpace(s.Find("h3 a").Text())

		parts := strings.SplitN(dt, "-", 2)
		if len(parts) != 2 {
			slog.Warn("unexpected date format", "date", dt)
			return
		}
		combined := strings.TrimSpace(parts[0]) + " " + strings.TrimSpace(parts[1])

		var eventTime time.Time
		var parseErr error
		for _, layout := range layouts {
			eventTime, parseErr = time.Parse(layout, combined)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			slog.Warn("failed to parse event date", "combined", combined)
			return
		}

		if eventTime.Year() == target.Year() &&
			eventTime.Month() == target.Month() &&
			eventTime.Day() == target.Day() {
			events = append(events, event.Event{
				Title: title,
				When:  eventTime.Format("2006-01-02 3:04PM"),
			})
		}
	})

	return events, nil
}

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

// Fetcher abstracts the HTTP layer so Scrape can be tested without network calls.
type Fetcher interface {
	Fetch(url string) (io.ReadCloser, error)
}

// HTTPFetcher is the production Fetcher backed by net/http.
type HTTPFetcher struct {
	client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HTTPFetcher) Fetch(url string) (io.ReadCloser, error) {
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// Scrape fetches the AAC events page using f and returns events on target date.
func Scrape(f Fetcher, target time.Time) ([]event.Event, error) {
	body, err := f.Fetch("https://www.americanairlinescenter.com/events")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScrape, err)
	}
	defer body.Close()
	return parseEvents(body, target)
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

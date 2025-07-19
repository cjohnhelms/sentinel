package scraper

import (
	"fmt"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"github.com/gocolly/colly"
	"log/slog"
	"strings"
	"time"
)

func scrape(target time.Time) ([]event.Event, error) {
	var todaysEvents []event.Event

	c := colly.NewCollector(colly.AllowedDomains("www.americanairlinescenter.com"))
	c.OnResponse(func(r *colly.Response) {
		slog.Info("visited site")
	})
	c.OnError(func(r *colly.Response, err error) {
		slog.Info(fmt.Sprintf("failed to scrape page: %s", err))
	})
	c.SetRequestTimeout(5 * time.Second)
	c.OnHTML("div.info.clearfix", func(e *colly.HTMLElement) {
		// collect raw data of all events on the page
		dt := e.ChildText("div.date")
		title := e.ChildText("h3 a")

		parts := strings.Split(dt, "-")
		if len(parts) != 2 {
			fmt.Println("Unexpected format")
			return
		}
		dateStr := strings.TrimSpace(parts[0])
		timeStr := strings.TrimSpace(parts[1])
		combined := fmt.Sprintf("%s %s", dateStr, timeStr)

		layouts := [2]string{
			"Jan 2, 2006 3:04PM", // for e.g. 6:30PM
			"Jan 2, 2006 3PM",    // for e.g. 6PM
		}

		var eventTime time.Time
		var err error
		for _, layout := range layouts {
			eventTime, err = time.Parse(layout, combined)
			if err == nil {
				break
			}
		}
		if err != nil {
			slog.Warn("failed to parse date")
		}

		if eventTime.Year() == target.Year() && eventTime.Month() == target.Month() && eventTime.Day() == target.Day() {
			todaysEvents = append(todaysEvents, event.Event{
				Title: title,
				When:  eventTime.Format("2006-01-02 3:04PM"),
			})
		}
	})

	// initiate page visit and collect raw data
	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		return nil, fmt.Errorf("failed to scrape: %s", err)
	}

	return todaysEvents, nil
}

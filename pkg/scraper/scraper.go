package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/email"
	"github.com/cjohnhelms/sentinel/pkg/structs"
)

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

func scrape(logger *slog.Logger) structs.Event {
	var when time.Time

	var event = structs.Event{
		Title: "No event today",
		When:  when,
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.americanairlinescenter.com"))

	c.OnRequest(func(r *colly.Request) {
		logger.Info(fmt.Sprintf("Visiting: %s", r.URL.String()))
	})
	c.OnResponse(func(r *colly.Response) {
		logger.Info(fmt.Sprintf("Visited: %s", r.Request.URL.String()))
	})
	c.OnError(func(r *colly.Response, err error) {
		logger.Info(fmt.Sprintf("Failed to scrape page: %s", err))
	})
	c.OnHTML("div.info.clearfix", func(e *colly.HTMLElement) {
		dt := e.ChildText("div.date")
		title := e.ChildText("h3 a")

		a := strings.Split(dt, "-")
		for i := range a {
			a[i] = strings.TrimSpace(a[i])
		}
		a[1] = strings.ToLower(a[1])

		dayTime := strings.Join(a, " ")

		when, err := time.Parse("Jan 2, 2006 3:04pm", dayTime)
		if err != nil {
			logger.Debug("Trying other parse layout")
			when, err = time.Parse("Jan 2, 2006 3pm", dayTime)
			if err != nil {
				logger.Error(fmt.Sprintf("both time layouts failed: %s", title))
			}
		}

		if isToday(when) {
			event.When = when
			event.Title = title
			logger.Info(fmt.Sprintf("Found event today, queueing email: %+v", event))
		}
	})
	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		logger.Error(fmt.Sprintf("Failed: %s\n", err))
	}

	return event
}

func Run(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := cfg.Logger

	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), cfg.ScrapeHour, cfg.ScrapeMin, 0, 0, now.Location())

		if now.After(next) {
			// if past scrape time, schedule it for the next day
			next = next.Add(24 * time.Hour)
		}

		duration := time.Until(next)
		logger.Debug(fmt.Sprintf("Scraping again in %v", duration))
		timer := time.NewTimer(duration)

		select {
		case <-ctx.Done():
			logger.Info("Scraper routine recieved signal, killing")
			return
		case <-timer.C:
			event := scrape(logger)
			if isToday(event.When) {
				wg.Add(1)
				go email.ScheduleEmail(ctx, cfg, wg, event)
			}
		}
	}

}

package scraper

import (
	"context"
	"errors"
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

func parseDt(dt string) (string, string, error) {
	cleaned := strings.Split(strings.Join(strings.Fields(dt), " "), " - ")
	dateStr := cleaned[0]
	timeStr := cleaned[1]
	date, err := time.Parse("Jan 2, 2006", dateStr)
	if err != nil {
		return "", "", errors.New("Could not parse date: " + dateStr)
	}
	isoDate := date.Format("2006-01-02")
	return isoDate, timeStr, nil
}

func scrape(logger *slog.Logger) structs.Event {
	today := time.Now().Format("2006-01-02")

	var event = structs.Event{
		Title: "No event today",
		Start: "",
		Date:  "",
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

		isoDate, timeStr, err := parseDt(dt)
		if err != nil {
			logger.Error(err.Error())
		}

		if isoDate == today {
			event.Date = isoDate
			event.Start = timeStr
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
			// if past 2 PM, schedule it for the next day
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
			today := time.Now().Format("2006-01-02")
			if event.Date == today {
				wg.Add(1)
				go email.ScheduleEmail(ctx, cfg, wg, event)
			}
		}
	}

}

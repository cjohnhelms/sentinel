package scraper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"

	log "github.com/cjohnhelms/sentinel/pkg/logging"
)

type Event struct {
	Date  string
	Start string
	Title string
}

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

func Scrape() Event {
	today := time.Now().Format("2006-01-02")

	var event = Event{
		Title: "No event today",
		Start: "",
		Date:  "",
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.americanairlinescenter.com"))

	c.OnRequest(func(r *colly.Request) {
		log.Info(fmt.Sprintf("Visiting: %s", r.URL.String()))
	})
	c.OnResponse(func(r *colly.Response) {
		log.Info(fmt.Sprintf("Visited: %s", r.Request.URL.String()))
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Info(fmt.Sprintf("Failed to scrape page: %s", err))
	})
	c.OnHTML("div.info.clearfix", func(e *colly.HTMLElement) {
		dt := e.ChildText("div.date")
		title := e.ChildText("h3 a")

		isoDate, timeStr, err := parseDt(dt)
		if err != nil {
			log.Error(err.Error())
		}

		if isoDate == today {
			event = Event{Date: isoDate, Start: timeStr, Title: title}
			log.Info(fmt.Sprintf("Found event today: %s", event.Title))
		}
	})
	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		log.Error(fmt.Sprintf("Failed: %s\n", err))
	}

	return event
}

func FetchEvents(ctx context.Context, wg *sync.WaitGroup, ch chan<- Event) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Info("Killing scraper routine")
			return
		default:
			// scrape events
			event := Scrape()
			log.Debug("Sending event in channel")
			ch <- event

			// Get the current time
			now := time.Now()

			// Calculate the next 2 PM
			nextScrape := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
			if now.After(nextScrape) {
				// If it’s already past 2 PM, schedule it for the next day
				nextScrape = nextScrape.Add(24 * time.Hour)
			}

			// Calculate the duration until the next 2 PM
			duration := nextScrape.Sub(now)

			// Sleep until the next 2 PM
			time.Sleep(duration)
		}
	}
}

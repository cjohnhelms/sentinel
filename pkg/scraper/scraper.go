package scraper

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
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

	var event Event

	c := colly.NewCollector(
		colly.AllowedDomains("www.americanairlinescenter.com"))

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting: ", r.URL.String())
	})
	c.OnResponse(func(r *colly.Response) {
		log.Println("Visited: ", r.Request.URL.String())
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Failed to scrape page: ", err)
	})
	c.OnHTML("div.info.clearfix", func(e *colly.HTMLElement) {
		dt := e.ChildText("div.date")
		title := e.ChildText("h3 a")

		isoDate, timeStr, err := parseDt(dt)
		if err != nil {
			log.Println(err)
		}

		if isoDate == today {
			event = Event{Date: isoDate, Start: timeStr, Title: title}
			log.Println("Found event today:", event.Title)
		}
	})
	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		log.Printf("Failed: %s\n", err)
	}

	return event
}

func FetchEvents(ch chan<- Event) {
	for {
		// scrape events
		event := Scrape()
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

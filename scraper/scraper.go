package scraper

import (
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"strings"
	"time"
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

func Scrape() []Event {
	var events []Event

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
		event := Event{Date: isoDate, Start: timeStr, Title: title}
		events = append(events, event)
	})
	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		log.Printf("Failed: %s\n", err)
	}

	var output []string
	output = append(output, "Events:")
	for _, event := range events {
		output = append(output, fmt.Sprintf(" {%s, %s} ", event.Title, event.Date))
	}
	log.Println(output)
	return events
}

func FetchEvents(ch chan<- []Event) {
	for {
		// Get the current time
		now := time.Now()

		// Calculate the next 2 PM
		next2PM := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, now.Location())
		if now.After(next2PM) {
			// If it’s already past 2 PM, schedule it for the next day
			next2PM = next2PM.Add(24 * time.Hour)
		}

		// Calculate the duration until the next 2 PM
		duration := next2PM.Sub(now)

		// Sleep until the next 2 PM
		time.Sleep(duration)

		events := Scrape()
		ch <- events
	}
}

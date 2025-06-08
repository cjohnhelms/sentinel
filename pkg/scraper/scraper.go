package scraper

import (
	"fmt"
	"github.com/gocolly/colly"
	"log/slog"
	"strings"
	"time"
)

type Event struct {
	Title string
	When  string
}

func ScrapeEvents() ([]Event, error) {
	slog.Info("starting scrape task")

	todaysEvents, err := scrape()
	if err != nil {
		return nil, err
	}
	if len(todaysEvents) == 0 {
		slog.Debug("scrape function returned nil, must be no events today")
		return todaysEvents, nil
	} else {
		slog.Info("events found for today", len(todaysEvents))
	}
	return todaysEvents, nil
}

func scrape() ([]Event, error) {
	var rawData []struct {
		dateTime string
		title    string
	}
	var todaysEvents []Event

	c := colly.NewCollector(
		colly.AllowedDomains("www.americanairlinescenter.com"))

	c.OnRequest(func(r *colly.Request) {
		slog.Info(fmt.Sprintf("visiting: %s", r.URL.String()))
	})
	c.OnResponse(func(r *colly.Response) {
		slog.Info(fmt.Sprintf("visited: %s", r.Request.URL.String()))
	})
	c.OnError(func(r *colly.Response, err error) {
		slog.Info(fmt.Sprintf("failed to scrape page: %s", err))
	})
	c.SetRequestTimeout(10 * time.Second)
	c.OnHTML("div.info.clearfix", func(e *colly.HTMLElement) {
		// collect raw data of all events on the page
		dt := e.ChildText("div.date")
		title := e.ChildText("h3 a")

		rawData = append(rawData, struct {
			dateTime string
			title    string
		}{dateTime: dt, title: title})
	})

	// initiate page visit and collect raw data
	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		return nil, fmt.Errorf("failed to scrape: %s", err)
	}

	// process data and create events if there are any on the current date
	for _, data := range rawData {
		dt := data.dateTime
		title := data.title

		a := strings.Split(dt, "-")
		for i := range a {
			a[i] = strings.TrimSpace(a[i])
		}
		a[1] = strings.ToLower(a[1])

		dayTime := strings.Join(a, " ")

		when, err := time.Parse("Jan 2, 2006 3:04pm", dayTime)
		if err != nil {
			slog.Debug("trying other parse layout")
			when, err = time.Parse("Jan 2, 2006 3pm", dayTime)
			if err != nil {
				return nil, fmt.Errorf("both time layouts failed: %s", title)
			}
		}

		if isToday(when) {
			event := Event{}
			event.When = when.Format("3:04 PM")
			event.Title = title
			todaysEvents = append(todaysEvents, event)
		}
	}

	return todaysEvents, nil
}

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

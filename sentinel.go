package main

import (
	"fmt"
	"github.com/gocolly/colly"
)

type Event struct {
	date string
	name string
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.americanairlinescenter.com"))

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited: ", r.Request.URL.String())
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Failed to scrape page: ", err)
	})
	c.OnHTML("div.info.clearfix", func(e *colly.HTMLElement) {
		date := e.ChildText("div.date")
		title := e.ChildText("h3 a")

		fmt.Println(date, title)
	})

	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		fmt.Println("Failed: ", err)
	}
}

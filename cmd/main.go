package main

import (
	"os"
	"sentinel/pkg/display"
	"sentinel/pkg/notify"
	"sentinel/pkg/scraper"
)

func main() {
	var recipients = [2]string{os.Getenv("CHRIS_NUMBER"), os.Getenv("KYLE_NUMBER")}

	initEvent := scraper.Scrape()
	if initEvent.Title != "" {
		display.Write(initEvent)
	}

	ch := make(chan scraper.Event, 1)
	go scraper.FetchEvents(ch)
	go notify.Notify(ch, recipients)

	select {}
}

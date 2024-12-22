package main

import (
	"log"
	"os"
	"sentinel/pkg/notify"
	"sentinel/pkg/scraper"
	"time"
)

func main() {
	log.Println("Service starting")
	var recipients = [2]string{os.Getenv("CHRIS_NUMBER"), os.Getenv("KYLE_NUMBER")}

	ch := make(chan scraper.Event, 1)
	go scraper.FetchEvents(ch)
	go notify.Notify(ch, recipients)

	test1 := scraper.Event{
		Title: "short title",
		Start: "8pm",
		Date:  "2024-12-22",
	}
	test2 := scraper.Event{
		Title: "this title is really long and has many characters",
		Start: "8pm",
		Date:  "2024-12-22",
	}

	time.Sleep(5 + time.Second)
	ch <- test1
	time.Sleep(5 * time.Second)
	ch <- test2

	select {}
}

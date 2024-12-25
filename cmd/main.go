package main

import (
	"log"
	"sentinel/pkg/display"
	"sentinel/pkg/scraper"
)

func main() {
	log.Println("Service starting")
	//var recipients = [2]string{os.Getenv("CHRIS_NUMBER"), os.Getenv("KYLE_NUMBER")}

	ch := make(chan scraper.Event, 1)
	go scraper.FetchEvents(ch)
	//go notify.Notify(ch, recipients)
	go display.Update(ch)

	select {}
}

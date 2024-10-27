package main

import (
	"sentinel/notify"
	"sentinel/scraper"
)

//"8178452177@vtext.com"

var recipients = [1]string{"5129631386@txt.att.net"}

func main() {
	ch := make(chan []scraper.Event, 1)
	go scraper.FetchEvents(ch)
	go notify.Notify(ch, recipients)

	select {}
}

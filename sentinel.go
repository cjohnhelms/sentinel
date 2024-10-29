package main

import (
	"sentinel/pkg/notify"
	"sentinel/pkg/scraper"
	"sentinel/pkg/ui"
)

//"8178452177@vtext.com"

var recipients = [2]string{"5129631386@txt.att.net", "8178452177@vtext.com"}

func main() {
	ch := make(chan []scraper.Event, 1)
	go scraper.FetchEvents(ch)
	go notify.Notify(ch, recipients)
	go ui.Update(ch)

	select {}
}

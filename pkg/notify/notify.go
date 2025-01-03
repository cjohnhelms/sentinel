package notify

import (
	"fmt"
	"log"
	"log/slog"
	"sentinel/pkg/scraper"
	"time"
)

func SendEmail(event scraper.Event, recipient string) error {
	b := fmt.Sprint("AAC Event: %s\nStart time: %s", event.Title, event.Start)

}

func Notify(ch <-chan scraper.Event, recipients [2]string) {
	for {
		// Get the current time
		now := time.Now()

		// Calculate the next 2 PM
		nextNotify := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, now.Location())

		if now.After(nextNotify) {
			// If it’s already past 2 PM, schedule it for the next day
			nextNotify = nextNotify.Add(24 * time.Hour)
		}

		// Calculate the duration until the next 2 PM
		duration := nextNotify.Sub(now)

		// Sleep until the next 2 PM
		time.Sleep(duration)

		select {
		case data := <-ch:
			event := data
			if event.Start == "" {
				slog.Info("No event found today")
			} else {
				for _, recipient := range recipients {
					if err := SendText(event, recipient); err != nil {
						log.Println(err)
					}
				}
			}
		default:
			slog.Info("No new events in the channel")
		}
	}
}

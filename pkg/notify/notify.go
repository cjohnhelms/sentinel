package notify

import (
	"fmt"
	"log"
	"sentinel/pkg/scraper"
	"time"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

func SendText(event scraper.Event, recipient string) error {
	b := fmt.Sprint("AAC Event: %s\nStart time: %s", event.Title, event.Start)
	client := twilio.NewRestClient()

	params := &api.CreateMessageParams{}
	params.SetBody(b)
	params.SetFrom("+15127069578")
	params.SetTo(recipient)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		log.Println(err)
	} else {
		if resp.Body != nil {
			fmt.Println(*resp.Body)
		} else {
			fmt.Println(resp.Body)
		}
	}
	return nil
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
			if event.Title == "" {
				log.Println("No event found today")
			} else {
				for _, recipient := range recipients {
					if err := SendText(event, recipient); err != nil {
						log.Println(err)
					}
				}
			}
		default:
			log.Println("No new events in the channel")
		}
	}
}

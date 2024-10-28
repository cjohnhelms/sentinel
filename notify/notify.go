package notify

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"sentinel/scraper"
	"time"
)

func SendMail(body string, recipient string) error {
	from := "cjohnhelms@gmail.com"
	pass := "dnax gtze bvvy fbao"
	to := recipient

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Sentinel Report\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))
	if err != nil {
		return errors.New("smtp error: " + err.Error())
	}
	log.Println("Successfully sent to " + to)
	return nil
}

func Notify(ch <-chan []scraper.Event, recipients [1]string) {
	for {
		// Get the current time
		now := time.Now()

		// Calculate the next 2 PM
		next2PM := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, now.Location())
		if now.After(next2PM) {
			// If it’s already past 2 PM, schedule it for the next day
			next2PM = next2PM.Add(24 * time.Hour)
		}

		// Calculate the duration until the next 2 PM
		duration := next2PM.Sub(now)

		// Sleep until the next 2 PM
		time.Sleep(duration)

		today := time.Now().Format("2006-01-02")
		select {
		case data := <-ch:
			for _, event := range data {
				if event.Date == today {
					body := fmt.Sprintf("\n%s - %s\nUse Harry Hines Blvd", event.Title, event.Start)
					for _, recipient := range recipients {
						err := SendMail(body, recipient)
						if err != nil {
							log.Println("Error sending notification: " + err.Error())
						}
					}
				}
			}
		default:
			log.Println("No events in the channel")
		}
	}
}

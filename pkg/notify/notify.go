package notify

import (
	"errors"
	"fmt"
	"net/smtp"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/config"
	log "github.com/cjohnhelms/sentinel/pkg/logging"
	"github.com/cjohnhelms/sentinel/pkg/scraper"
)

type Email struct {
	Password  string
	FromName  string
	FromEmail string
	ToEmail   string
	Subject   string
	Message   string
}

func (em *Email) Send() error {
	auth := smtp.PlainAuth("", em.FromEmail, em.Password, "smtp.gmail.com")
	msg := fmt.Sprintf("From: %s %s\nTo: %s\nSubject: %s\n\n%s", em.FromName, em.FromEmail, em.ToEmail, em.Subject, em.Message)

	err := smtp.SendMail("smtp.gmail.com:587",
		auth,
		em.FromEmail,
		[]string{em.ToEmail},
		[]byte(msg),
	)
	if err != nil {
		return errors.New("smtp error: " + err.Error())
	}
	log.Info("Successfully sent to " + em.ToEmail)
	return nil
}

func Notify(ch <-chan scraper.Event, cfg *config.Config) {
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
				log.Info("No event found today")
			} else {
				for _, recipient := range cfg.Emails {
					m := &Email{
						FromName:  "Sentinel",
						FromEmail: cfg.Sender,
						Password:  cfg.Password,
						ToEmail:   recipient,
						Subject:   "Sentinel Report",
						Message:   "AAC Event: %s - %s\n\nConsider alternate routes. Recommended to approach via Harry Hines Blvd.",
					}
					if err := m.Send(); err != nil {
						log.Error(err.Error())
					}
				}
			}
		default:
			log.Info("No new events in the channel")
		}
	}
}

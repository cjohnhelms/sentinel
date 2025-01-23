package notify

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"sync"
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
	log.Info("Successfully sent to "+em.ToEmail, "SERIVCE", "NOTIFY")
	return nil
}

func Notify(ctx context.Context, wg *sync.WaitGroup, ch <-chan scraper.Event, cfg *config.Config) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Info("Killing notify routine")
			return
		default:
			// Get the current time
			now := time.Now()

			// Calculate the next 2 PM
			nextNotify := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, now.Location())

			if now.After(nextNotify) {
				// If it’s already past 2 PM, schedule it for the next day
				nextNotify = nextNotify.Add(24 * time.Hour)
			}

			// Calculate the duration until the next 2 PM
			duration := nextNotify.Sub(now)

			// Sleep until the next 2 PM
			time.Sleep(duration)

			event := <-ch
			today := time.Now().Format("2006-01-02")
			if event.Date == today {
				log.Info(fmt.Sprintf("Sending emails to: %v", cfg.Emails), "SERVICE", "NOTIFY")
				for _, recipient := range cfg.Emails {
					m := &Email{
						FromName:  "Sentinel",
						FromEmail: cfg.Sender,
						Password:  cfg.Password,
						ToEmail:   recipient,
						Subject:   "Sentinel Report",
						Message:   fmt.Sprintf("AAC Event: %s - %s\n\nConsider alternate routes. Recommended to approach via Harry Hines Blvd.", event.Title, event.Start),
					}
					if err := m.Send(); err != nil {
						log.Error(err.Error(), "SERVICE", "NOTIFY")
					}
				}
			} else {
				log.Info("No event today", "SERIVCE", "NOTIFY")
			}
		}
	}
}

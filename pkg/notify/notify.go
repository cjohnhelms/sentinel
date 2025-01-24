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

	notifyWg := new(sync.WaitGroup)

	for {
		select {
		case <-ctx.Done():
			log.Info("Killing notify routine")
			notifyWg.Wait()
			return
		case event := <-ch:
			today := time.Now().Format("2006-01-02")
			if event.Date == today {
				log.Info(fmt.Sprintf("Sending emails to: %v", cfg.Emails), "SERVICE", "NOTIFY")

				notifyWg.Add(1)
				go func(wg *sync.WaitGroup, event scraper.Event, timer int) {
					for {
						select {
						case <-ctx.Done():
							log.Info("Killing email routine")
							wg.Done()
							return
						default:
							if time.Now().Hour() == timer {
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
								log.Info("Emails sent, killing routine")
								wg.Done()
								return
							} else {
								log.Debug("Its not 3pm yet")
							}
						}
					}
				}(notifyWg, event, 15)

			} else {
				log.Info("No event today", "SERIVCE", "NOTIFY")
			}
		}
	}
}

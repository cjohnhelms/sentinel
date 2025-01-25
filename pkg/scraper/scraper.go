package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"

	"github.com/cjohnhelms/sentinel/pkg/config"
	log "github.com/cjohnhelms/sentinel/pkg/logging"
)

type Event struct {
	Date  string
	Start string
	Title string
}

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

func parseDt(dt string) (string, string, error) {
	cleaned := strings.Split(strings.Join(strings.Fields(dt), " "), " - ")
	dateStr := cleaned[0]
	timeStr := cleaned[1]
	date, err := time.Parse("Jan 2, 2006", dateStr)
	if err != nil {
		return "", "", errors.New("Could not parse date: " + dateStr)
	}
	isoDate := date.Format("2006-01-02")
	return isoDate, timeStr, nil
}

func Scrape(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup) Event {
	today := time.Now().Format("2006-01-02")

	var event = Event{
		Title: "No event today",
		Start: "",
		Date:  "",
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.americanairlinescenter.com"))

	c.OnRequest(func(r *colly.Request) {
		log.Info(fmt.Sprintf("Visiting: %s", r.URL.String()))
	})
	c.OnResponse(func(r *colly.Response) {
		log.Info(fmt.Sprintf("Visited: %s", r.Request.URL.String()))
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Info(fmt.Sprintf("Failed to scrape page: %s", err))
	})
	c.OnHTML("div.info.clearfix", func(e *colly.HTMLElement) {
		dt := e.ChildText("div.date")
		title := e.ChildText("h3 a")

		isoDate, timeStr, err := parseDt(dt)
		if err != nil {
			log.Error(err.Error())
		}

		if isoDate == today {
			event = Event{Date: isoDate, Start: timeStr, Title: title}
			log.Info(fmt.Sprintf("Found event today, queueing email: %+v", event))

			wg.Add(1)
			go func(wg *sync.WaitGroup, cfg *config.Config, event Event, timer int) {
				log.Info("Email queued")
				for {
					select {
					case <-ctx.Done():
						log.Info("Killing email routine")
						wg.Done()
						return
					default:
						if time.Now().Hour() == timer {
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
							log.Info("Emails sent, killing routine")
							wg.Done()
							return
						}
					}
				}
			}(wg, cfg, event, 15)
		} else {
			log.Info("No event today")
		}
	})
	err := c.Visit("https://www.americanairlinescenter.com/events")
	if err != nil {
		log.Error(fmt.Sprintf("Failed: %s\n", err))
	}

	return event
}

func FetchEvents(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup, ch chan<- Event) {
	defer wg.Done()

	// send inital scrape
	event := Scrape(ctx, cfg, wg)
	ch <- event

	for {
		select {
		case <-ctx.Done():
			log.Info("Killing scraper routine")
			return
		default:
			// scrape events
			if time.Now().Hour() == 2 && time.Now().Minute() == 0 && time.Now().Second() == 0 {
				event := Scrape(ctx, cfg, wg)
				log.Debug("Sending event in channel")
				ch <- event
			}
			time.Sleep(1 * time.Minute)
		}
	}
}

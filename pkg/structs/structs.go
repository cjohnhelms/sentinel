package structs

import (
	"errors"
	"fmt"
	"net/smtp"

	"github.com/cjohnhelms/sentinel/pkg/config"
	log "github.com/cjohnhelms/sentinel/pkg/logging"
)

type Event struct {
	Date  string
	Start string
	Title string
}

type Email struct {
	FromName string
	ToEmail  string
	Subject  string
	Message  string
}

func (em *Email) Send(cfg *config.Config) error {
	auth := smtp.PlainAuth("", cfg.Sender, cfg.Password, "smtp.gmail.com")
	msg := fmt.Sprintf("From: %s %s\nTo: %s\nSubject: %s\n\n%s", em.FromName, cfg.Sender, em.ToEmail, em.Subject, em.Message)

	err := smtp.SendMail("smtp.gmail.com:587",
		auth,
		cfg.Sender,
		[]string{em.ToEmail},
		[]byte(msg),
	)
	if err != nil {
		return errors.New("smtp error: " + err.Error())
	}
	log.Info("Successfully sent to "+em.ToEmail, "SERIVCE", "NOTIFY")
	return nil
}

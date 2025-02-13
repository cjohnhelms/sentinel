package structs

import (
	"errors"
	"fmt"
	"net/smtp"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/config"
)

type Event struct {
	Title string
	When  time.Time
}

type Email struct {
	FromName string
	ToEmail  string
	Subject  string
	Message  string
}

func (em *Email) Send(cfg *config.Config) error {
	logger := cfg.Logger

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
	logger.Info("Successfully sent to "+em.ToEmail, "SERIVCE", "NOTIFY")
	return nil
}

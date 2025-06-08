package config

import (
	"errors"
	"fmt"
	"github.com/cjohnhelms/sentinel/pkg/scraper"
	"log/slog"
	"net/smtp"
	"strings"
	"text/template"
)

type Email struct {
	Recipient string
	Sender    string
	Subject   string
	Body      string

	Server   string
	Password string
}

func (e *Email) Send(events []scraper.Event) error {
	err := e.generateBody(events)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", e.Sender, e.Password, e.Server)
	fullMsg := fmt.Sprintf("From: Sentinel %s\nTo: %s\nSubject: %s\n\n%s", e.Sender, e.Recipient, e.Subject, e.Body)

	err = smtp.SendMail(
		fmt.Sprintf("%s:587", e.Server),
		auth,
		e.Sender,
		[]string{e.Recipient},
		[]byte(fullMsg),
	)
	if err != nil {
		return errors.New("smtp error: " + err.Error())
	}
	slog.Info(fmt.Sprintf("email sent successfully to %s", e.Recipient))
	return nil
}

func (e *Email) generateBody(events []scraper.Event) error {
	var message strings.Builder
	templ := template.New("emailTemplate")
	templ, err := templ.Parse("AAC Events:\n\n {{range .}}{{.Title}} - {{.When}}\n{{ end}}\n\nConsider alternate routes.")
	if err != nil {
		return fmt.Errorf("error parsing email template %s", err)
	}
	err = templ.Execute(&message, events)
	if err != nil {
		return fmt.Errorf("error executing email template %s", err)
	}
	e.Body = message.String()
	return nil
}

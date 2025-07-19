package email

import (
	"fmt"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"net/smtp"
	"strings"
	"text/template"
)

type Emails struct {
	Recipient []string
	Sender    string
	Server    string
	Password  string
}

func (e *Emails) Send(events []event.Event) error {
	body, err := e.generateBody(events)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", e.Sender, e.Password, e.Server)
	fullMsg := fmt.Sprintf("From: Sentinel %s\nSubject: %s\n\n%s", e.Sender, "AAC Sentinel Notification", body)

	if err := smtp.SendMail(
		fmt.Sprintf("%s:587", e.Server),
		auth,
		e.Sender,
		e.Recipient,
		[]byte(fullMsg),
	); err != nil {
		return err
	}
	return nil
}

func (e *Emails) generateBody(events []event.Event) (string, error) {
	var message strings.Builder
	templ := template.New("emailTemplate")
	templ, err := templ.Parse("AAC Events:\n\n {{range .}}{{.Title}} - {{.When}}\n{{ end}}\n\nConsider alternate routes.")
	if err != nil {
		return "", fmt.Errorf("error parsing email template %s", err)
	}
	err = templ.Execute(&message, events)
	if err != nil {
		return "", fmt.Errorf("error executing email template %s", err)
	}
	return message.String(), nil
}

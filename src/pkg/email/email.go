package email

import (
	"fmt"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/event"
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

// MailSender abstracts the SMTP transport layer.
type MailSender interface {
	SendMail(addr, from string, to []string, msg []byte) error
}

// SMTPSender is the production MailSender backed by net/smtp.
type SMTPSender struct {
	password string
}

func NewSMTPSender(password string) *SMTPSender {
	return &SMTPSender{password: password}
}

func (s *SMTPSender) SendMail(addr, from string, to []string, msg []byte) error {
	server := strings.Split(addr, ":")[0]
	auth := smtp.PlainAuth("", from, s.password, server)
	return smtp.SendMail(addr, auth, from, to, msg)
}

// Send renders the email body and sends it to all recipients.
func Send(s MailSender, cfg *config.Config, events []event.Event) error {
	body, err := generateBody(events)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:587", cfg.EmailServer)
	msg := []byte(fmt.Sprintf("From: Sentinel %s\nSubject: AAC Sentinel Notification\n\n%s",
		cfg.ServiceEmail, body))
	if err := s.SendMail(addr, cfg.ServiceEmail, cfg.RecipientEmails, msg); err != nil {
		return fmt.Errorf("%w: %v", ErrEmail, err)
	}
	return nil
}

// generateBody renders the email body for the given events.
func generateBody(events []event.Event) (string, error) {
	var message strings.Builder
	tmpl, err := template.New("emailTemplate").Parse(
		"AAC Events:\n\n {{range .}}{{.Title}} - {{.When}}\n{{ end}}\n\nConsider alternate routes.",
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrBodyGenerate, err)
	}
	if err := tmpl.Execute(&message, events); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBodyGenerate, err)
	}
	return message.String(), nil
}

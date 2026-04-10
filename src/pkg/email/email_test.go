package email

import (
	"errors"
	"strings"
	"testing"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/event"
)

type fakeSender struct {
	capturedAddr string
	capturedFrom string
	capturedTo   []string
	capturedMsg  []byte
	err          error
}

func (f *fakeSender) SendMail(addr, from string, to []string, msg []byte) error {
	f.capturedAddr = addr
	f.capturedFrom = from
	f.capturedTo = to
	f.capturedMsg = msg
	return f.err
}

func TestSend(t *testing.T) {
	cfg := &config.Config{
		ServiceEmail:        "sentinel@example.com",
		EmailServer:         "smtp.example.com",
		RecipientEmails:     []string{"user@example.com"},
		EmailServerPassword: "secret",
	}
	events := []event.Event{
		{Title: "Test Event", When: "2026-04-09 7:30PM"},
	}
	sender := &fakeSender{}

	if err := Send(sender, cfg, events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.capturedAddr != "smtp.example.com:587" {
		t.Errorf("addr = %q, want smtp.example.com:587", sender.capturedAddr)
	}
	if sender.capturedFrom != "sentinel@example.com" {
		t.Errorf("from = %q, want sentinel@example.com", sender.capturedFrom)
	}
	if len(sender.capturedTo) != 1 || sender.capturedTo[0] != "user@example.com" {
		t.Errorf("to = %v, want [user@example.com]", sender.capturedTo)
	}
	if !strings.Contains(string(sender.capturedMsg), "Test Event") {
		t.Error("message does not contain event title")
	}
}

func TestSend_senderError(t *testing.T) {
	cfg := &config.Config{
		ServiceEmail:    "sentinel@example.com",
		EmailServer:     "smtp.example.com",
		RecipientEmails: []string{"user@example.com"},
	}
	events := []event.Event{{Title: "Test", When: "2026-04-09 7:30PM"}}
	sender := &fakeSender{err: errors.New("connection refused")}

	err := Send(sender, cfg, events)
	if !errors.Is(err, ErrEmail) {
		t.Errorf("expected ErrEmail, got %v", err)
	}
}

func TestGenerateBody(t *testing.T) {
	events := []event.Event{
		{Title: "Dallas Stars vs. Avalanche", When: "2026-04-09 7:30PM"},
		{Title: "Concert", When: "2026-04-09 9PM"},
	}
	body, err := generateBody(events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "Dallas Stars vs. Avalanche") {
		t.Error("body missing first event title")
	}
	if !strings.Contains(body, "Concert") {
		t.Error("body missing second event title")
	}
	if !strings.Contains(body, "Consider alternate routes") {
		t.Error("body missing footer")
	}
}

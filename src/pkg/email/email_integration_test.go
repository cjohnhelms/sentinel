//go:build integration

package email

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/event"
)

func TestSend_integration(t *testing.T) {
	vars, err := loadEnv()
	if err != nil {
		t.Fatalf("missing env vars: %v", err)
	}
	cfg := &config.Config{
		RecipientEmails:     strings.Split(vars["email_recipients"], ","),
		ServiceEmail:        vars["service_email"],
		EmailServer:         vars["email_server"],
		EmailServerPassword: vars["email_server_password"],
	}
	events := []event.Event{
		{Title: "Test title 1", When: time.Now().Format("2006-01-02 3:04PM")},
		{Title: "Test title 2", When: time.Now().Format("2006-01-02 3:04PM")},
	}
	if err := Send(NewSMTPSender(cfg.EmailServerPassword), cfg, events); err != nil {
		t.Error(err)
	}
}

func loadEnv() (map[string]string, error) {
	keys := []string{"EMAIL_RECIPIENTS", "EMAIL_SERVER", "EMAIL_SERVER_PASSWORD", "SERVICE_EMAIL"}
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		if !ok {
			return nil, errors.New(k + " not set")
		}
		out[strings.ToLower(k)] = v
	}
	return out, nil
}

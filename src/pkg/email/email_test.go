package email

import (
	"errors"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"os"
	"strings"
	"testing"
	"time"
)

func initialize() (map[string]string, error) {
	rc, ok := os.LookupEnv("EMAIL_RECIPIENTS")
	if !ok {
		return nil, errors.New("EMAIL_RECIPIENTS environment variable not set")
	}
	es, ok := os.LookupEnv("EMAIL_SERVER")
	if !ok {
		return nil, errors.New("EMAIL_SERVER environment variable not set")
	}
	esp, ok := os.LookupEnv("EMAIL_SERVER_PASSWORD")
	if !ok {
		return nil, errors.New("EMAIL_SERVER_PASSWORD environment variable not set")
	}
	se, ok := os.LookupEnv("SERVICE_EMAIL")
	if !ok {
		return nil, errors.New("SERVICE_EMAIL environment variable not set")
	}
	return map[string]string{
		"email_recipients":      rc,
		"email_server":          es,
		"email_server_password": esp,
		"email_service":         se,
	}, nil
}

func TestEmail_Send(t *testing.T) {
	vars, err := initialize()
	if err != nil {
		t.Fatalf("failed to initialize: %s", err)
	}

	e := &Emails{
		Recipient: strings.Split(vars["email_recipients"], ","),
		Server:    vars["email_server"],
		Sender:    vars["email_server_password"],
		Password:  vars["email_service"],
	}
	events := []event.Event{
		{
			Title: "Test title 1",
			When:  time.Now().Format("2006-01-02 3:04PM"),
		},
		{
			Title: "Test title 2",
			When:  time.Now().Format("2006-01-02 3:04PM"),
		},
	}
	if err = e.Send(events); err != nil {
		t.Error(err)
	}
}

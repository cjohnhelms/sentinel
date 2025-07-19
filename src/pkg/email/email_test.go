package email

import (
	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"testing"
	"time"
)

func TestEmail_Send(t *testing.T) {
	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	e := &Emails{
		Recipient: cfg.RecipientEmails,
		Server:    cfg.EmailServer,
		Sender:    cfg.ServiceEmail,
		Password:  cfg.EmailServerPassword,
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

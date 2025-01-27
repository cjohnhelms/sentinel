package email

import (
	"testing"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/structs"
)

func TestSendEmail(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatal("Failed to generate config")
	}
	e := &structs.Email{
		FromName: "TEST Sentinel",
		ToEmail:  "cjohnhelms@gmail.com",
		Subject:  "TEST Sentinel Report",
		Message:  "TEST CONTENT",
	}
	if err := e.Send(cfg); err != nil {
		t.Fatal("Failed to send email")
	}
}

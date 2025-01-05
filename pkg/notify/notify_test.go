package notify

import (
	"fmt"
	"os"
	"testing"
)

func TestSendMail(t *testing.T) {
	m := &Email{
		FromName:  "Sentinel",
		FromEmail: os.Getenv("SENDER"),
		ToEmail:   "cjohnhelms@gmail.com",
		Password:  os.Getenv("PASSWORD"),
		Subject:   "Sentinel Report Test",
		Message:   "This is a test.",
	}
	fmt.Printf("%+v", *m)
	if err := m.Send(); err != nil {
		t.Fatalf("Failed to send email: %s", err)
	}
}

package scraper

import (
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/structs"
)

func TestDateVerification(t *testing.T) {
	event1 := structs.Event{
		When: time.Now(),
	}

	if !isToday(event1.When) {
		t.Fatalf("Failed to properly parse event on the current day")
	}
}

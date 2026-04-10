package email

import (
	"strings"
	"testing"

	"github.com/cjohnhelms/sentinel/pkg/event"
)

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

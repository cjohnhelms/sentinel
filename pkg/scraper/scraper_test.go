package scraper

import (
	"os"
	"testing"
	"time"
)

func TestDateVerification(t *testing.T) {
	raw := os.Getenv("EVENT")

	date, _, err := parseDt(raw)
	if err != nil {
		t.Fatal("Failed to parse date")
	}

	today := time.Now().Format("2006-01-02")

	if date != today {
		t.Fatal("Unable to verify today's date")
	}

}

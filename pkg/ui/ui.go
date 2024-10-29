package ui

import (
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"sentinel/pkg/scraper"
	"time"
)

func Update(ch <-chan []scraper.Event) {
	for {
		// Get the current time
		now := time.Now()

		// Calculate the next 2 PM
		next2PM := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
		if now.After(next2PM) {
			// If it’s already past 2 PM, schedule it for the next day
			next2PM = next2PM.Add(24 * time.Hour)
		}

		// Calculate the duration until the next 2 PM
		duration := next2PM.Sub(now)

		// Sleep until the next 2 PM
		time.Sleep(duration)

		today := time.Now().Format("2006-01-02")
		select {
		case data := <-ch:
			for _, event := range data {
				if event.Date == today {
					a, _ := pterm.DefaultArea.WithFullscreen().WithCenter().Start()
					s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString(event.Title)).Srender()
					t, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString(event.Start)).Srender()
					a.Update(s + "\n" + t)
				}
			}
		default:
			a, _ := pterm.DefaultArea.WithFullscreen().WithCenter().Start()
			s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("None")).Srender()
			a.Update(s + "\n")
		}
	}
}

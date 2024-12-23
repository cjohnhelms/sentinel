package display

import (
	"log"
	"sentinel/pkg/scraper"
	"time"

	lcd "github.com/mskrha/rpi-lcd"
)

func Write(event scraper.Event) {
	screen := lcd.New(lcd.LCD{Bus: "/dev/i2c-1", Address: 0x27, Rows: 2, Cols: 16, Backlight: true})

	if err := screen.Init(); err != nil {
		panic(err)
	}

	// write time first because this is static
	if err := screen.Print(2, 0, event.Start); err != nil {
		log.Println("screen update failure:", err)
	}

	if len(event.Title) <= 16 {
		screen.Print(1, 0, event.Title)
	} else {
		var i int
		var max = len(event.Title) - 15
		for {
			if err := screen.Print(1, 0, event.Title); err != nil {
				log.Println("screen update failure:", err)
			}
			time.Sleep(800 * time.Millisecond)
			i++
			if i == max {
				i = 0
				time.Sleep(3 * time.Second)
			}
		}
	}
}

func Update(ch <-chan scraper.Event) {
	for {
		// Get the current time
		now := time.Now()

		// Calculate the next 2 PM
		nextUpdate := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
		if now.After(nextUpdate) {
			// If it’s already past 2 PM, schedule it for the next day
			nextUpdate = nextUpdate.Add(24 * time.Hour)
		}

		// Calculate the duration until the next 2 PM
		duration := nextUpdate.Sub(now)

		// Sleep until the next 2 PM
		time.Sleep(duration)

		select {
		case data := <-ch:
			event := data
			Write(event)

		default:
			log.Println("no events in the channel, something might be wrong")
		}
	}
}

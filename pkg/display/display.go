package display

import (
	"fmt"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/scraper"

	log "github.com/cjohnhelms/sentinel/pkg/logging"
	lcd "github.com/mskrha/rpi-lcd"
)

func Write(event scraper.Event) {
	screen := lcd.New(lcd.LCD{Bus: "/dev/i2c-1", Address: 0x27, Rows: 2, Cols: 16, Backlight: true})
	log.Debug(fmt.Sprintf("Screen: %+v", screen))

	if err := screen.Init(); err != nil {
		log.Error("Failed to init screen, proceeding with SMS")
	}

	log.Debug(fmt.Sprintf("Writing screen: %s - %s", event.Title, event.Start))

	// write time first because this is static
	if event.Start != "" {
		if err := screen.Print(2, 0, event.Start); err != nil {
			log.Error(fmt.Sprintf("Screen update failure: %s", err))
		}
	}

	if len(event.Title) <= 16 {
		screen.Print(1, 0, event.Title)
	} else {
		var i int
		var max = len(event.Title) - 15
		for {
			if err := screen.Print(1, 0, event.Title[i:(i+16)]); err != nil {
				log.Error(fmt.Sprintf("Screen update failure: %s", err))
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
		time.Sleep(60 * time.Second)

		select {
		case data := <-ch:
			event := data
			Write(event)

		default:
			log.Debug("No new events found in the channel")
		}
	}
}

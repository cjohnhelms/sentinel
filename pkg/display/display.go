package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/scraper"

	log "github.com/cjohnhelms/sentinel/pkg/logging"
	lcd "github.com/mskrha/rpi-lcd"
)

func writeTitle(event scraper.Event, screen *lcd.LCD, quit <-chan bool) {
	var i int
	var max = len(event.Title) - 15
	for {
		select {
		case <-quit:
			log.Info("New event found and quit recieved, killing goroutine")
			return
		default:
			if err := screen.Print(1, 0, event.Title[i:(i+16)]); err != nil {
				log.Error(fmt.Sprintf("Screen update failure: %s", err), "SERIVCE", "DISPLAY")
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

func Update(ch <-chan scraper.Event, quit <-chan bool) {
	for {
		time.Sleep(60 * time.Second)

		select {
		case event := <-ch:
			screen := lcd.New(lcd.LCD{Bus: "/dev/i2c-1", Address: 0x27, Rows: 2, Cols: 16, Backlight: true})
			log.Debug(fmt.Sprintf("Screen: %+v", screen), "SERVICE", "DISPLAY")

			if err := screen.Init(); err != nil {
				log.Error("Failed to init screen, proceeding with SMS", "SERVICE", "DISPLAY")
			}

			log.Debug(fmt.Sprintf("Writing screen: %s - %s", event.Title, event.Start), "SERVICE", "DISPLAY")

			// write time first because this is static
			if event.Start != "" {
				if err := screen.Print(2, 0, event.Start); err != nil {
					log.Error(fmt.Sprintf("Screen update failure: %s", err), "SERVICE", "DISPLAY")
				}
			} else {
				if err := screen.Print(2, 0, strings.Repeat(" ", 16)); err != nil {
					log.Error(fmt.Sprintf("Screen update failure: %s", err), "SERVICE", "DISPLAY")
				}
			}

			if len(event.Title) <= 16 {
				r := (16 - len(event.Title))
				t := event.Title + strings.Repeat(" ", r)
				screen.Print(1, 0, t)
			} else {
				go writeTitle(event, screen, quit)
			}
		default:
			log.Debug("No new data in the channel", "SERVICE", "DISPLAY")
		}
	}
}

package display

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/scraper"

	log "github.com/cjohnhelms/sentinel/pkg/logging"
	lcd "github.com/mskrha/rpi-lcd"
)

func writeScreen(event scraper.Event, quit <-chan bool) {
	log.Debug("Starting screen write routine")
	for {
		select {
		case <-quit:
			log.Info("Quit recieved, killing routine")
			return
		default:
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
				var max = len(event.Title) - 15
				for i := 0; i < max; i++ {
					if err := screen.Print(1, 0, event.Title[i:(i+16)]); err != nil {
						log.Error(fmt.Sprintf("Screen update failure: %s", err), "SERIVCE", "DISPLAY")
					}
					time.Sleep(800 * time.Millisecond)
				}
				time.Sleep(3 * time.Second)
			}
		}
	}
}

func Update(ctx context.Context, wg *sync.WaitGroup, ch <-chan scraper.Event, quit chan bool) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Info("Killing display routine and children")
			quit <- true
			return
		default:
			event := <-ch
			go writeScreen(event, quit)
		}
	}
}

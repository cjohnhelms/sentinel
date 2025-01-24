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

func writeScreen(wg *sync.WaitGroup, event scraper.Event, quit <-chan bool) {
	defer wg.Done()

	// screen init
	log.Debug("Starting new screen write routine")
	screen := lcd.New(lcd.LCD{Bus: "/dev/i2c-1", Address: 0x27, Rows: 2, Cols: 16, Backlight: true})
	log.Debug(fmt.Sprintf("Screen: %+v", screen), "SERVICE", "DISPLAY")

	if err := screen.Init(); err != nil {
		log.Error("Failed to init screen, proceeding with SMS", "SERVICE", "DISPLAY")
	}

	// loop
	for {
		select {
		case <-quit:
			log.Info("Quit recieved, killing routine")
			return
		default:
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
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func Update(ctx context.Context, wg *sync.WaitGroup, ch <-chan scraper.Event) {
	defer wg.Done()

	time.Sleep(10 * time.Second)

	quit := make(chan bool, 1)
	writeWg := new(sync.WaitGroup)

	// wait for first event and launch initial routine
	event := <-ch
	log.Debug("Launching inital screen write routine")
	go writeScreen(writeWg, event, quit)
	writeWg.Add(1) // add wait

	// start loop
	for {
		select {
		case <-ctx.Done():
			log.Info("Killing display routine and children")
			quit <- true
			writeWg.Wait()
			log.Debug("Screen write waitgroup done, display routine exiting")
			return
		case event := <-ch:
			// when new event enters the channel, kill old screen write routine
			log.Debug("New event recieved, sending quit signal to kill last routine")
			quit <- true
			writeWg.Wait() // wait for done
			log.Debug("Screen write waitgroup done, launching new routine")
			go writeScreen(writeWg, event, quit) // launch new routine
			writeWg.Add(1)                       // add new wait
		}
	}
}

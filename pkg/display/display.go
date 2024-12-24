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
		log.Println("Failed to init screen, proceeding with SMS")
	}

	// write time first because this is static
	if err := screen.Print(2, 0, event.Start); err != nil {
		log.Println("Screen update failure:", err)
	}

	if len(event.Title) <= 16 {
		screen.Print(1, 0, event.Title)
	} else {
		var i int
		var max = len(event.Title) - 15
		for {
			if err := screen.Print(1, 0, event.Title[i:(i+16)]); err != nil {
				log.Println("Screen update failure:", err)
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
		time.Sleep(30 * time.Second)

		select {
		case data := <-ch:
			event := data
			Write(event)

		default:
			log.Println("No new events found in the channel")
		}
	}
}

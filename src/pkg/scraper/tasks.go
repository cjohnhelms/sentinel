package scraper

import (
	"context"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"github.com/cjohnhelms/sentinel/pkg/scheduler"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

func NewScrapeTask() scheduler.Task {
	return scheduler.Task{
		ID:         uuid.New(),
		RunOnStart: true,
		Scheduled:  false,
		Cron:       "",
		TaskFunc: func(ctx context.Context, eventChan chan []event.Event) {
			slog.Info("starting scrape task")

			select {
			case <-ctx.Done():
				slog.Error("cancel received, ending routine")
				return
			default:
				zone, err := time.LoadLocation("America/Chicago")
				if err != nil {
					slog.Warn("failed to load timezone, using local")
					zone = time.Local
				}
				todaysEvents, err := scrape(time.Now().In(zone))
				if err != nil {
					slog.Error(err.Error())
					return
				}
				if len(todaysEvents) == 0 {
					slog.Info("no events found")
					close(eventChan)
					return
				} else {
					eventChan <- todaysEvents
					close(eventChan)
					return
				}
			}
		},
	}
}

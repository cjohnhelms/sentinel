package email

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/structs"
)

func ScheduleEmail(ctx context.Context, cfg *config.Config, wg *sync.WaitGroup, event structs.Event) {
	defer wg.Done()

	logger := cfg.Logger

	logger.Info("Email queued")
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), cfg.EmailHour, cfg.EmailMin, 0, 0, now.Location())

		if now.After(next) {
			// if past the time, schedule it for the next day
			next = next.Add(24 * time.Hour)
		}

		duration := time.Until(next)
		logger.Debug(fmt.Sprintf("Sending email in %v", duration))
		timer := time.NewTimer(duration)

		select {
		case <-ctx.Done():
			logger.Info("Email routine recieved signal, killing")
			return
		case <-timer.C:
			for _, recipient := range cfg.Emails {
				m := &structs.Email{
					FromName: "Sentinel",
					ToEmail:  recipient,
					Subject:  "Sentinel Report",
					Message:  fmt.Sprintf("AAC Event: %s - %s\n\nConsider alternate routes.", event.Title, event.Start),
				}
				if err := m.Send(cfg); err != nil {
					logger.Error(err.Error(), "SERVICE", "NOTIFY")
				} else {
					logger.Info(fmt.Sprintf("Successful email: %s", recipient))
				}
			}
			logger.Info("Email process complete, killing routine")
			return
		}
	}
}

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
	loc, _ := time.LoadLocation("America/Chicago")
	event.When = event.When.In(loc)
	emailTime := event.When.Add(-(4 * time.Hour))
	logger.Info(fmt.Sprintf("Event time: %s", event.When))
	logger.Info(fmt.Sprintf("Email time: %s", emailTime))

	if emailTime.Before(time.Now()) {
		logger.Error("Email scheduled after the desired email time, sending ASAP")
		for _, recipient := range cfg.Emails {
			m := &structs.Email{
				FromName: "Sentinel",
				ToEmail:  recipient,
				Subject:  "Sentinel Report",
				Message:  fmt.Sprintf("AAC Event: %s - %s\n\nConsider alternate routes.", event.Title, event.When.Format("3:04 PM")),
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

	wait := time.Until(emailTime)
	logger.Debug(fmt.Sprintf("Sending email in %v", wait))
	timer := time.NewTimer(wait)

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
				Message:  fmt.Sprintf("AAC Event: %s - %s\n\nConsider alternate routes.", event.Title, event.When.Format("3:04 PM")),
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

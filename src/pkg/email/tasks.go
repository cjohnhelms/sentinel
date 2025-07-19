package email

import (
	"context"
	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"github.com/cjohnhelms/sentinel/pkg/scheduler"
	"github.com/google/uuid"
	"log/slog"
)

func NewEmailTask() scheduler.Task {
	return scheduler.Task{
		ID:         uuid.New(),
		RunOnStart: true,
		Scheduled:  false,
		Cron:       "",
		TaskFunc: func(ctx context.Context, eventChan chan []event.Event) {
			slog.Info("starting email task")
			cfg, err := config.NewConfig()
			if err != nil {
				slog.Error("error reading config")
			}

			select {
			case value, ok := <-eventChan:
				if !ok {
					slog.Debug("task channel closed, ending routine")
					return
				}

				// events found
				e := &Emails{
					Sender:    cfg.ServiceEmail,
					Server:    cfg.EmailServer,
					Password:  cfg.EmailServerPassword,
					Recipient: cfg.RecipientEmails,
				}
				err := e.Send(value)
				if err != nil {
					slog.Error(err.Error())
				}
				return
			case <-ctx.Done():
				slog.Error("cancel received, ending routine")
				return
			}
		},
	}
}

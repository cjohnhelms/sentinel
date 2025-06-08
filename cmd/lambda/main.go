package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/notifications"
	"github.com/cjohnhelms/sentinel/pkg/scraper"
	"log/slog"
	"os"
)

func main() {
	lambda.Start(handler)
}

func handler() {
	cfg, err := config.New("./config.toml")
	if err != nil {
		panic(err)
	}

	// logging setup
	config.InitLogger(cfg.Log)

	// gather events
	events, err := scraper.ScrapeEvents()
	if err != nil { // exit 0 if no events today
		// exit 1 if error occurs
		slog.Error(err.Error())
		os.Exit(1)
	}
	if len(events) == 0 { // exit 0 if no events today
		slog.Info("no events found today")
		os.Exit(0)
	}
	slog.Debug(fmt.Sprintf("Events today: %+v", events))

	// send notifications
	err = notifications.Send(cfg, events)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("notifications sent successfully")
}

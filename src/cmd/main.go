package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/email"
	"github.com/cjohnhelms/sentinel/pkg/scraper"
)

func main() {
	lambda.Start(handler)
}

func handler() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("%w: %v", config.ErrConfig, err)
	}
	config.InitLogger(cfg.Log)

	zone, err := time.LoadLocation("America/Chicago")
	if err != nil {
		slog.Warn("failed to load timezone, using UTC")
		zone = time.UTC
	}

	events, err := scraper.Scrape(scraper.NewHTTPFetcher(), time.Now().In(zone))
	if err != nil {
		return err
	}
	if len(events) == 0 {
		slog.Info("no events today")
		return nil
	}

	return email.Send(email.NewSMTPSender(cfg.EmailServerPassword), cfg, events)
}

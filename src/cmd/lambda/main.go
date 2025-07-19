package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/email"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"github.com/cjohnhelms/sentinel/pkg/scheduler"
	"github.com/cjohnhelms/sentinel/pkg/scraper"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	lambda.Start(handler)
}

func handler() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}
	// logging setup
	config.InitLogger(cfg.Log)

	scrapeTask := scraper.NewScrapeTask()
	emailTask := email.NewEmailTask()

	eventChan := make(chan []event.Event)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func(chan os.Signal) {
		<-sigs
		slog.Info("signal received, shutting down")
		cancel()
		return
	}(sigs)

	taskScheduler := scheduler.NewScheduler(ctx, eventChan, scrapeTask, emailTask)
	taskScheduler.Run()

	slog.Info("server exiting")
	return nil
}

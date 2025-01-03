package main

import (
	"os"
	"sentinel/pkg/config"
	"sentinel/pkg/display"
	"sentinel/pkg/scraper"

	"golang.org/x/exp/slog"
)

type loggeyKey struct{}

func main() {
	cfg := config.New()
	var level slog.Level

	switch cfg.LogLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo

	}
	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}
	textHandler := slog.NewTextHandler(os.Stdout, handlerOpts)
	logger := slog.New(textHandler)
	slog.SetDefault(logger)

	slog.Info("Service starting")

	ch := make(chan scraper.Event, 1)
	go scraper.FetchEvents(ch)
	//go notify.Notify(ch, recipients)
	go display.Update(ch)

	select {}
}

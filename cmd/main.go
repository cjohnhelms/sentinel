package main

import (
	"fmt"
	"os"
	"sentinel/pkg/config"
	"sentinel/pkg/display"
	"sentinel/pkg/notify"
	"sentinel/pkg/scraper"

	"golang.org/x/exp/slog"
)

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
	slog.Debug(fmt.Sprintf("Config: %+v", cfg))

	ch := make(chan scraper.Event, 1)
	go scraper.FetchEvents(ch)
	go notify.Notify(ch, cfg)
	go display.Update(ch)

	select {}
}

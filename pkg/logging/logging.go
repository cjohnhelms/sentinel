package logging

import (
	"log/slog"
	"os"

	"github.com/cjohnhelms/sentinel/pkg/config"
)

var logger *slog.Logger

func init() {
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
		Level:     level,
		AddSource: true,
	}
	textHandler := slog.NewTextHandler(os.Stdout, handlerOpts)
	logger = slog.New(textHandler)
}

func Info(msg string) {
	logger.Info(msg)
}

func Debug(msg string) {
	logger.Debug(msg)
}

func Error(msg string) {
	logger.Error(msg)
}

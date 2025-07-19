package config

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Log                 string
	RecipientEmails     []string
	ServiceEmail        string
	EmailServer         string
	EmailServerPassword string
}

func NewConfig() (*Config, error) {
	log := os.Getenv("LOG_LEVEL")
	emails, ok := os.LookupEnv("EMAIL_RECIPIENTS")
	if !ok {
		return nil, errors.New("EMAIL_RECIPIENTS environment variable not set")
	}
	server, ok := os.LookupEnv("EMAIL_SERVER")
	if !ok {
		return nil, errors.New("EMAIL_SERVER environment variable not set")
	}
	password, ok := os.LookupEnv("EMAIL_SERVER_PASSWORD")
	if !ok {
		return nil, errors.New("EMAIL_SERVER_PASSWORD environment variable not set")
	}
	semail, ok := os.LookupEnv("SERVICE_EMAIL")
	if !ok {
		return nil, errors.New("SERVICE_EMAIL environment variable not set")
	}
	return &Config{
		Log:                 log,
		RecipientEmails:     strings.Split(emails, ","),
		ServiceEmail:        semail,
		EmailServerPassword: password,
		EmailServer:         server,
	}, nil
}

func InitLogger(l string) {
	var level slog.Level
	switch l {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	default:
		level = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	slog.SetDefault(logger)
}

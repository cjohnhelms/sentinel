package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Sender     string
	Password   string
	Emails     []string
	Logger     *slog.Logger
	Version    string
	ScrapeHour int
	ScrapeMin  int
	EmailHour  int
	EmailMin   int
}

func newErr(env string) error {
	return fmt.Errorf("env variable %s is unset", env)
}

func New() (*Config, error) {
	var level slog.Level

	var sh int
	var sm int
	var eh int
	var em int

	sender, ok := os.LookupEnv("SENDER")
	if !ok {
		return nil, newErr("SENDER")
	}
	password, ok := os.LookupEnv("PASSWORD")
	if !ok {
		return nil, newErr("PASSWORD")
	}
	emails, ok := os.LookupEnv("NOTIFY")
	if !ok {
		return nil, newErr("NOTIFY")
	}
	loglvl, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return nil, newErr("LOG_LEVEL")
	}
	version, ok := os.LookupEnv("VERSION")
	if !ok {
		return nil, newErr("VERSION")
	}

	switch loglvl {
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
	jsonHandler := slog.NewJSONHandler(os.Stdout, handlerOpts)
	logger := slog.New(jsonHandler)

	// get times or set defaults
	sh, err := strconv.Atoi(os.Getenv("SCRAPE_HOUR"))
	if err != nil {
		sh = 2
	}
	sm, err = strconv.Atoi(os.Getenv("SCRAPE_MIN"))
	if err != nil {
		sm = 0
	}
	eh, err = strconv.Atoi(os.Getenv("EMAIL_HOUR"))
	if err != nil {
		eh = 3
	}
	em, err = strconv.Atoi(os.Getenv("EMAIL_MIN"))
	if err != nil {
		em = 0
	}
	return &Config{
		Sender:     sender,
		Password:   password,
		Emails:     strings.Split(emails, ","),
		Logger:     logger,
		Version:    version,
		ScrapeHour: sh,
		ScrapeMin:  sm,
		EmailHour:  eh,
		EmailMin:   em,
	}, nil
}

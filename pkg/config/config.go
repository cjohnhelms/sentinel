package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Sender     string
	Password   string
	Emails     []string
	LogLevel   string
	Version    string
	ScrapeHour int
	ScrapeMin  int
	EmailHour  int
	EmailMin   int
}

func New() (*Config, error) {
	sh, err := strconv.Atoi(os.Getenv("SCRAPE_HOUR"))
	if err != nil {
		return nil, err
	}
	sm, err := strconv.Atoi(os.Getenv("SCRAPE_MIN"))
	if err != nil {
		return nil, err
	}
	eh, err := strconv.Atoi(os.Getenv("EMAIL_HOUR"))
	if err != nil {
		return nil, err
	}
	em, err := strconv.Atoi(os.Getenv("EMAIL_MIN"))
	if err != nil {
		return nil, err
	}
	return &Config{
		Sender:     os.Getenv("SENDER"),
		Password:   os.Getenv("PASSWORD"),
		Emails:     strings.Split(os.Getenv("NOTIFY"), ","),
		LogLevel:   os.Getenv("LOG_LEVEL"),
		Version:    os.Getenv("VERSION"),
		ScrapeHour: sh,
		ScrapeMin:  sm,
		EmailHour:  eh,
		EmailMin:   em,
	}, nil
}

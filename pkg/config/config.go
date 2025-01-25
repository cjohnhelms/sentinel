package config

import (
	"os"
	"strconv"
)

type Config struct {
	Sender     string
	Password   string
	Emails     [2]string
	LogLevel   string
	Version    string
	ScrapeTime int
	EmailTime  int
}

func New() (*Config, error) {
	s, err := strconv.Atoi(os.Getenv("SCRAPE_TIME"))
	if err != nil {
		return nil, err
	}
	e, err := strconv.Atoi(os.Getenv("EMAIL_TIME"))
	if err != nil {
		return nil, err
	}
	return &Config{
		Sender:   os.Getenv("SENDER"),
		Password: os.Getenv("PASSWORD"),
		Emails: [2]string{
			os.Getenv("CHRIS_EMAIL"), os.Getenv("KYLE_EMAIL"),
		},
		LogLevel:   os.Getenv("LOG_LEVEL"),
		Version:    os.Getenv("VERSION"),
		ScrapeTime: s,
		EmailTime:  e,
	}, nil
}

package config

import (
	"fmt"
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

func newErr(env string) error {
	return fmt.Errorf("env variable %s is unset", env)
}

func New() (*Config, error) {
	var sh int
	var sm int
	var eh int
	var em int

	sender, ok := os.LookupEnv("SENDER")
	if !ok {
		return nil, newErr("SENDER")
	}
	password, ok := os.LookupEnv("SENDER")
	if !ok {
		return nil, newErr("SENDER")
	}
	emails, ok := os.LookupEnv("SENDER")
	if !ok {
		return nil, newErr("SENDER")
	}
	loglvl, ok := os.LookupEnv("SENDER")
	if !ok {
		return nil, newErr("SENDER")
	}
	version, ok := os.LookupEnv("SENDER")
	if !ok {
		return nil, newErr("SENDER")
	}

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
		LogLevel:   loglvl,
		Version:    version,
		ScrapeHour: sh,
		ScrapeMin:  sm,
		EmailHour:  eh,
		EmailMin:   em,
	}, nil
}

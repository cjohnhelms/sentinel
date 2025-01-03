package config

import "os"

type Config struct {
	Emails   [2]string
	LogLevel string
	Version  string
}

func New() *Config {
	return &Config{
		Emails: [2]string{
			os.Getenv("CHRIS_EMAIL"), os.Getenv("KYLE_EMAIL"),
		},
		LogLevel: os.Getenv("LOG_LEVEL"),
		Version:  os.Getenv("VERSION"),
	}
}

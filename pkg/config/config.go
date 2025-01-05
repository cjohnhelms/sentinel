package config

import "os"

type Config struct {
	Sender   string
	Password string
	Emails   [2]string
	LogLevel string
	Version  string
}

func New() *Config {
	return &Config{
		Sender:   os.Getenv("SENDER"),
		Password: os.Getenv("PASSWORD"),
		Emails: [2]string{
			os.Getenv("CHRIS_EMAIL"), os.Getenv("KYLE_EMAIL"),
		},
		LogLevel: os.Getenv("LOG_LEVEL"),
		Version:  os.Getenv("VERSION"),
	}
}

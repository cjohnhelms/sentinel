package config

import (
	"github.com/joho/godotenv"
)

type Config struct {
	Sender     string
	Password   string
	Emails     [2]string
	LogLevel   string
	Version    string
	ScrapeTime string
	EmailTime  string
}

func New() *Config {
	envFile, _ := godotenv.Read(".env")
	return &Config{
		Sender:   envFile["SENDER"],
		Password: envFile["PASSWORD"],
		Emails: [2]string{
			envFile["CHRIS_EMAIL"], envFile["KYLE_EMAIL"],
		},
		LogLevel:   envFile["LOG_LEVEL"],
		Version:    envFile["VERSION"],
		ScrapeTime: envFile["SCRAPE_TIME"],
		EmailTime:  envFile["EMAIL_TIME"],
	}
}

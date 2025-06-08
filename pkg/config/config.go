package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Log           string
	Version       string
	Notifications struct {
		Email []*Email
		SMS   []*SMS
	}
}

func New(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var config Config
	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}

	fmt.Println(config)
	return &config, nil
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

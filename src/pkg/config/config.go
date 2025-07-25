package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Config struct {
	Log                 string
	RecipientEmails     []string
	ServiceEmail        string
	EmailServer         string
	EmailServerPassword string
}

type secrets struct {
	EmailRecipients     string `json:"email_recipients"`
	ServiceEmail        string `json:"service_email"`
	EmailServer         string `json:"email_server"`
	EmailServerPassword string `json:"email_server_password"`
}

func NewConfig() (*Config, error) {
	secretName := "sentinel-prod-secrets-use1"
	region := "us-east-1"

	config, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	scm := secretsmanager.NewFromConfig(config)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := scm.GetSecretValue(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var secretString string = *result.SecretString
	var secretConfig secrets
	if err = json.Unmarshal([]byte(secretString), &secretConfig); err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("secrets: %+v", secretConfig))

	return &Config{
		Log:                 os.Getenv("LOG_LEVEL"),
		RecipientEmails:     strings.Split(secretConfig.EmailRecipients, ","),
		ServiceEmail:        secretConfig.ServiceEmail,
		EmailServerPassword: secretConfig.EmailServerPassword,
		EmailServer:         secretConfig.EmailServer,
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

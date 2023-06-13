package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type CIServerConfig struct {
	Host      string
	Port      string
	SecretKey string

	MongoURI string

	RabbitMQURI string

	GitHubClientID     string
	GitHubClientSecret string

	GitLabClientID     string
	GitLabClientSecret string
}

func NewCIServerConfigFromEnv() (CIServerConfig, error) {
	config := CIServerConfig{
		Host:               stringEnv("HOST", ""),
		Port:               stringEnv("PORT", "8080"),
		SecretKey:          stringEnv("SECRET_KEY", ""),
		MongoURI:           stringEnv("MONGO_URI", "mongodb://localhost:27017"),
		RabbitMQURI:        stringEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672"),
		GitHubClientID:     stringEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: stringEnv("GITHUB_CLIENT_SECRET", ""),
		GitLabClientID:     stringEnv("GITLAB_CLIENT_ID", ""),
		GitLabClientSecret: stringEnv("GITLAB_CLIENT_SECRET", ""),
	}
	err := config.validate()
	return config, err
}

func (c CIServerConfig) validate() error {
	if c.Host == "" {
		return errors.New("HOST is required")
	}
	if c.SecretKey == "" {
		return errors.New("SECRET_KEY is required")
	}
	if c.GitHubClientID != "" && c.GitHubClientSecret == "" {
		return errors.New("GITHUB_CLIENT_SECRET is required when GITHUB_CLIENT_ID is set")
	}
	if c.GitLabClientID != "" && c.GitLabClientSecret == "" {
		return errors.New("GITLAB_CLIENT_SECRET is required when GITLAB_CLIENT_ID is set")
	}
	if c.GitHubClientID == "" && c.GitLabClientID == "" {
		return errors.New("either GITHUB_CLIENT_ID or GITLAB_CLIENT_ID must be set")
	}

	return nil
}

type WorkerConfig struct {
	MaxWorkers int
	ReposPath  string

	CIServerHost string
	CIServerPort string

	RabbitMQURI string
}

func NewWorkerConfigFromEnv() (WorkerConfig, error) {
	config := WorkerConfig{
		MaxWorkers:   intEnv("MAX_WORKERS", 4),
		ReposPath:    stringEnv("REPOS_PATH", "./repos"),
		CIServerHost: stringEnv("CISERVER_HOST", "localhost"),
		CIServerPort: stringEnv("CISERVER_PORT", "8080"),
		RabbitMQURI:  stringEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
	}

	return config, nil
}

func stringEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func intEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}

		return v
	}

	return fallback
}

func boolEnv(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if strings.ToLower(value) == "true" || value == "1" {
			return true
		}

		return false
	}

	return fallback
}

package config

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var ServerConf ServerConfig
var WorkerConf WorkerConfig

type ServerConfig struct {
	Host      string
	Port      string
	GRPCPort  string
	SecretKey string

	DB DatabaseConfig
	MQ MessageQueueConfig

	GitHub ServiceConfig
	GitLab ServiceConfig
}

type DatabaseConfig struct {
	URI string
}

type MessageQueueConfig struct {
	URI string
}

type ServiceConfig struct {
	ClientID     string
	ClientSecret string
}

type WorkerConfig struct {
	MaxWorkers int
	ReposPath  string

	MQ MessageQueueConfig

	CIServerHost     string
	CIServerGRPCPort string
}

func LoadServerConfigFromEnv() error {
	config := ServerConfig{
		Host:      stringEnv("HOST", "localhost"),
		Port:      stringEnv("PORT", "8000"),
		GRPCPort:  stringEnv("GRPC_PORT", "9000"),
		SecretKey: stringEnv("SECRET_KEY", ""),
		DB: DatabaseConfig{
			URI: stringEnv("DB_URI", "postgres://localhost/shark-ci"),
		},
		MQ: MessageQueueConfig{
			URI: stringEnv("MQ_URI", "amqp://guest:guest@localhost"),
		},
		GitHub: ServiceConfig{
			ClientID:     stringEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: stringEnv("GITHUB_CLIENT_SECRET", ""),
		},
		GitLab: ServiceConfig{
			ClientID:     stringEnv("GITLAB_CLIENT_ID", ""),
			ClientSecret: stringEnv("GITLAB_CLIENT_SECRET", ""),
		},
	}
	err := config.validate()
	if err != nil {
		return err
	}

	ServerConf = config
	return nil
}

func (c ServerConfig) validate() error {
	if c.SecretKey == "" {
		return errors.New("config: SECRET_KEY is required")
	}
	if c.GitHub.ClientID == "" && c.GitLab.ClientID == "" {
		return errors.New("config: either GITHUB_CLIENT_ID or GITLAB_CLIENT_ID must be set")
	}
	if c.GitHub.ClientID != "" && c.GitHub.ClientSecret == "" {
		return errors.New("config: GITHUB_CLIENT_SECRET is required when GITHUB_CLIENT_ID is set")
	}
	if c.GitLab.ClientID != "" && c.GitLab.ClientSecret == "" {
		return errors.New("config: GITLAB_CLIENT_SECRET is required when GITLAB_CLIENT_ID is set")
	}

	return nil
}

func LoadWorkerConfigFromEnv() error {
	config := WorkerConfig{
		MaxWorkers:       intEnv("MAX_WORKERS", runtime.GOMAXPROCS(0)),
		ReposPath:        stringEnv("REPOS_PATH", "./repos"),
		CIServerHost:     stringEnv("HOST", "localhost"),
		CIServerGRPCPort: stringEnv("GRPC_PORT", "9000"),
		MQ: MessageQueueConfig{
			URI: stringEnv("MQ_URI", "amqp://guest:guest@localhost"),
		},
	}

	WorkerConf = config
	return nil
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

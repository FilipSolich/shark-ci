package config

import (
	"errors"

	"github.com/shark-ci/shark-ci/shared/env"
)

var Conf Config

type CIServerConfig struct {
	Host      string
	Port      string
	GRPCPort  string
	SecretKey string
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

type Config struct {
	CIServer CIServerConfig

	DB DatabaseConfig
	MQ MessageQueueConfig

	GitHub ServiceConfig
	GitLab ServiceConfig
}

func NewConfigFromEnv() (Config, error) {
	config := Config{
		CIServer: CIServerConfig{
			Host:      env.StringEnv("HOST", ""),
			Port:      env.StringEnv("PORT", "8000"),
			GRPCPort:  env.StringEnv("GRPC_PORT", "9000"),
			SecretKey: env.StringEnv("SECRET_KEY", ""),
		},
		DB: DatabaseConfig{
			URI: env.StringEnv("DB_URI", "postgres://localhost/shark-ci"),
		},
		MQ: MessageQueueConfig{
			URI: env.StringEnv("MQ_URI", "amqp://guest:guest@localhost"),
		},
		GitHub: ServiceConfig{
			ClientID:     env.StringEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: env.StringEnv("GITHUB_CLIENT_SECRET", ""),
		},
		GitLab: ServiceConfig{
			ClientID:     env.StringEnv("GITLAB_CLIENT_ID", ""),
			ClientSecret: env.StringEnv("GITLAB_CLIENT_SECRET", ""),
		},
	}
	err := config.validate()
	if err != nil {
		return Config{}, err
	}

	return config, err
}

func (c Config) validate() error {
	if c.CIServer.Host == "" {
		return errors.New("HOST is required")
	}
	if c.CIServer.SecretKey == "" {
		return errors.New("SECRET_KEY is required")
	}
	if c.GitHub.ClientID != "" && c.GitHub.ClientSecret == "" {
		return errors.New("GITHUB_CLIENT_SECRET is required when GITHUB_CLIENT_ID is set")
	}
	if c.GitLab.ClientID != "" && c.GitLab.ClientSecret == "" {
		return errors.New("GITLAB_CLIENT_SECRET is required when GITLAB_CLIENT_ID is set")
	}
	if c.GitHub.ClientID == "" && c.GitLab.ClientID == "" {
		return errors.New("either GITHUB_CLIENT_ID or GITLAB_CLIENT_ID must be set")
	}

	return nil
}

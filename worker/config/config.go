package config

import "github.com/FilipSolich/shark-ci/shared/env"

type WorkerConfig struct {
	MaxWorkers int
	ReposPath  string
}

type CIServerConfig struct {
	Host string
	Port string
}

type MessageQueueConfig struct {
	URI string
}

type Config struct {
	Worker   WorkerConfig
	CIServer CIServerConfig

	MQ MessageQueueConfig
}

func NewConfigFromEnv() (Config, error) {
	config := Config{
		Worker: WorkerConfig{
			MaxWorkers: env.IntEnv("MAX_WORKERS", 4),
			ReposPath:  env.StringEnv("REPOS_PATH", "./repos"),
		},
		CIServer: CIServerConfig{
			Host: env.StringEnv("CISERVER_HOST", "localhost"),
			Port: env.StringEnv("CISERVER_PORT", "8080"),
		},
		MQ: MessageQueueConfig{
			URI: env.StringEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
		},
	}

	return config, nil
}

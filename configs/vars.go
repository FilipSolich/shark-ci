package configs

import (
	"errors"
	"os"
)

const CIServer = "CI Server"
const EventHandlerPath = "/event_handler"

var (
	Host          string
	Port          string
	SessionSecret string
	CSRFSecret    string
	WebhookSecret string

	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUsername string
	RabbitMQPassword string

	GitHubService      bool
	GitHubClientID     string
	GitHubClientSecret string

	GitLabService      bool
	GitLabClientID     string
	GitLabClientSecret string
)

func LoadEnv() error {
	Host = getEnv("HOST", "")
	Port = getEnv("PORT", "8080")
	SessionSecret = getEnv("SESSION_SECRET", "insecure-secret")
	CSRFSecret = getEnv("CSRF_SECRET", "insecure-secret")
	WebhookSecret = getEnv("WEBHOOK_SECRET", "insecure-secret")

	RabbitMQHost = getEnv("RABBITMQ_HOST", "localhost")
	RabbitMQPort = getEnv("RABBITMQ_PORT", "5672")
	RabbitMQUsername = getEnv("RABBITMQ_USERNAME", "guest")
	RabbitMQPassword = getEnv("RABBITMQ_PASSWORD", "guest")

	GitHubService = boolEnv(getEnv("GITHUB_SERVICE", "false"))
	GitHubClientID = getEnv("GITHUB_CLIENT_ID", "")
	GitHubClientSecret = getEnv("GITHUB_CLIENT_SECRET", "")

	GitLabService = boolEnv(getEnv("GITLAB_SERVICE", "false"))
	GitLabClientID = getEnv("GITLAB_CLIENT_ID", "")
	GitLabClientSecret = getEnv("GITLAB_CLIENT_SECRET", "")

	return validateEnv()
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func boolEnv(value string) bool {
	if value == "true" {
		return true
	}

	return false
}

func validateEnv() error {
	if len(Host) == 0 {
		return errors.New("HOST must be set")
	}

	if !GitHubService && !GitLabService {
		return errors.New("at least one service (*_SERVICE) must be set as `true`")
	}

	if GitHubService && (GitHubClientID == "" || GitHubClientSecret == "") {
		return errors.New("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set")
	}

	if GitLabService && (GitLabClientID == "" || GitLabClientSecret == "") {
		return errors.New("GITLAB_CLIENT_ID and GITLAB_CLIENT_SECRET must be set")
	}

	return nil
}

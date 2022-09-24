package configs

import (
	"errors"
	"os"
)

var (
	Hostname      string
	Port          string
	SessionSecret string
	CSRFSecret    string
	WebhookSecret string

	GitHubService      bool
	GitHubClientID     string
	GitHubClientSecret string

	GitLabService      bool
	GitLabClientID     string
	GitLabClientSecret string
)

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

func LoadEnv() error {
	Hostname = getEnv("HOSTNAME", "")
	Port = getEnv("PORT", "8080")
	SessionSecret = getEnv("SESSION_SECRET", "insecure-secret")
	CSRFSecret = getEnv("CSRF_SECRET", "insecure-secret")
	WebhookSecret = getEnv("WEBHOOK_SECRET", "insecure-secret")

	GitHubService = boolEnv(getEnv("GITHUB_SERVICE", "false"))
	GitHubClientID = getEnv("GITHUB_CLIENT_ID", "")
	GitHubClientSecret = getEnv("GITHUB_CLIENT_SECRET", "")

	GitLabService = boolEnv(getEnv("GITLAB_SERVICE", "false"))
	GitLabClientID = getEnv("GITLAB_CLIENT_ID", "")
	GitLabClientSecret = getEnv("GITLAB_CLIENT_SECRET", "")

	return validateEnv()
}

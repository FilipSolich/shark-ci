package env

import (
	"os"
	"strconv"
	"strings"
)

func StringEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func IntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}

		return v
	}

	return fallback
}

func BoolEnv(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if strings.ToLower(value) == "true" || value == "1" {
			return true
		}

		return false
	}

	return fallback
}

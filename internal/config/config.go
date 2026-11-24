// Package config loads service configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration for the ratings API.
type Config struct {
	Port          int
	LogLevel      string
	FailReadiness bool
}

// FromEnv builds a Config from environment variables.
func FromEnv() Config {
	return Config{
		Port:          intEnv("PORT", 8083),
		LogLevel:      strEnv("LOG_LEVEL", "info"),
		FailReadiness: boolEnv("FAIL_READINESS", false),
	}
}

func strEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func intEnv(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func boolEnv(key string, def bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

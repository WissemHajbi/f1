package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address   string
	DBPath    string
	UserAgent string
	Timeout   time.Duration
	MaxBody   int64
}

func Load() Config {
	return Config{
		Address:   env("HTTP_ADDRESS", ":8080"),
		DBPath:    env("DB_PATH", "data/oidysts.db"),
		UserAgent: env("UPSTREAM_USER_AGENT", "oidysts/0.1"),
		Timeout:   durationEnv("UPSTREAM_TIMEOUT", 20*time.Second),
		MaxBody:   int64Env("UPSTREAM_MAX_BODY_BYTES", 10<<20),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func int64Env(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

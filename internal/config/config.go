package config

import (
	"os"
	"strings"
	"time"
)

const (
	DefaultPort               = "8080"
	DefaultMaxMessageSize     = 1024
	DefaultTypingMinInterval  = 200 * time.Millisecond
	DefaultMessageMinInterval = 100 * time.Millisecond
)

type Config struct {
	Port               string
	AllowedOrigins     []string
	MaxMessageSize     int64
	TypingMinInterval  time.Duration
	MessageMinInterval time.Duration
}

func Load() Config {
	return Config{
		Port:               stringFromEnv("APP_PORT", DefaultPort),
		AllowedOrigins:     ParseAllowedOrigins(os.Getenv("APP_ALLOWED_ORIGINS")),
		MaxMessageSize:     DefaultMaxMessageSize,
		TypingMinInterval:  DefaultTypingMinInterval,
		MessageMinInterval: DefaultMessageMinInterval,
	}
}

func ParseAllowedOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}

		origins = append(origins, origin)
	}

	return origins
}

func stringFromEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

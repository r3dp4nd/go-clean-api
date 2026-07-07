package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App  AppConfig
	HTTP HTTPConfig
	Log  LogConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
}

type HTTPConfig struct {
	Host              string
	Port              int
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	httpPort, err := getEnvAsInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, err
	}

	httpHost := getEnv("HTTP_HOST", "")

	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "go-clean-api"),
			Version:     getEnv("APP_VERSION", "v0.1.0"),
			Environment: getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Host:              httpHost,
			Port:              httpPort,
			Addr:              fmt.Sprintf("%s:%d", httpHost, httpPort),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %q", key, value)
	}

	return parsedValue, nil
}

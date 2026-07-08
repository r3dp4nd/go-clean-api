package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Log      LogConfig
	CORS     CORSConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
}

type HTTPConfig struct {
	Host                   string
	Port                   int
	Addr                   string
	ReadHeaderTimeout      time.Duration
	ReadTimeout            time.Duration
	WriteTimeout           time.Duration
	IdleTimeout            time.Duration
	ShutdownTimeout        time.Duration
	ShutdownTimeoutSeconds int
}

type LogConfig struct {
	Level  string
	Format string
}

type CORSConfig struct {
	Enabled        bool
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAgeSeconds  int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
	DSN      string
}

func Load() (*Config, error) {
	httpPort, err := getEnvAsInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, err
	}

	shutdownTimeoutSeconds, err := getEnvAsInt("HTTP_SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil {
		return nil, err
	}

	if shutdownTimeoutSeconds <= 0 {
		return nil, fmt.Errorf("HTTP_SHUTDOWN_TIMEOUT_SECONDS must be greater than zero")
	}

	corsEnabled, err := getEnvAsBool("CORS_ENABLED", true)
	if err != nil {
		return nil, err
	}

	corsMaxAgeSeconds, err := getEnvAsInt("CORS_MAX_AGE_SECONDS", 600)
	if err != nil {
		return nil, err
	}

	if corsMaxAgeSeconds < 0 {
		return nil, fmt.Errorf("CORS_MAX_AGE_SECONDS must be greater than or equal to zero")
	}

	dbPort, err := getEnvAsInt("DB_PORT", 5432)
	if err != nil {
		return nil, err
	}

	if dbPort <= 0 {
		return nil, fmt.Errorf("DB_PORT must be greater than zero")
	}

	httpHost := getEnv("HTTP_HOST", "")

	dbConfig := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		Name:     getEnv("DB_NAME", "go_clean_api"),
		User:     getEnv("DB_USER", "app"),
		Password: getEnv("DB_PASSWORD", "app"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	dbConfig.DSN = buildPostgresDSN(dbConfig)

	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "go-clean-api"),
			Version:     getEnv("APP_VERSION", "v0.1.0"),
			Environment: getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Host:                   httpHost,
			Port:                   httpPort,
			Addr:                   fmt.Sprintf("%s:%d", httpHost, httpPort),
			ReadHeaderTimeout:      5 * time.Second,
			ReadTimeout:            10 * time.Second,
			WriteTimeout:           10 * time.Second,
			IdleTimeout:            60 * time.Second,
			ShutdownTimeout:        time.Duration(shutdownTimeoutSeconds) * time.Second,
			ShutdownTimeoutSeconds: shutdownTimeoutSeconds,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		CORS: CORSConfig{
			Enabled: corsEnabled,
			AllowedOrigins: getEnvAsCSV(
				"CORS_ALLOWED_ORIGINS",
				[]string{
					"http://localhost:3000",
					"http://localhost:5173",
					"http://localhost:4200",
				},
			),
			AllowedMethods: getEnvAsCSV(
				"CORS_ALLOWED_METHODS",
				[]string{
					"GET",
					"POST",
					"PUT",
					"DELETE",
					"OPTIONS",
				},
			),
			AllowedHeaders: getEnvAsCSV(
				"CORS_ALLOWED_HEADERS",
				[]string{
					"Content-Type",
					"Authorization",
					"X-Request-ID",
				},
			),
			MaxAgeSeconds: corsMaxAgeSeconds,
		},
		Database: dbConfig,
	}

	return cfg, nil
}

func buildPostgresDSN(cfg DatabaseConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)
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

func getEnvAsBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value for %s: %q", key, value)
	}

	return parsedValue, nil
}

func getEnvAsCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		return fallback
	}

	return items
}

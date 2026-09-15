package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"shortner/internal/server"
	"shortner/internal/service"
	"shortner/internal/storage"
)

type Config struct {
	Server  server.Config
	Storage storage.Config
	Service service.Config
}

func Load() (*Config, error) {
	cfg := &Config{
		Server:  loadServerConfig(),
		Storage: loadStorageConfig(),
		Service: loadServiceConfig(),
	}

	if cfg.Storage.DSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}

	return cfg, nil
}

func loadServerConfig() server.Config {
	return server.Config{
		Address:      getEnv("SERVER_ADDRESS", ":8080"),
		ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 5*time.Second),
		WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
	}
}

func loadStorageConfig() storage.Config {
	return storage.Config{
		DSN:             getEnv("DATABASE_DSN", "postgres://tiffany:password@localhost:5432/linkdb?sslmode=disable"),
		MaxOpenConns:    getEnvInt("DATABASE_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DATABASE_MAX_IDLE_CONNS", 25),
		ConnMaxLifetime: getEnvDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

func loadServiceConfig() service.Config {
	return service.Config{
		BaseURL:         getEnv("LINK_BASE_URL", "http://localhost:8080"),
		ShortLinkLength: getEnvInt("LINK_SHORT_LENGTH", 6),
		MaxRetries:      getEnvInt("LINK_MAX_RETRIES", 3),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

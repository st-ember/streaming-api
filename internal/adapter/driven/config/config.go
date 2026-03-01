package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ConnStr        string
	ServerAdd      string
	StoragePath    string
	WorkerLimit    int
	PollInterval   time.Duration
	WorkerWaitTime time.Duration
}

func Load() *Config {
	return &Config{
		ConnStr:        getEnv("DB_URL", ""),
		ServerAdd:      getEnv("SERVER_ADD", "8085"),
		StoragePath:    getEnv("STORAGE_PATH", "./storage"),
		WorkerLimit:    getEnvInt("WORKER_LIMIT", 5),
		PollInterval:   time.Duration(getEnvInt("POLL_INTERVAL_SEC", 10)) * time.Second,
		WorkerWaitTime: time.Duration(getEnvInt("WORKER_WAIT_SEC", 60)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return fallback
}

package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	BaseURL         string
	DBPath          string
	CodeLength      int
	RateLimitPerMin int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
}

func Load() Config {
	port := getEnv("PORT", "8080")
	cfg := Config{
		Port:            port,
		BaseURL:         strings.TrimRight(getEnv("BASE_URL", "http://localhost:"+port), "/"),
		DBPath:          getEnv("DB_PATH", "shortlink.db"),
		CodeLength:      getEnvInt("CODE_LENGTH", 6),
		RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MIN", 120),
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		MaxBodyBytes:    4 << 10,
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

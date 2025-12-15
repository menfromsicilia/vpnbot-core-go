package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	APIKey          string
	XrayNodeToken   string
	DBPath          string
	RequestTimeout  time.Duration
	NodeTimeout     time.Duration
	LogLevel        string
	LogOutput       string // "stdout" or "file"
	LogFile         string
	LogMaxSize      int // MB
	LogMaxBackups   int
	LogMaxAge       int // days
	LogCompress     bool
}

func Load() *Config {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		APIKey:         getEnv("API_KEY_REQUESTS", ""),
		XrayNodeToken:  getEnv("XRAY_NODE_TOKEN", ""),
		DBPath:         getEnv("DB_PATH", "./vpnbot.db"),
		RequestTimeout: parseDuration(getEnv("REQUEST_TIMEOUT", "10s")),
		NodeTimeout:    parseDuration(getEnv("NODE_TIMEOUT", "3s")),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		LogOutput:      getEnv("LOG_OUTPUT", "stdout"),
		LogFile:        getEnv("LOG_FILE", "./logs/vpnbot.log"),
		LogMaxSize:     parseInt(getEnv("LOG_MAX_SIZE", "100")),
		LogMaxBackups:  parseInt(getEnv("LOG_MAX_BACKUPS", "3")),
		LogMaxAge:      parseInt(getEnv("LOG_MAX_AGE", "7")),
		LogCompress:    parseBool(getEnv("LOG_COMPRESS", "true")),
	}

	if cfg.APIKey == "" {
		log.Fatal("API_KEY_REQUESTS environment variable is required")
	}
	if cfg.XrayNodeToken == "" {
		log.Fatal("XRAY_NODE_TOKEN environment variable is required")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("Invalid duration format: %s", s)
	}
	return d
}

func parseInt(s string) int {
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		log.Fatalf("Invalid integer format: %s", s)
	}
	return i
}

func parseBool(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}


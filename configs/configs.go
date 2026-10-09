package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv         string
	AppPort        string
	Database       DatabaseConfig
	ReservationTTL time.Duration
	SchedulerInterval time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadConfig() *Config {
	return &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "8080"),
		ReservationTTL:    getEnvDuration("RESERVATION_TTL", 5*time.Minute),
		SchedulerInterval: getEnvDuration("SCHEDULER_INTERVAL", 10*time.Second),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "inventory_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(valStr)
	if err != nil {
		if sec, errSec := strconv.Atoi(valStr); errSec == nil {
			return time.Duration(sec) * time.Second
		}
		return defaultVal
	}
	return d
}

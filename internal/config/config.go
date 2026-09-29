package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func Load() Config {
	return Config{
		HTTPAddr:   envOr("HTTP_ADDR", ":8080"),
		DBHost:     envOr("DB_HOST", "localhost"),
		DBPort:     envOr("DB_PORT", "3306"),
		DBUser:     envOr("DB_USER", "pismo"),
		DBPassword: envOr("DB_PASSWORD", "pismopw"),
		DBName:     envOr("DB_NAME", "pismo"),
	}
}

// DSN returns a MySQL data source name in the go-sql-driver format.
// parseTime=true makes DATETIME columns return time.Time instead of []byte.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci&loc=UTC",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

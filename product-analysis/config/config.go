package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
	Port   int
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(getEnvOrDefault("PORT", "50052"))
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	dbPort, err := strconv.Atoi(getEnvOrDefault("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid db port: %v", err)
	}

	return &Config{
		DBHost: getEnvOrDefault("DB_HOST", "localhost"),
		DBPort: dbPort,
		DBUser: getEnvOrDefault("DB_USER", "postgres"),
		DBPass: getEnvOrDefault("DB_PASS", "postgres"),
		DBName: getEnvOrDefault("DB_NAME", "ecommerce"),
		Port:   port,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

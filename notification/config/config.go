package config

import "os"

type Config struct {
	DBHost       string
	DBPort      string
	DBUser      string
	DBPass      string
	DBName      string
	ServerPort  string
	KafkaBroker string
}

func LoadConfig() (*Config, error) {
	return &Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPass:      getEnv("DB_PASS", "postgres"),
		DBName:      getEnv("DB_NAME", "ecommerce"),
		ServerPort:  getEnv("SERVER_PORT", "8082"),
		KafkaBroker: getEnv("KAFKA_BROKERS", "localhost:9092"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

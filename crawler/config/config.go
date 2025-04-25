package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerPort           string
	DBHost               string
	DBPort               int
	DBUser               string
	DBPass               string
	DBName               string
	CrawlerServiceAddr   string
	ProductAnalysisServiceAddr string
	BaseURL              string
}

func LoadConfig() *Config {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	if dbPort == 0 {
		dbPort = 5432
	}

	crawlerPort, _ := strconv.Atoi(os.Getenv("CRAWLER_PORT"))
	if crawlerPort == 0 {
		crawlerPort = 50051
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "postgres"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ecommerce_crawler"
	}

	return &Config{
		ServerPort:           os.Getenv("SERVER_PORT"),
		DBHost:               dbHost,
		DBPort:               dbPort,
		DBUser:               dbUser,
		DBPass:               dbPass,
		DBName:               dbName,
		CrawlerServiceAddr:   fmt.Sprintf(":%d", crawlerPort),
		ProductAnalysisServiceAddr: os.Getenv("PRODUCT_ANALYSIS_SERVICE_ADDR"),
		BaseURL:              os.Getenv("BASE_URL"),
	}
}
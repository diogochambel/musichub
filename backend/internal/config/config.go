package config

import (
	"os"
)

type Config struct {
	MongoURI    string
	DBName      string
	ServerPort  string
	FrontendURL string
}

func Load() *Config {
	return &Config{
		MongoURI:    getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:      getEnv("DB_NAME", "psi_db"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:4200"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	PostgresConnectionUrl string
	RiotApiKey            string
}

func LoadConfig() Config {
	// .env là optional: không có file thì dùng env của hệ thống
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	return Config{
		Port:                  port,
		PostgresConnectionUrl: os.Getenv("POSTGRES_CONNECTION_URL"),
		RiotApiKey:            os.Getenv("RIOT_API_KEY"),
	}
}

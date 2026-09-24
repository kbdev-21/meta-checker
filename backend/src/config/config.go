package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Trần thật của Riot API key. Rate limiter tự chừa đệm an toàn bên dưới các số này.
const (
	defaultRiotRatePerSec  = 20
	defaultRiotRatePer2Min = 100
)

type Config struct {
	Port                  string
	PostgresConnectionUrl string
	RiotApiKey            string
	RiotTimelineApiKey    string // key thứ 2, chỉ dùng fetch match timeline
	RiotRatePerSec        int
	RiotRatePer2Min       int
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
		RiotTimelineApiKey:    os.Getenv("RIOT_TIMELINE_API_KEY"),
		RiotRatePerSec:        intEnv("RIOT_RATE_PER_SEC", defaultRiotRatePerSec),
		RiotRatePer2Min:       intEnv("RIOT_RATE_PER_2MIN", defaultRiotRatePer2Min),
	}
}

// ---------- private ----------

// Thiếu hoặc không parse được thì dùng giá trị mặc định.
func intEnv(key string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

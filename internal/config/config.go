package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	LiveKitURL       string
	LiveKitAPIKey    string
	LiveKitAPISecret string

	PingooServerInternalURL string
	PingooInternalSecret    string
}

func Load() (*Config, error) {
	env := getEnv("APP_ENV", "dev")
	envFile := ".env." + env
	_ = godotenv.Load(env)

	if err := godotenv.Load(envFile); err == nil {
		fmt.Printf("loaded env file: %s\n", envFile)
	}

	cfg := &Config{
		Port: getEnv("PORT", "8081"),

		LiveKitURL:       os.Getenv("LIVEKIT_URL"),
		LiveKitAPIKey:    os.Getenv("LIVEKIT_API_KEY"),
		LiveKitAPISecret: os.Getenv("LIVEKIT_API_SECRET"),

		PingooServerInternalURL: os.Getenv("PINGOO_SERVER_INTERNAL_URL"),
		PingooInternalSecret:    os.Getenv("PINGOO_INTERNAL_SECRET"),
	}

	if cfg.LiveKitURL == "" {
		return nil, fmt.Errorf("LIVEKIT_URL is required")
	}

	if cfg.LiveKitAPIKey == "" {
		return nil, fmt.Errorf("LIVEKIT_API_KEY is required")
	}

	if cfg.LiveKitAPISecret == "" {
		return nil, fmt.Errorf("LIVEKIT_API_SECRET is required")
	}

	if cfg.PingooInternalSecret == "" {
		return nil, fmt.Errorf("PINGOO_INTERNAL_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

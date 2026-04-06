package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string

	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string

	OpenAIAPIURL string
	OpenAIAPIKey string
	OpenAIModel  string

	S3Endpoint     string
	S3Region       string
	S3PathStyle    bool
	S3Bucket       string
	S3AccessKey    string
	S3SecretKey    string
	S3PublicPrefix string

	TelegramBotToken string
	B24WebhookURL    string
	B24APIKey        string

	WorkerPoolSize int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/build_assistant"),
		Port:             getEnv("PORT", "8080"),
		OIDCIssuer:       getEnv("OIDC_ISSUER", ""),
		OIDCClientID:     getEnv("OIDC_CLIENT_ID", ""),
		OIDCClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:  getEnv("OIDC_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		OpenAIAPIURL:     getEnv("OPENAI_API_URL", "https://openrouter.ai/api/v1"),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:      getEnv("OPENAI_MODEL", "qwen/qwen3-235b-a22b-2507"),
		S3Endpoint:       getEnv("S3_ENDPOINT", ""),
		S3Region:         getEnv("S3_REGION", "us-east-1"),
		S3PathStyle:      getEnvBool("S3_PATH_STYLE", false),
		S3Bucket:         getEnv("S3_BUCKET", ""),
		S3AccessKey:      getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:      getEnv("S3_SECRET_KEY", ""),
		S3PublicPrefix:   getEnv("S3_PUBLIC_PREFIX", ""),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		B24WebhookURL:    getEnv("B24_WEBHOOK_URL", ""),
		B24APIKey:        getEnv("B24_API_KEY", ""),
		WorkerPoolSize:   getEnvInt("WORKER_POOL_SIZE", 5),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

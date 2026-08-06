package config

import "os"
import "github.com/joho/godotenv"

type Config struct {
	DBHost		string
	DBPort		string
	DBUser		string
	DBPassword	string
	DBName		string
	GeminiAPIKey string
	ResendAPIKey string
}

func Load() Config{

	_ = godotenv.Load()

	return Config{
	DBHost:		getEnv("DB_HOST", "localhost"),
	DBPort:		getEnv("DB_PORT", "5432"),
	DBUser:		getEnv("DB_USER", "postgres"),
	DBPassword:		getEnv("DB_PASSWORD", "postgres"),
	DBName:		getEnv("DB_NAME", "todo_db"),
	GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
	ResendAPIKey: getEnv("RESEND_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

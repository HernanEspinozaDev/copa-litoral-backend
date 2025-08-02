package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost             string
	DBPort             int
	DBUser             string
	DBPassword         string
	DBName             string
	APIPort            int
	JWTSecret          string
	CORSAllowedOrigins string
	Environment        string
}

func LoadConfig() *Config {
	// Cargar variables de entorno desde .env
	err := godotenv.Load()
	if err != nil {
		log.Printf("Advertencia: No se pudo cargar el archivo .env: %v", err)
	}

	config := &Config{}

	// Cargar variables de entorno con valores por defecto
	config.DBHost = getEnv("DB_HOST", "localhost")
	config.DBPort = getEnvAsInt("DB_PORT", 5432)
	config.DBUser = getEnv("DB_USER", "nandev")
	config.DBPassword = getEnv("DB_PASSWORD", "Admin1234")
	config.DBName = getEnv("DB_NAME", "copa_litoral")
	config.APIPort = getEnvAsInt("API_PORT", 8089)
	config.JWTSecret = getEnv("JWT_SECRET", "supersecretkeyforexample")
	config.CORSAllowedOrigins = getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")
	config.Environment = getEnv("ENVIRONMENT", "development")

	// Validar variables críticas solo en producción
	if config.JWTSecret == "supersecretkeyforexample" {
		log.Printf("Advertencia: JWT_SECRET está usando valor por defecto. Cambia esto en producción.")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Advertencia: %s no es un número válido, usando valor por defecto: %d", key, defaultValue)
	}
	return defaultValue
} 
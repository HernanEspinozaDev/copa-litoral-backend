package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	APIPort             string
	JWTSecret           string
	CORSAllowedOrigins  string
	Environment         string
	LogLevel            string
	LogFile             string
	// Database Pool Configuration
	DBMaxOpenConns      int
	DBMaxIdleConns      int
	DBConnMaxLifetime   int // minutes
	DBConnMaxIdleTime   int // minutes
	// Backup Configuration
	BackupEnabled       bool
	BackupInterval      int    // hours
	BackupRetention     int    // number of backups to keep
	BackupDirectory     string
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
	config.DBPort = getEnv("DB_PORT", "5432")
	config.DBUser = getEnv("DB_USER", "nandev")
	config.DBPassword = getEnv("DB_PASSWORD", "Admin1234")
	config.DBName = getEnv("DB_NAME", "copa_litoral")
	config.APIPort = getEnv("API_PORT", "8089")
	config.JWTSecret = getEnv("JWT_SECRET", "supersecretkeyforexample")
	config.CORSAllowedOrigins = getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")
	config.Environment = getEnv("ENVIRONMENT", "development")
	config.LogLevel = getEnv("LOG_LEVEL", "info")
	config.LogFile = getEnv("LOG_FILE", "")

	// Database Pool Configuration
	config.DBMaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	config.DBMaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	config.DBConnMaxLifetime = getEnvAsInt("DB_CONN_MAX_LIFETIME", 5) // minutes
	config.DBConnMaxIdleTime = getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 2) // minutes

	// Backup Configuration
	config.BackupEnabled = getEnvAsBool("BACKUP_ENABLED", false)
	config.BackupInterval = getEnvAsInt("BACKUP_INTERVAL", 24) // hours
	config.BackupRetention = getEnvAsInt("BACKUP_RETENTION", 7) // backups
	config.BackupDirectory = getEnv("BACKUP_DIRECTORY", "backups")

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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
} 
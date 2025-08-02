package tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"copa-litoral-backend/config"
	"copa-litoral-backend/database"

	_ "github.com/lib/pq"
)

// TestConfig configuración para testing
type TestConfig struct {
	DB     *sql.DB
	Config *config.Config
}

var testConfig *TestConfig

// SetupTestDB configura la base de datos de pruebas
func SetupTestDB() *TestConfig {
	if testConfig != nil {
		return testConfig
	}

	// Configuración de pruebas
	cfg := &config.Config{
		DBHost:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		DBPassword: getEnvOrDefault("TEST_DB_PASSWORD", "password"),
		DBName:     getEnvOrDefault("TEST_DB_NAME", "copa_litoral_test"),
		JWTSecret:  "test-jwt-secret-key-for-testing-only",
		APIPort:    "8081",
		Environment: "test",
		CORSAllowedOrigins: "http://localhost:3000",
		LogLevel:   "error", // Reducir logs en testing
	}

	// Conectar a la base de datos de pruebas
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	testConfig = &TestConfig{
		DB:     db,
		Config: cfg,
	}

	return testConfig
}

// TeardownTestDB limpia la base de datos de pruebas
func TeardownTestDB() {
	if testConfig != nil && testConfig.DB != nil {
		testConfig.DB.Close()
		testConfig = nil
	}
}

// CleanupTestData limpia los datos de prueba de las tablas
func CleanupTestData(db *sql.DB) error {
	tables := []string{
		"partidos",
		"jugadores",
		"usuarios",
		"torneos",
		"categorias",
		"schema_migrations",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clean table %s: %w", table, err)
		}
	}

	return nil
}

// CreateTestUser crea un usuario de prueba
func CreateTestUser(db *sql.DB, username, email, rol string) (int, error) {
	var userID int
	query := `
		INSERT INTO usuarios (nombre_usuario, email, password_hash, rol, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	err := db.QueryRow(query, username, email, "hashed_password", rol).Scan(&userID)
	return userID, err
}

// CreateTestJugador crea un jugador de prueba
func CreateTestJugador(db *sql.DB, nombre, apellido string, categoriaID *int) (int, error) {
	var jugadorID int
	query := `
		INSERT INTO jugadores (nombre, apellido, categoria_id, club, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	err := db.QueryRow(query, nombre, apellido, categoriaID, "Club Test").Scan(&jugadorID)
	return jugadorID, err
}

// CreateTestCategoria crea una categoría de prueba
func CreateTestCategoria(db *sql.DB, nombre string) (int, error) {
	var categoriaID int
	query := `
		INSERT INTO categorias (nombre, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		RETURNING id
	`
	err := db.QueryRow(query, nombre).Scan(&categoriaID)
	return categoriaID, err
}

// CreateTestTorneo crea un torneo de prueba
func CreateTestTorneo(db *sql.DB, nombre string, anio int) (int, error) {
	var torneoID int
	query := `
		INSERT INTO torneos (nombre, anio, fecha_inicio, fecha_fin, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW() + INTERVAL '30 days', NOW(), NOW())
		RETURNING id
	`
	err := db.QueryRow(query, nombre, anio).Scan(&torneoID)
	return torneoID, err
}

// RunTestWithCleanup ejecuta una prueba con limpieza automática
func RunTestWithCleanup(t *testing.T, testFunc func(*testing.T, *TestConfig)) {
	cfg := SetupTestDB()
	
	// Limpiar datos antes de la prueba
	if err := CleanupTestData(cfg.DB); err != nil {
		t.Fatalf("Failed to cleanup test data: %v", err)
	}

	// Ejecutar la prueba
	testFunc(t, cfg)

	// Limpiar datos después de la prueba
	if err := CleanupTestData(cfg.DB); err != nil {
		t.Errorf("Failed to cleanup test data after test: %v", err)
	}
}

// getEnvOrDefault obtiene una variable de entorno o devuelve un valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// AssertEqual verifica que dos valores sean iguales
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotNil verifica que un valor no sea nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	if value == nil {
		t.Errorf("%s: expected non-nil value", message)
	}
}

// AssertNil verifica que un valor sea nil
func AssertNil(t *testing.T, value interface{}, message string) {
	if value != nil {
		t.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertNoError verifica que no haya error
func AssertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertError verifica que haya un error
func AssertError(t *testing.T, err error, message string) {
	if err == nil {
		t.Errorf("%s: expected error, got nil", message)
	}
}

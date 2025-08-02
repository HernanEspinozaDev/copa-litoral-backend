package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"copa-litoral-backend/config"
	"copa-litoral-backend/routes"
	"copa-litoral-backend/utils"

	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()
	log.Println("Configuración cargada exitosamente")

	// Inicializar logger estructurado
	utils.InitLogger(cfg)
	utils.LogInfo("Logger inicializado", map[string]interface{}{
		"environment": cfg.Environment,
		"log_level":   cfg.LogLevel,
	})

	// Inicializar métricas
	utils.InitMetrics()
	utils.LogInfo("Métricas inicializadas", nil)

	// Conectar a la base de datos con pool mejorado
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		utils.LogError("Failed to open database connection", err, nil)
		return
	}
	defer db.Close()
	
	// Configurar pool de conexiones
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.DBConnMaxIdleTime) * time.Minute)
	
	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		utils.LogError("Failed to ping database", err, nil)
		return
	}
	
	// Ejecutar migraciones si están habilitadas
	if err := runMigrations(db); err != nil {
		utils.LogError("Failed to run migrations", err, nil)
	}
	
	// Configurar health checks con la conexión de base de datos
	utils.SetGlobalDB(db)
	
	utils.LogInfo("Database connection pool configured", map[string]interface{}{
		"max_open_conns": cfg.DBMaxOpenConns,
		"max_idle_conns": cfg.DBMaxIdleConns,
	})

	// Crear router
	router := mux.NewRouter()

	// Configurar rutas
	routes.SetupRoutes(router, cfg)

	// Crear servidor HTTP
	server := &http.Server{
		Addr:         ":" + fmt.Sprintf("%d", cfg.APIPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	utils.LogInfo("Servidor configurado", map[string]interface{}{
		"port":         cfg.APIPort,
		"environment": cfg.Environment,
	})

	// Canal para errores del servidor
	serverErrors := make(chan error, 1)

	// Iniciar servidor en una goroutine separada
	go func() {
		utils.LogInfo("Servidor iniciando", map[string]interface{}{
			"port": cfg.APIPort,
			"url":  fmt.Sprintf("http://localhost:%d", cfg.APIPort),
		})
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Configurar manejo de señales para graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Esperar por señal de cierre o error del servidor
	select {
	case err := <-serverErrors:
		utils.LogError("Error del servidor", err, nil)
		return
	case sig := <-quit:
		utils.LogInfo("Señal de cierre recibida", map[string]interface{}{
			"signal": sig.String(),
		})
	}

	// Crear contexto con timeout para el shutdown
	shutdownTimeout := 30 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	utils.LogInfo("Iniciando graceful shutdown", map[string]interface{}{
		"timeout": shutdownTimeout.String(),
	})

	// Intentar cerrar el servidor gracefully
	if err := server.Shutdown(ctx); err != nil {
		utils.LogError("Error durante graceful shutdown", err, nil)
		return
	}

	utils.LogInfo("Servidor cerrado correctamente", nil)
}

// runMigrations ejecuta las migraciones de base de datos
func runMigrations(db *sql.DB) error {
	// Crear tabla de migraciones si no existe
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// Verificar si la migración inicial ya fue aplicada
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	
	if count > 0 {
		utils.LogInfo("Migrations already applied", nil)
		return nil
	}
	
	// Aplicar migración inicial básica (crear tablas principales si no existen)
	initialMigration := `
	-- Tabla para las categorías
	CREATE TABLE IF NOT EXISTS categorias (
		id SERIAL PRIMARY KEY,
		nombre VARCHAR(100) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Tabla para los jugadores
	CREATE TABLE IF NOT EXISTS jugadores (
		id SERIAL PRIMARY KEY,
		nombre VARCHAR(255) NOT NULL,
		apellido VARCHAR(255) NOT NULL,
		telefono_wsp VARCHAR(50),
		contacto_visible_en_web BOOLEAN DEFAULT FALSE,
		categoria_id INTEGER REFERENCES categorias(id) ON DELETE SET NULL,
		club VARCHAR(255),
		estado_participacion VARCHAR(50) DEFAULT 'Activo',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Tabla para los usuarios
	CREATE TABLE IF NOT EXISTS usuarios (
		id SERIAL PRIMARY KEY,
		nombre_usuario VARCHAR(255) NOT NULL UNIQUE,
		email VARCHAR(255) UNIQUE,
		password_hash TEXT NOT NULL,
		rol VARCHAR(50) NOT NULL DEFAULT 'jugador',
		jugador_id INTEGER UNIQUE REFERENCES jugadores(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Índices básicos
	CREATE INDEX IF NOT EXISTS idx_jugadores_categoria ON jugadores (categoria_id);
	CREATE INDEX IF NOT EXISTS idx_usuarios_jugador ON usuarios (jugador_id);
	`
	
	// Ejecutar migración en transacción
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer tx.Rollback()
	
	if _, err := tx.Exec(initialMigration); err != nil {
		return fmt.Errorf("failed to execute initial migration: %w", err)
	}
	
	// Registrar migración como aplicada
	if _, err := tx.Exec("INSERT INTO schema_migrations (version, name) VALUES (1, 'initial_schema')"); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}
	
	utils.LogInfo("Initial migration applied successfully", nil)
	return nil
}
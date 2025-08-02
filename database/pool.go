package database

import (
	"database/sql"
	"fmt"
	"time"

	"copa-litoral-backend/config"
	"copa-litoral-backend/utils"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// DBPool representa un pool de conexiones a la base de datos
type DBPool struct {
	DB     *sql.DB
	config *config.Config
	logger *logrus.Logger
}

// NewDBPool crea un nuevo pool de conexiones con configuración optimizada
func NewDBPool(cfg *config.Config) (*DBPool, error) {
	logger := utils.Logger
	
	// Construir string de conexión
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	// Abrir conexión
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.WithError(err).Error("Failed to open database connection")
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configurar pool de conexiones
	pool := &DBPool{
		DB:     db,
		config: cfg,
		logger: logger,
	}

	if err := pool.configurePool(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}

	// Verificar conexión
	if err := pool.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection pool initialized successfully")
	return pool, nil
}

// configurePool configura los parámetros del pool de conexiones
func (p *DBPool) configurePool() error {
	// Configuración del pool basada en el entorno
	var (
		maxOpenConns    int
		maxIdleConns    int
		connMaxLifetime time.Duration
		connMaxIdleTime time.Duration
	)

	if p.config.Environment == "production" {
		// Configuración para producción
		maxOpenConns = 25    // Máximo 25 conexiones abiertas
		maxIdleConns = 10    // Máximo 10 conexiones idle
		connMaxLifetime = 5 * time.Minute  // Vida máxima de conexión
		connMaxIdleTime = 2 * time.Minute  // Tiempo máximo idle
	} else {
		// Configuración para desarrollo
		maxOpenConns = 10    // Máximo 10 conexiones abiertas
		maxIdleConns = 5     // Máximo 5 conexiones idle
		connMaxLifetime = 3 * time.Minute  // Vida máxima de conexión
		connMaxIdleTime = 1 * time.Minute  // Tiempo máximo idle
	}

	// Aplicar configuración
	p.DB.SetMaxOpenConns(maxOpenConns)
	p.DB.SetMaxIdleConns(maxIdleConns)
	p.DB.SetConnMaxLifetime(connMaxLifetime)
	p.DB.SetConnMaxIdleTime(connMaxIdleTime)

	p.logger.WithFields(logrus.Fields{
		"max_open_conns":     maxOpenConns,
		"max_idle_conns":     maxIdleConns,
		"conn_max_lifetime":  connMaxLifetime,
		"conn_max_idle_time": connMaxIdleTime,
		"environment":        p.config.Environment,
	}).Info("Database connection pool configured")

	return nil
}

// Ping verifica la conectividad con la base de datos
func (p *DBPool) Ping() error {
	ctx, cancel := utils.GetContextWithTimeout(5 * time.Second)
	defer cancel()

	if err := p.DB.PingContext(ctx); err != nil {
		p.logger.WithError(err).Error("Database ping failed")
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats obtiene estadísticas del pool de conexiones
func (p *DBPool) GetStats() sql.DBStats {
	return p.DB.Stats()
}

// LogStats registra las estadísticas del pool en los logs
func (p *DBPool) LogStats() {
	stats := p.GetStats()
	p.logger.WithFields(logrus.Fields{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}).Debug("Database connection pool stats")
}

// Close cierra el pool de conexiones
func (p *DBPool) Close() error {
	p.logger.Info("Closing database connection pool")
	if err := p.DB.Close(); err != nil {
		p.logger.WithError(err).Error("Error closing database connection pool")
		return fmt.Errorf("failed to close database pool: %w", err)
	}
	return nil
}

// HealthCheck verifica el estado de salud de la base de datos
func (p *DBPool) HealthCheck() error {
	// Ping básico
	if err := p.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Verificar estadísticas del pool
	stats := p.GetStats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open connections available")
	}

	// Test query simple
	ctx, cancel := utils.GetContextWithTimeout(3 * time.Second)
	defer cancel()

	var result int
	err := p.DB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected test query result: %d", result)
	}

	return nil
}

// RunMigrations ejecuta las migraciones de base de datos
func (p *DBPool) RunMigrations() error {
	migrationManager := NewMigrationManager(p.DB, "database/migrations")
	return migrationManager.RunMigrations()
}

// GetMigrationStatus obtiene el estado de las migraciones
func (p *DBPool) GetMigrationStatus() ([]map[string]interface{}, error) {
	migrationManager := NewMigrationManager(p.DB, "database/migrations")
	return migrationManager.GetMigrationStatus()
}

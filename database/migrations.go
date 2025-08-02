package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"copa-litoral-backend/utils"

	"github.com/sirupsen/logrus"
)

// Migration representa una migración de base de datos
type Migration struct {
	Version   int
	Name      string
	UpSQL     string
	DownSQL   string
	Timestamp time.Time
}

// MigrationManager maneja las migraciones de base de datos
type MigrationManager struct {
	db             *sql.DB
	migrationsPath string
	logger         *logrus.Logger
}

// NewMigrationManager crea un nuevo manager de migraciones
func NewMigrationManager(db *sql.DB, migrationsPath string) *MigrationManager {
	return &MigrationManager{
		db:             db,
		migrationsPath: migrationsPath,
		logger:         utils.GetLogger(),
	}
}

// InitMigrationsTable crea la tabla de migraciones si no existe
func (mm *MigrationManager) InitMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := mm.db.Exec(query)
	if err != nil {
		mm.logger.WithError(err).Error("Error creating migrations table")
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	mm.logger.Info("Migrations table initialized")
	return nil
}

// LoadMigrations carga todas las migraciones desde el directorio
func (mm *MigrationManager) LoadMigrations() ([]Migration, error) {
	files, err := ioutil.ReadDir(mm.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") && strings.Contains(file.Name(), "_up.sql") {
			migration, err := mm.parseMigrationFile(file.Name())
			if err != nil {
				mm.logger.WithError(err).WithField("file", file.Name()).Error("Error parsing migration file")
				continue
			}
			migrations = append(migrations, migration)
		}
	}

	// Ordenar por versión
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFile parsea un archivo de migración
func (mm *MigrationManager) parseMigrationFile(filename string) (Migration, error) {
	// Formato esperado: 001_create_users_up.sql
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return Migration{}, fmt.Errorf("invalid version in filename %s: %w", filename, err)
	}

	name := strings.Join(parts[1:len(parts)-2], "_") // Excluir version y "up.sql"

	// Leer archivo UP
	upPath := filepath.Join(mm.migrationsPath, filename)
	upSQL, err := ioutil.ReadFile(upPath)
	if err != nil {
		return Migration{}, fmt.Errorf("failed to read up migration file %s: %w", filename, err)
	}

	// Leer archivo DOWN (opcional)
	downFilename := strings.Replace(filename, "_up.sql", "_down.sql", 1)
	downPath := filepath.Join(mm.migrationsPath, downFilename)
	downSQL, _ := ioutil.ReadFile(downPath) // No error si no existe

	return Migration{
		Version: version,
		Name:    name,
		UpSQL:   string(upSQL),
		DownSQL: string(downSQL),
	}, nil
}

// GetAppliedMigrations obtiene las migraciones ya aplicadas
func (mm *MigrationManager) GetAppliedMigrations() (map[int]bool, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := mm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// RunMigrations ejecuta todas las migraciones pendientes
func (mm *MigrationManager) RunMigrations() error {
	if err := mm.InitMigrationsTable(); err != nil {
		return err
	}

	migrations, err := mm.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := mm.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var executed int
	for _, migration := range migrations {
		if applied[migration.Version] {
			mm.logger.WithFields(logrus.Fields{
				"version": migration.Version,
				"name":    migration.Name,
			}).Debug("Migration already applied, skipping")
			continue
		}

		mm.logger.WithFields(logrus.Fields{
			"version": migration.Version,
			"name":    migration.Name,
		}).Info("Applying migration")

		if err := mm.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}

		executed++
	}

	if executed > 0 {
		mm.logger.WithField("count", executed).Info("Migrations completed successfully")
	} else {
		mm.logger.Info("No pending migrations to apply")
	}

	return nil
}

// applyMigration aplica una migración específica
func (mm *MigrationManager) applyMigration(migration Migration) error {
	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Ejecutar la migración
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Registrar la migración como aplicada
	insertQuery := "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)"
	if _, err := tx.Exec(insertQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}

// RollbackMigration deshace una migración específica
func (mm *MigrationManager) RollbackMigration(version int) error {
	migrations, err := mm.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	var targetMigration *Migration
	for _, migration := range migrations {
		if migration.Version == version {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration version %d not found", version)
	}

	if targetMigration.DownSQL == "" {
		return fmt.Errorf("no down migration available for version %d", version)
	}

	mm.logger.WithFields(logrus.Fields{
		"version": version,
		"name":    targetMigration.Name,
	}).Info("Rolling back migration")

	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", err)
	}
	defer tx.Rollback()

	// Ejecutar rollback
	if _, err := tx.Exec(targetMigration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Eliminar registro de migración
	deleteQuery := "DELETE FROM schema_migrations WHERE version = $1"
	if _, err := tx.Exec(deleteQuery, version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	mm.logger.WithField("version", version).Info("Migration rolled back successfully")
	return nil
}

// GetMigrationStatus obtiene el estado de todas las migraciones
func (mm *MigrationManager) GetMigrationStatus() ([]map[string]interface{}, error) {
	migrations, err := mm.LoadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := mm.GetAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var status []map[string]interface{}
	for _, migration := range migrations {
		status = append(status, map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
			"applied": applied[migration.Version],
		})
	}

	return status, nil
}

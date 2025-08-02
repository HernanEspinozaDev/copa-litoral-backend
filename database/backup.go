package database

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"copa-litoral-backend/config"
	"copa-litoral-backend/utils"

	"github.com/sirupsen/logrus"
)

// BackupManager maneja los respaldos de base de datos
type BackupManager struct {
	config    *config.Config
	logger    *logrus.Logger
	backupDir string
}

// BackupOptions opciones para configurar respaldos
type BackupOptions struct {
	IncludeSchema bool
	IncludeData   bool
	Compress      bool
	Tables        []string // Tablas específicas, vacío = todas
	MaxBackups    int      // Número máximo de backups a mantener
}

// NewBackupManager crea un nuevo manager de respaldos
func NewBackupManager(cfg *config.Config, backupDir string) *BackupManager {
	if backupDir == "" {
		backupDir = "backups"
	}

	return &BackupManager{
		config:    cfg,
		logger:    utils.Logger,
		backupDir: backupDir,
	}
}

// CreateBackup crea un respaldo de la base de datos
func (bm *BackupManager) CreateBackup(opts *BackupOptions) (string, error) {
	if opts == nil {
		opts = &BackupOptions{
			IncludeSchema: true,
			IncludeData:   true,
			Compress:      true,
			MaxBackups:    10,
		}
	}

	// Crear directorio de backups si no existe
	if err := os.MkdirAll(bm.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generar nombre de archivo
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("copa_litoral_backup_%s.sql", timestamp)
	if opts.Compress {
		filename += ".gz"
	}
	backupPath := filepath.Join(bm.backupDir, filename)

	bm.logger.WithFields(logrus.Fields{
		"backup_path":    backupPath,
		"include_schema": opts.IncludeSchema,
		"include_data":   opts.IncludeData,
		"compress":       opts.Compress,
		"tables":         opts.Tables,
	}).Info("Starting database backup")

	start := time.Now()

	// Construir comando pg_dump
	args := bm.buildPgDumpArgs(opts, backupPath)
	cmd := exec.Command("pg_dump", args...)

	// Configurar variables de entorno para pg_dump
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", bm.config.DBPassword),
	)

	// Ejecutar backup
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		bm.logger.WithFields(logrus.Fields{
			"error":    err,
			"output":   string(output),
			"duration": duration,
		}).Error("Database backup failed")
		return "", fmt.Errorf("backup failed: %w, output: %s", err, string(output))
	}

	// Verificar que el archivo se creó
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return "", fmt.Errorf("backup file was not created: %s", backupPath)
	}

	// Obtener tamaño del archivo
	fileInfo, _ := os.Stat(backupPath)
	fileSize := fileInfo.Size()

	bm.logger.WithFields(logrus.Fields{
		"backup_path": backupPath,
		"file_size":   fileSize,
		"duration":    duration,
	}).Info("Database backup completed successfully")

	// Limpiar backups antiguos
	if err := bm.cleanOldBackups(opts.MaxBackups); err != nil {
		bm.logger.WithError(err).Warn("Failed to clean old backups")
	}

	// Registrar métrica
	utils.RecordDBQuery("backup", "database", duration)

	return backupPath, nil
}

// buildPgDumpArgs construye los argumentos para pg_dump
func (bm *BackupManager) buildPgDumpArgs(opts *BackupOptions, backupPath string) []string {
	args := []string{
		"-h", bm.config.DBHost,
		"-p", bm.config.DBPort,
		"-U", bm.config.DBUser,
		"-d", bm.config.DBName,
		"--verbose",
		"--no-password", // Usamos PGPASSWORD env var
	}

	// Opciones de contenido
	if !opts.IncludeSchema {
		args = append(args, "--data-only")
	} else if !opts.IncludeData {
		args = append(args, "--schema-only")
	}

	// Tablas específicas
	if len(opts.Tables) > 0 {
		for _, table := range opts.Tables {
			args = append(args, "-t", table)
		}
	}

	// Compresión y archivo de salida
	if opts.Compress {
		args = append(args, "--compress=9")
		args = append(args, "-f", backupPath)
	} else {
		args = append(args, "-f", backupPath)
	}

	return args
}

// RestoreBackup restaura una base de datos desde un respaldo
func (bm *BackupManager) RestoreBackup(backupPath string, dropExisting bool) error {
	// Verificar que el archivo existe
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	bm.logger.WithFields(logrus.Fields{
		"backup_path":   backupPath,
		"drop_existing": dropExisting,
	}).Info("Starting database restore")

	start := time.Now()

	// Si se requiere, eliminar la base de datos existente
	if dropExisting {
		if err := bm.dropDatabase(); err != nil {
			return fmt.Errorf("failed to drop existing database: %w", err)
		}
		if err := bm.createDatabase(); err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Construir comando de restauración
	var cmd *exec.Cmd
	if strings.HasSuffix(backupPath, ".gz") {
		// Archivo comprimido
		cmd = exec.Command("sh", "-c", fmt.Sprintf("gunzip -c %s | psql", backupPath))
	} else {
		// Archivo sin comprimir
		cmd = exec.Command("psql", "-f", backupPath)
	}

	// Configurar variables de entorno
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGHOST=%s", bm.config.DBHost),
		fmt.Sprintf("PGPORT=%s", bm.config.DBPort),
		fmt.Sprintf("PGUSER=%s", bm.config.DBUser),
		fmt.Sprintf("PGPASSWORD=%s", bm.config.DBPassword),
		fmt.Sprintf("PGDATABASE=%s", bm.config.DBName),
	)

	// Ejecutar restauración
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		bm.logger.WithFields(logrus.Fields{
			"error":    err,
			"output":   string(output),
			"duration": duration,
		}).Error("Database restore failed")
		return fmt.Errorf("restore failed: %w, output: %s", err, string(output))
	}

	bm.logger.WithFields(logrus.Fields{
		"backup_path": backupPath,
		"duration":    duration,
	}).Info("Database restore completed successfully")

	return nil
}

// dropDatabase elimina la base de datos
func (bm *BackupManager) dropDatabase() error {
	cmd := exec.Command("dropdb",
		"-h", bm.config.DBHost,
		"-p", bm.config.DBPort,
		"-U", bm.config.DBUser,
		"--if-exists",
		bm.config.DBName,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", bm.config.DBPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to drop database: %w, output: %s", err, string(output))
	}

	return nil
}

// createDatabase crea la base de datos
func (bm *BackupManager) createDatabase() error {
	cmd := exec.Command("createdb",
		"-h", bm.config.DBHost,
		"-p", bm.config.DBPort,
		"-U", bm.config.DBUser,
		bm.config.DBName,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", bm.config.DBPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create database: %w, output: %s", err, string(output))
	}

	return nil
}

// cleanOldBackups elimina backups antiguos manteniendo solo los más recientes
func (bm *BackupManager) cleanOldBackups(maxBackups int) error {
	if maxBackups <= 0 {
		return nil
	}

	// Obtener lista de archivos de backup
	files, err := filepath.Glob(filepath.Join(bm.backupDir, "copa_litoral_backup_*.sql*"))
	if err != nil {
		return fmt.Errorf("failed to list backup files: %w", err)
	}

	if len(files) <= maxBackups {
		return nil // No hay necesidad de limpiar
	}

	// Ordenar archivos por fecha de modificación (más reciente primero)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			path:    file,
			modTime: info.ModTime(),
		})
	}

	// Ordenar por fecha de modificación (más reciente primero)
	for i := 0; i < len(fileInfos)-1; i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.Before(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	// Eliminar archivos antiguos
	var deletedCount int
	for i := maxBackups; i < len(fileInfos); i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			bm.logger.WithError(err).WithField("file", fileInfos[i].path).Warn("Failed to delete old backup")
		} else {
			deletedCount++
		}
	}

	if deletedCount > 0 {
		bm.logger.WithFields(logrus.Fields{
			"deleted_count": deletedCount,
			"max_backups":   maxBackups,
		}).Info("Cleaned old backup files")
	}

	return nil
}

// ListBackups lista todos los archivos de backup disponibles
func (bm *BackupManager) ListBackups() ([]map[string]interface{}, error) {
	files, err := filepath.Glob(filepath.Join(bm.backupDir, "copa_litoral_backup_*.sql*"))
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}

	var backups []map[string]interface{}
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		backups = append(backups, map[string]interface{}{
			"path":         file,
			"filename":     filepath.Base(file),
			"size":         info.Size(),
			"created_at":   info.ModTime(),
			"is_compressed": strings.HasSuffix(file, ".gz"),
		})
	}

	return backups, nil
}

// ScheduleBackup programa un backup automático (implementación básica)
func (bm *BackupManager) ScheduleBackup(interval time.Duration, opts *BackupOptions) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			backupPath, err := bm.CreateBackup(opts)
			if err != nil {
				bm.logger.WithError(err).Error("Scheduled backup failed")
			} else {
				bm.logger.WithField("backup_path", backupPath).Info("Scheduled backup completed")
			}
		}
	}()

	bm.logger.WithField("interval", interval).Info("Backup scheduler started")
}

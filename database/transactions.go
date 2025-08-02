package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"copa-litoral-backend/utils"

	"github.com/sirupsen/logrus"
)

// TxManager maneja las transacciones de base de datos
type TxManager struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewTxManager crea un nuevo manager de transacciones
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{
		db:     db,
		logger: utils.Logger,
	}
}

// TxFunc representa una función que se ejecuta dentro de una transacción
type TxFunc func(*sql.Tx) error

// WithTransaction ejecuta una función dentro de una transacción con rollback automático
func (tm *TxManager) WithTransaction(ctx context.Context, fn TxFunc) error {
	return tm.WithTransactionAndOptions(ctx, nil, fn)
}

// TxOptions opciones para configurar transacciones
type TxOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
	Timeout   time.Duration
}

// WithTransactionAndOptions ejecuta una función dentro de una transacción con opciones específicas
func (tm *TxManager) WithTransactionAndOptions(ctx context.Context, opts *TxOptions, fn TxFunc) error {
	// Configurar timeout por defecto
	if opts == nil {
		opts = &TxOptions{
			Timeout: 30 * time.Second,
		}
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	// Crear contexto con timeout
	txCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Configurar opciones de transacción
	txOpts := &sql.TxOptions{
		Isolation: opts.Isolation,
		ReadOnly:  opts.ReadOnly,
	}

	// Iniciar transacción
	tx, err := tm.db.BeginTx(txCtx, txOpts)
	if err != nil {
		tm.logger.WithError(err).Error("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Configurar rollback automático
	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				tm.logger.WithError(rollbackErr).Error("Failed to rollback transaction after panic")
			}
			panic(p) // Re-throw panic
		}
	}()

	// Ejecutar función
	start := time.Now()
	err = fn(tx)
	duration := time.Since(start)

	if err != nil {
		// Rollback en caso de error
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			tm.logger.WithFields(logrus.Fields{
				"original_error": err,
				"rollback_error": rollbackErr,
				"duration":       duration,
			}).Error("Failed to rollback transaction")
			return fmt.Errorf("transaction failed and rollback failed: original error: %w, rollback error: %v", err, rollbackErr)
		}

		tm.logger.WithFields(logrus.Fields{
			"error":    err,
			"duration": duration,
		}).Debug("Transaction rolled back due to error")

		return fmt.Errorf("transaction failed: %w", err)
	}

	// Commit transacción
	if err := tx.Commit(); err != nil {
		tm.logger.WithFields(logrus.Fields{
			"error":    err,
			"duration": duration,
		}).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tm.logger.WithField("duration", duration).Debug("Transaction committed successfully")

	// Actualizar métricas
	utils.RecordDBQuery("transaction", "database", duration)

	return nil
}

// WithReadOnlyTransaction ejecuta una función dentro de una transacción de solo lectura
func (tm *TxManager) WithReadOnlyTransaction(ctx context.Context, fn TxFunc) error {
	opts := &TxOptions{
		ReadOnly: true,
		Timeout:  15 * time.Second, // Timeout más corto para lecturas
	}
	return tm.WithTransactionAndOptions(ctx, opts, fn)
}

// WithSerializableTransaction ejecuta una función con nivel de aislamiento SERIALIZABLE
func (tm *TxManager) WithSerializableTransaction(ctx context.Context, fn TxFunc) error {
	opts := &TxOptions{
		Isolation: sql.LevelSerializable,
		Timeout:   45 * time.Second, // Timeout más largo para transacciones serializables
	}
	return tm.WithTransactionAndOptions(ctx, opts, fn)
}

// BatchInsert realiza inserción por lotes dentro de una transacción
func (tm *TxManager) BatchInsert(ctx context.Context, query string, values [][]interface{}) error {
	return tm.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare batch insert statement: %w", err)
		}
		defer stmt.Close()

		for i, row := range values {
			if _, err := stmt.ExecContext(ctx, row...); err != nil {
				return fmt.Errorf("failed to execute batch insert at row %d: %w", i, err)
			}
		}

		tm.logger.WithField("rows_inserted", len(values)).Debug("Batch insert completed")
		return nil
	})
}

// BatchUpdate realiza actualización por lotes dentro de una transacción
func (tm *TxManager) BatchUpdate(ctx context.Context, query string, values [][]interface{}) error {
	return tm.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare batch update statement: %w", err)
		}
		defer stmt.Close()

		var totalAffected int64
		for i, row := range values {
			result, err := stmt.ExecContext(ctx, row...)
			if err != nil {
				return fmt.Errorf("failed to execute batch update at row %d: %w", i, err)
			}

			affected, _ := result.RowsAffected()
			totalAffected += affected
		}

		tm.logger.WithFields(logrus.Fields{
			"rows_processed": len(values),
			"rows_affected":  totalAffected,
		}).Debug("Batch update completed")

		return nil
	})
}

// ExecuteInTransaction ejecuta múltiples queries dentro de una transacción
func (tm *TxManager) ExecuteInTransaction(ctx context.Context, queries []string, args [][]interface{}) error {
	if len(queries) != len(args) {
		return fmt.Errorf("number of queries (%d) must match number of argument sets (%d)", len(queries), len(args))
	}

	return tm.WithTransaction(ctx, func(tx *sql.Tx) error {
		for i, query := range queries {
			_, err := tx.ExecContext(ctx, query, args[i]...)
			if err != nil {
				return fmt.Errorf("failed to execute query %d: %w", i, err)
			}
		}

		tm.logger.WithField("queries_executed", len(queries)).Debug("Multiple queries executed in transaction")
		return nil
	})
}

// TransactionStats estadísticas de transacciones
type TransactionStats struct {
	ActiveTransactions int
	TotalCommitted     int64
	TotalRolledBack    int64
	AverageDuration    time.Duration
}

// GetTransactionStats obtiene estadísticas de transacciones (simuladas para este ejemplo)
func (tm *TxManager) GetTransactionStats() TransactionStats {
	// En una implementación real, estas estadísticas se mantendrían en memoria
	// o se obtendrían de métricas de Prometheus
	return TransactionStats{
		ActiveTransactions: 0, // Se obtendría de un contador global
		TotalCommitted:     0, // Se obtendría de métricas
		TotalRolledBack:    0, // Se obtendría de métricas
		AverageDuration:    0, // Se calcularía de métricas históricas
	}
}

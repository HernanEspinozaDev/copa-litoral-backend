package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// HealthStatus representa el estado de salud de un componente
type HealthStatus struct {
	Status    string                 `json:"status"`
	Component string                 `json:"component"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// OverallHealth representa el estado general del sistema
type OverallHealth struct {
	Status     string         `json:"status"`
	Timestamp  time.Time      `json:"timestamp"`
	Uptime     string         `json:"uptime"`
	Version    string         `json:"version"`
	Components []HealthStatus `json:"components"`
}

var startTime = time.Now()
var globalDB *sql.DB

// SetGlobalDB establece la conexión global de base de datos para health checks
func SetGlobalDB(db *sql.DB) {
	globalDB = db
}

// HealthHandler maneja las requests de health check completo
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Realizar todos los checks
	checks := []HealthStatus{
		checkDatabase(),
		checkMemory(),
		checkSystem(),
	}
	
	// Determinar estado general
	overallStatus := "healthy"
	for _, check := range checks {
		if check.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		} else if check.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}
	
	// Crear respuesta
	health := OverallHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Uptime:     time.Since(startTime).String(),
		Version:    "1.0.0",
		Components: checks,
	}
	
	// Establecer código de respuesta HTTP
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusOK // Degraded pero aún funcional
	}
	
	// Responder
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
	
	// Log del health check
	duration := time.Since(start)
	LogInfo("Health check completed", map[string]interface{}{
		"status":   overallStatus,
		"duration": duration,
		"checks":   len(checks),
	})
}

// ReadinessHandler verifica si el sistema está listo para recibir tráfico
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check básico de base de datos
	dbCheck := checkDatabase()
	
	status := "ready"
	statusCode := http.StatusOK
	
	if dbCheck.Status == "unhealthy" {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}
	
	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"database":  dbCheck.Status,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// LivenessHandler verifica si el sistema está vivo
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime).String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// checkDatabase verifica el estado de la base de datos
func checkDatabase() HealthStatus {
	if globalDB == nil {
		return HealthStatus{
			Status:    "unhealthy",
			Component: "database",
			Message:   "Database connection not initialized",
			Timestamp: time.Now(),
		}
	}
	
	// Ping con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	if err := globalDB.PingContext(ctx); err != nil {
		return HealthStatus{
			Status:    "unhealthy",
			Component: "database",
			Message:   "Database ping failed: " + err.Error(),
			Timestamp: time.Now(),
		}
	}
	
	// Obtener estadísticas
	stats := globalDB.Stats()
	details := map[string]interface{}{
		"open_connections": stats.OpenConnections,
		"in_use":          stats.InUse,
		"idle":            stats.Idle,
		"wait_count":      stats.WaitCount,
	}
	
	// Determinar estado basado en estadísticas
	status := "healthy"
	message := "Database connection is healthy"
	
	if stats.OpenConnections == 0 {
		status = "unhealthy"
		message = "No database connections available"
	} else if float64(stats.InUse)/float64(stats.OpenConnections) > 0.8 {
		status = "degraded"
		message = "Database connection pool usage is high"
	}
	
	return HealthStatus{
		Status:    status,
		Component: "database",
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// checkMemory verifica el uso de memoria
func checkMemory() HealthStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Convertir a MB
	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024
	
	details := map[string]interface{}{
		"alloc_mb":     allocMB,
		"sys_mb":       sysMB,
		"num_gc":       m.NumGC,
		"goroutines":   runtime.NumGoroutine(),
	}
	
	status := "healthy"
	message := "Memory usage is normal"
	
	// Alertas básicas de memoria
	if allocMB > 500 { // Más de 500MB
		status = "degraded"
		message = "High memory usage detected"
	}
	
	if allocMB > 1000 { // Más de 1GB
		status = "unhealthy"
		message = "Critical memory usage detected"
	}
	
	return HealthStatus{
		Status:    status,
		Component: "memory",
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// checkSystem verifica el estado general del sistema
func checkSystem() HealthStatus {
	details := map[string]interface{}{
		"uptime":     time.Since(startTime).String(),
		"goroutines": runtime.NumGoroutine(),
		"version":    runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
	
	status := "healthy"
	message := "System is running normally"
	
	// Check básico de goroutines
	if runtime.NumGoroutine() > 1000 {
		status = "degraded"
		message = "High number of goroutines detected"
	}
	
	return HealthStatus{
		Status:    status,
		Component: "system",
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

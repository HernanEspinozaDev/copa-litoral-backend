package utils

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"copa-litoral-backend/config"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger inicializa el logger estructurado
func InitLogger(cfg *config.Config) {
	Logger = logrus.New()

	// Configurar formato
	if cfg.Environment == "production" {
		// JSON format para producción
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		// Formato texto para desarrollo
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}

	// Configurar nivel de log
	level := getLogLevel(cfg.LogLevel)
	Logger.SetLevel(level)

	// Configurar salida
	if cfg.LogFile != "" {
		setupLogFile(cfg.LogFile)
	} else {
		Logger.SetOutput(os.Stdout)
	}

	// Log inicial
	Logger.WithFields(logrus.Fields{
		"environment": cfg.Environment,
		"log_level":   level.String(),
		"version":     "1.0.0",
	}).Info("Logger inicializado correctamente")
}

// getLogLevel convierte string a logrus.Level
func getLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// setupLogFile configura el archivo de log
func setupLogFile(logFile string) {
	// Crear directorio si no existe
	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		Logger.WithError(err).Warn("No se pudo crear directorio de logs")
		Logger.SetOutput(os.Stdout)
		return
	}

	// Abrir archivo de log
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.WithError(err).Warn("No se pudo abrir archivo de log")
		Logger.SetOutput(os.Stdout)
		return
	}

	Logger.SetOutput(file)
}

// LoggingMiddleware middleware para logging de requests
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrapper para capturar status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Procesar request
			next.ServeHTTP(wrapped, r)

			// Log de la request
			duration := time.Since(start)
			
			entry := Logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      wrapped.statusCode,
				"duration":    duration.String(),
				"duration_ms": duration.Milliseconds(),
				"ip":          getClientIP(r),
				"user_agent":  r.UserAgent(),
				"size":        wrapped.size,
			})

			if wrapped.statusCode >= 500 {
				entry.Error("Server error")
			} else if wrapped.statusCode >= 400 {
				entry.Warn("Client error")
			} else {
				entry.Info("Request completed")
			}
		})
	}
}

// responseWriter wrapper para capturar status code y size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// getClientIP extrae la IP del cliente
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}

// Helper functions para logging estructurado

// LogInfo log de información
func LogInfo(message string, fields logrus.Fields) {
	Logger.WithFields(fields).Info(message)
}

// LogWarn log de advertencia
func LogWarn(message string, fields logrus.Fields) {
	Logger.WithFields(fields).Warn(message)
}

// LogError log de error
func LogError(message string, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	Logger.WithFields(fields).Error(message)
}

// LogDebug log de debug
func LogDebug(message string, fields logrus.Fields) {
	Logger.WithFields(fields).Debug(message)
}

// LogFatal log fatal (termina la aplicación)
func LogFatal(message string, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	Logger.WithFields(fields).Fatal(message)
}

// GetContextWithTimeout crea un contexto con timeout
func GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

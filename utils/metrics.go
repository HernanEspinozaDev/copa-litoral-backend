package utils

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Database metrics
	dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	dbQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Authentication metrics
	authAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"result"}, // success, failure
	)

	// Rate limiting metrics
	rateLimitHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint", "ip"},
	)

	// Application metrics
	appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information",
		},
		[]string{"version", "environment"},
	)
)

// InitMetrics inicializa las métricas de Prometheus
func InitMetrics() {
	// Registrar métricas
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
		dbConnectionsActive,
		dbConnectionsIdle,
		dbQueriesTotal,
		dbQueryDuration,
		authAttemptsTotal,
		rateLimitHitsTotal,
		appInfo,
	)

	// Establecer información de la aplicación
	appInfo.WithLabelValues("1.0.0", "development").Set(1)
}

// MetricsMiddleware middleware para capturar métricas HTTP
func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Incrementar requests en vuelo
			httpRequestsInFlight.Inc()
			defer httpRequestsInFlight.Dec()

			// Wrapper para capturar status code
			wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: 200}

			// Procesar request
			next.ServeHTTP(wrapped, r)

			// Obtener endpoint pattern
			route := mux.CurrentRoute(r)
			endpoint := "unknown"
			if route != nil {
				if template, err := route.GetPathTemplate(); err == nil {
					endpoint = template
				}
			}

			// Registrar métricas
			duration := time.Since(start)
			status := strconv.Itoa(wrapped.statusCode)

			httpRequestsTotal.WithLabelValues(r.Method, endpoint, status).Inc()
			httpRequestDuration.WithLabelValues(r.Method, endpoint, status).Observe(duration.Seconds())
		})
	}
}

// metricsResponseWriter wrapper para capturar status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (mrw *metricsResponseWriter) WriteHeader(code int) {
	mrw.statusCode = code
	mrw.ResponseWriter.WriteHeader(code)
}

// Helper functions para métricas específicas

// RecordAuthAttempt registra un intento de autenticación
func RecordAuthAttempt(success bool) {
	result := "failure"
	if success {
		result = "success"
	}
	authAttemptsTotal.WithLabelValues(result).Inc()
}

// RecordRateLimitHit registra un hit de rate limiting
func RecordRateLimitHit(endpoint, ip string) {
	rateLimitHitsTotal.WithLabelValues(endpoint, ip).Inc()
}

// RecordDBQuery registra una query de base de datos
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateDBConnections actualiza las métricas de conexiones de BD
func UpdateDBConnections(active, idle int) {
	dbConnectionsActive.Set(float64(active))
	dbConnectionsIdle.Set(float64(idle))
}

// GetMetricsHandler retorna el handler de métricas de Prometheus
func GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

package middlewares

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"copa-litoral-backend/utils"
	"golang.org/x/time/rate"
)

// RateLimiter estructura para manejar rate limiting por IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
		cleanup:  time.Minute * 10, // Limpiar cada 10 minutos
	}
	
	// Iniciar goroutine de limpieza
	go rl.cleanupRoutine()
	
	return rl
}

// getLimiter obtiene o crea un limiter para una IP específica
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}
	
	return limiter
}

// cleanupRoutine limpia limiters inactivos
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		for ip, limiter := range rl.limiters {
			// Si el limiter no ha sido usado recientemente, eliminarlo
			if limiter.Tokens() == float64(rl.burst) {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow verifica si una IP puede hacer una request
func (rl *RateLimiter) Allow(ip string) bool {
	limiter := rl.getLimiter(ip)
	return limiter.Allow()
}

// getClientIP extrae la IP real del cliente
func getClientIP(r *http.Request) string {
	// Verificar headers de proxy
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Tomar la primera IP de la lista
		if ips := parseForwardedFor(forwarded); len(ips) > 0 {
			return ips[0]
		}
	}
	
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	
	// Usar RemoteAddr como fallback
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	
	return ip
}

// parseForwardedFor parsea el header X-Forwarded-For
func parseForwardedFor(forwarded string) []string {
	var ips []string
	for _, ip := range strings.Split(forwarded, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// RateLimitMiddleware middleware para rate limiting
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			
			if !rl.Allow(ip) {
				utils.RespondWithError(w, http.StatusTooManyRequests, "Demasiadas requests. Intenta más tarde.")
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Configuraciones predefinidas de rate limiting

// BasicRateLimit: 100 requests per minute
func BasicRateLimit() *RateLimiter {
	return NewRateLimiter(rate.Every(time.Minute/100), 10)
}

// StrictRateLimit: 30 requests per minute
func StrictRateLimit() *RateLimiter {
	return NewRateLimiter(rate.Every(time.Minute/30), 5)
}

// AuthRateLimit: 10 login attempts per minute
func AuthRateLimit() *RateLimiter {
	return NewRateLimiter(rate.Every(time.Minute/10), 3)
}

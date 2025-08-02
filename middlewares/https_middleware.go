package middlewares

import (
	"net/http"
	"strings"

	"copa-litoral-backend/config"
)

// HTTPSRedirectMiddleware redirige HTTP a HTTPS en producción
func HTTPSRedirectMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Solo aplicar en producción
			if cfg.Environment == "production" {
				// Verificar si la request no es HTTPS
				if r.Header.Get("X-Forwarded-Proto") != "https" && r.TLS == nil {
					// Construir URL HTTPS
					httpsURL := "https://" + r.Host + r.RequestURI
					http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware agrega headers de seguridad
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Headers de seguridad básicos
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Content Security Policy básico
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			
			// HSTS (HTTP Strict Transport Security) para HTTPS
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequestValidationMiddleware valida requests básicas
func RequestValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validar tamaño del body (máximo 10MB)
			r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
			
			// Validar Content-Type para requests POST/PUT
			if r.Method == "POST" || r.Method == "PUT" {
				contentType := r.Header.Get("Content-Type")
				if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
					http.Error(w, "Content-Type debe ser application/json", http.StatusUnsupportedMediaType)
					return
				}
			}
			
			// Validar User-Agent (rechazar requests sin User-Agent)
			if r.Header.Get("User-Agent") == "" {
				http.Error(w, "User-Agent requerido", http.StatusBadRequest)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

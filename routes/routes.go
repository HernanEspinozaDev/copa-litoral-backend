package routes

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"copa-litoral-backend/config"
	"copa-litoral-backend/handlers"
	"copa-litoral-backend/middlewares"
	"copa-litoral-backend/utils"
)

func SetupRoutes(db *sql.DB, cfg *config.Config) *mux.Router {
	r := mux.NewRouter()

	// Middlewares globales básicos
	r.Use(middlewares.CORS(strings.Split(cfg.CORSAllowedOrigins, ",")))
	// r.Use(middlewares.HTTPSRedirectMiddleware(cfg)) // Comentado: Nginx Proxy Manager maneja HTTPS
	r.Use(utils.LoggingMiddleware())
	r.Use(utils.MetricsMiddleware())
	r.Use(middlewares.RateLimitMiddleware(middlewares.NewRateLimiter(100, 1)))

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	}).Methods("GET")

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(nil, cfg) // TODO: Inject proper service

	// Rutas públicas básicas
	public := r.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	public.HandleFunc("/auth/register", authHandler.Register).Methods("POST")

	// Rutas protegidas básicas
	protected := r.PathPrefix("/api/v1/protected").Subrouter()
	protected.Use(middlewares.AuthMiddleware(cfg))
	protected.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Profile endpoint"})
	}).Methods("GET")

	return r
}
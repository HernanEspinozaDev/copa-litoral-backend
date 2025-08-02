package routes

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"copa-litoral-backend/config"
	"copa-litoral-backend/handlers"
	"copa-litoral-backend/middlewares"
	"copa-litoral-backend/utils"
)

func SetupRoutes(db *sql.DB, cfg *config.Config) *mux.Router {
	r := mux.NewRouter()

	// Inicializar managers
	versionManager := utils.NewVersionManager()
	filterManager := utils.NewFilterManager()

	// Middlewares globales
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.HTTPSRedirectMiddleware(cfg))
	r.Use(utils.LoggingMiddleware())
	r.Use(utils.MetricsMiddleware())
	r.Use(middlewares.RateLimitMiddleware(middlewares.NewRateLimiter(100, 1)))
	r.Use(versionManager.VersionMiddleware())
	r.Use(utils.ContentNegotiationMiddleware())

	// Documentación API
	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/static/")))).Methods("GET")
	r.HandleFunc("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./docs/swagger.json")
	}).Methods("GET")

	// Rutas públicas versionadas
	public := r.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/auth/login", handlers.LoginHandler(db, cfg.JWTSecret)).Methods("POST")
	public.HandleFunc("/auth/register", handlers.RegisterHandler(db, cfg.JWTSecret)).Methods("POST")
	public.HandleFunc("/jugadores", handlers.GetJugadoresHandler(db)).Methods("GET")
	public.HandleFunc("/jugadores/{id}", handlers.GetJugadorHandler(db)).Methods("GET")
	public.HandleFunc("/partidos", handlers.GetPartidosHandler(db)).Methods("GET")
	public.HandleFunc("/partidos/{id}", handlers.GetPartidoHandler(db)).Methods("GET")
	public.HandleFunc("/torneos", handlers.GetTorneosHandler(db)).Methods("GET")
	public.HandleFunc("/torneos/{id}", handlers.GetTorneoHandler(db)).Methods("GET")
	public.HandleFunc("/categorias", handlers.GetCategoriasHandler(db)).Methods("GET")
	public.HandleFunc("/categorias/{id}", handlers.GetCategoriaHandler(db)).Methods("GET")

	// Rutas protegidas (requieren autenticación) versionadas
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middlewares.AuthMiddleware(cfg))
	protected.Use(middlewares.RateLimitMiddleware(middlewares.NewRateLimiter(50, 1)))
	protected.HandleFunc("/profile", handlers.GetProfileHandler(db)).Methods("GET")
	protected.HandleFunc("/profile", handlers.UpdateProfileHandler(db)).Methods("PUT")
	protected.HandleFunc("/partidos/{id}/resultado", handlers.ReportarResultadoHandler(db)).Methods("POST")

	// Rutas administrativas (requieren rol admin) versionadas
	admin := r.PathPrefix("/api/v1/admin").Subrouter()
	admin.Use(middlewares.AuthMiddleware(cfg))
	admin.Use(middlewares.RoleMiddleware([]string{"administrador"}))
	admin.Use(middlewares.RateLimitMiddleware(middlewares.NewRateLimiter(30, 1)))
	admin.HandleFunc("/jugadores", handlers.CreateJugadorHandler(db)).Methods("POST")
	admin.HandleFunc("/jugadores/{id}", handlers.UpdateJugadorHandler(db)).Methods("PUT")
	admin.HandleFunc("/jugadores/{id}", handlers.DeleteJugadorHandler(db)).Methods("DELETE")
	admin.HandleFunc("/partidos", handlers.CreatePartidoHandler(db)).Methods("POST")
	admin.HandleFunc("/partidos/{id}", handlers.UpdatePartidoHandler(db)).Methods("PUT")
	admin.HandleFunc("/partidos/{id}", handlers.DeletePartidoHandler(db)).Methods("DELETE")
	admin.HandleFunc("/partidos/{id}/aprobar", handlers.AprobarResultadoHandler(db)).Methods("POST")
	admin.HandleFunc("/torneos", handlers.CreateTorneoHandler(db)).Methods("POST")
	admin.HandleFunc("/torneos/{id}", handlers.UpdateTorneoHandler(db)).Methods("PUT")
	admin.HandleFunc("/torneos/{id}", handlers.DeleteTorneoHandler(db)).Methods("DELETE")
	admin.HandleFunc("/usuarios", handlers.GetUsuariosHandler(db)).Methods("GET")
	admin.HandleFunc("/usuarios/{id}", handlers.UpdateUsuarioHandler(db)).Methods("PUT")
	admin.HandleFunc("/usuarios/{id}", handlers.DeleteUsuarioHandler(db)).Methods("DELETE")

	return r
}
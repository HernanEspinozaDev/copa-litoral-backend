package routes

import (
	"net/http"
	"strings"

	"copa-litoral-backend/config"
	"copa-litoral-backend/handlers"
	"copa-litoral-backend/middlewares"
	"copa-litoral-backend/services"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, cfg *config.Config) {
	// Inicializar servicios
	authService := services.NewAuthService(cfg)
	jugadorService := services.NewJugadorService()
	partidoService := services.NewPartidoService()
	torneoService := services.NewTorneoService()
	categoriaService := services.NewCategoriaService()

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService, cfg)
	jugadorHandler := handlers.NewJugadorHandler(jugadorService)
	partidoHandler := handlers.NewPartidoHandler(partidoService)
	torneoHandler := handlers.NewTorneoHandler(torneoService)
	categoriaHandler := handlers.NewCategoriaHandler(categoriaService)

	// Configurar middlewares de seguridad globales
	router.Use(middlewares.SecurityHeadersMiddleware())
	router.Use(middlewares.HTTPSRedirectMiddleware(cfg))
	router.Use(middlewares.RequestValidationMiddleware())
	
	// Rate limiting básico para todas las rutas
	basicRateLimit := middlewares.BasicRateLimit()
	router.Use(middlewares.RateLimitMiddleware(basicRateLimit))
	
	// Configurar CORS
	corsOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
	router.Use(middlewares.CORS(corsOrigins))

	// Rate limiting estricto para autenticación
	authRateLimit := middlewares.AuthRateLimit()
	authRouter := router.PathPrefix("/api/v1").Subrouter()
	authRouter.Use(middlewares.RateLimitMiddleware(authRateLimit))
	
	// Rutas públicas de autenticación con rate limiting estricto
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	
	// Endpoint de prueba
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Servidor funcionando correctamente"}`))
	}).Methods("GET")

	// Rutas públicas (sin autenticación)
	router.HandleFunc("/api/v1/jugadores", jugadorHandler.GetJugadores).Methods("GET")
	router.HandleFunc("/api/v1/jugadores/{id}", jugadorHandler.GetJugador).Methods("GET")
	router.HandleFunc("/api/v1/partidos", partidoHandler.GetPartidos).Methods("GET")
	router.HandleFunc("/api/v1/partidos/{id}", partidoHandler.GetPartido).Methods("GET")
	router.HandleFunc("/api/v1/torneos", torneoHandler.GetTorneos).Methods("GET")
	router.HandleFunc("/api/v1/torneos/{id}", torneoHandler.GetTorneo).Methods("GET")
	router.HandleFunc("/api/v1/categorias", categoriaHandler.GetCategorias).Methods("GET")
	router.HandleFunc("/api/v1/categorias/{id}", categoriaHandler.GetCategoria).Methods("GET")

	// Crear subrouter para rutas protegidas
	protectedRouter := router.PathPrefix("/api/v1").Subrouter()
	protectedRouter.Use(middlewares.AuthMiddleware(cfg))

	// Rutas protegidas para jugadores (admin)
	protectedRouter.HandleFunc("/admin/jugadores", jugadorHandler.CreateJugador).Methods("POST")
	protectedRouter.HandleFunc("/admin/jugadores/{id}", jugadorHandler.UpdateJugador).Methods("PUT")
	protectedRouter.HandleFunc("/admin/jugadores/{id}", jugadorHandler.DeleteJugador).Methods("DELETE")

	// Rutas protegidas para partidos (admin)
	protectedRouter.HandleFunc("/admin/partidos", partidoHandler.CreatePartido).Methods("POST")
	protectedRouter.HandleFunc("/admin/partidos/{id}", partidoHandler.UpdatePartido).Methods("PUT")
	protectedRouter.HandleFunc("/admin/partidos/{id}", partidoHandler.DeletePartido).Methods("DELETE")
	protectedRouter.HandleFunc("/admin/partidos/{id}/approve-result", partidoHandler.ApproveResult).Methods("PUT")

	// Rutas protegidas para torneos (admin)
	protectedRouter.HandleFunc("/admin/torneos", torneoHandler.CreateTorneo).Methods("POST")
	protectedRouter.HandleFunc("/admin/torneos/{id}", torneoHandler.UpdateTorneo).Methods("PUT")
	protectedRouter.HandleFunc("/admin/torneos/{id}", torneoHandler.DeleteTorneo).Methods("DELETE")

	// Rutas protegidas para categorías (admin)
	protectedRouter.HandleFunc("/admin/categorias", categoriaHandler.CreateCategoria).Methods("POST")
	protectedRouter.HandleFunc("/admin/categorias/{id}", categoriaHandler.UpdateCategoria).Methods("PUT")
	protectedRouter.HandleFunc("/admin/categorias/{id}", categoriaHandler.DeleteCategoria).Methods("DELETE")

	// Rutas protegidas para jugadores (funcionalidades específicas)
	protectedRouter.HandleFunc("/player/partidos/{id}/propose-time", partidoHandler.ProposeMatchTime).Methods("POST")
	protectedRouter.HandleFunc("/player/partidos/{id}/accept-time", partidoHandler.AcceptMatchTime).Methods("POST")
	protectedRouter.HandleFunc("/player/partidos/{id}/report-result", partidoHandler.ReportMatchResult).Methods("POST")

	// Middleware de roles para rutas específicas de administrador
	adminRouter := protectedRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(func(next http.Handler) http.Handler {
		return middlewares.RoleMiddleware([]string{"administrador"}, next)
	})

	// Aplicar middleware de roles a las rutas de admin
	// (Las rutas ya están definidas arriba, pero aquí se aplica el middleware adicional)
	adminRouter.HandleFunc("/jugadores", jugadorHandler.CreateJugador).Methods("POST")
	adminRouter.HandleFunc("/jugadores/{id}", jugadorHandler.UpdateJugador).Methods("PUT")
	adminRouter.HandleFunc("/jugadores/{id}", jugadorHandler.DeleteJugador).Methods("DELETE")
	adminRouter.HandleFunc("/partidos", partidoHandler.CreatePartido).Methods("POST")
	adminRouter.HandleFunc("/partidos/{id}", partidoHandler.UpdatePartido).Methods("PUT")
	adminRouter.HandleFunc("/partidos/{id}", partidoHandler.DeletePartido).Methods("DELETE")
	adminRouter.HandleFunc("/partidos/{id}/approve-result", partidoHandler.ApproveResult).Methods("PUT")
	adminRouter.HandleFunc("/torneos", torneoHandler.CreateTorneo).Methods("POST")
	adminRouter.HandleFunc("/torneos/{id}", torneoHandler.UpdateTorneo).Methods("PUT")
	adminRouter.HandleFunc("/torneos/{id}", torneoHandler.DeleteTorneo).Methods("DELETE")
	adminRouter.HandleFunc("/categorias", categoriaHandler.CreateCategoria).Methods("POST")
	adminRouter.HandleFunc("/categorias/{id}", categoriaHandler.UpdateCategoria).Methods("PUT")
	adminRouter.HandleFunc("/categorias/{id}", categoriaHandler.DeleteCategoria).Methods("DELETE")
} 
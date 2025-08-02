package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"copa-litoral-backend/config"
	"copa-litoral-backend/database"
	"copa-litoral-backend/routes"

	"github.com/gorilla/mux"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()
	log.Println("Configuración cargada exitosamente")

	// Inicializar conexión a la base de datos
	database.InitDB(cfg)
	defer database.CloseDB()

	// Crear router
	router := mux.NewRouter()

	// Configurar rutas
	routes.SetupRoutes(router, cfg)

	// Crear servidor HTTP
	server := &http.Server{
		Addr:         ":" + fmt.Sprintf("%d", cfg.APIPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Printf("Configurando servidor en puerto %d", cfg.APIPort)

	// Iniciar servidor en una goroutine separada
	go func() {
		log.Printf("Servidor iniciando en el puerto %d", cfg.APIPort)
		log.Printf("Servidor escuchando en http://localhost:%d", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar el servidor: %v", err)
		}
	}()

	// Configurar canal para señales de interrupción
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Esperar señal de interrupción
	<-quit
	log.Println("Cerrando servidor...")

	// Crear contexto con timeout para el apagado
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Intentar apagar el servidor de forma elegante
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error al apagar el servidor: %v", err)
	}

	log.Println("Servidor cerrado exitosamente")
} 
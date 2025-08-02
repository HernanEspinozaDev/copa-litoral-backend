package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"copa-litoral-backend/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(config *config.Config) {
	// Construir la cadena de conexión DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	// Abrir la conexión a la base de datos
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error al abrir la conexión a la base de datos: %v", err)
	}

	// Verificar la conexión
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}

	// Configurar el pool de conexiones
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Conexión a la base de datos establecida exitosamente")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Conexión a la base de datos cerrada")
	}
} 
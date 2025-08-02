package models

import (
	"database/sql"
	"time"
)

type Auspiciador struct {
	ID          int            `json:"id"`
	Nombre      string         `json:"nombre"`
	LogoURL     sql.NullString `json:"logo_url"`
	EnlaceWeb   sql.NullString `json:"enlace_web"`
	Descripcion sql.NullString `json:"descripcion"`
	Activo      bool           `json:"activo"`
	Orden       int            `json:"orden"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
} 
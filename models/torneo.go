package models

import (
	"database/sql"
	"time"
)

type Torneo struct {
	ID              int            `json:"id"`
	Nombre          string         `json:"nombre"`
	Anio            int            `json:"anio"`
	FechaInicio     sql.NullTime   `json:"fecha_inicio"`
	FechaFin        sql.NullTime   `json:"fecha_fin"`
	FotoURL         string         `json:"foto_url"`
	FraseDestacada  string         `json:"frase_destacada"`
	Activo          bool           `json:"activo"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
} 
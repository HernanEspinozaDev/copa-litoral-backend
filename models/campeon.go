package models

import (
	"database/sql"
	"time"
)

type Campeon struct {
	ID          int            `json:"id"`
	TorneoID    int            `json:"torneo_id"`
	CategoriaID int            `json:"categoria_id"`
	JugadorID   sql.NullInt32  `json:"jugador_id"`
	Anio        int            `json:"anio"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
} 
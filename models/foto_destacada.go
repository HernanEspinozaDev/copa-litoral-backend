package models

import (
	"database/sql"
	"time"
)

type FotoDestacada struct {
	ID          int            `json:"id"`
	Titulo      sql.NullString `json:"titulo"`
	Descripcion sql.NullString `json:"descripcion"`
	URL         string         `json:"url"`
	PartidoID   sql.NullInt32  `json:"partido_id"`
	TorneoID    sql.NullInt32  `json:"torneo_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
} 
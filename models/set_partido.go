package models

import (
	"database/sql"
	"time"
)

type SetPartido struct {
	ID            int            `json:"id"`
	PartidoID     int            `json:"partido_id"`
	NumeroSet     int            `json:"numero_set"`
	ScoreJugador1 int            `json:"score_jugador1"`
	ScoreJugador2 int            `json:"score_jugador2"`
	TieBreakJ1    sql.NullInt32  `json:"tie_break_j1"`
	TieBreakJ2    sql.NullInt32  `json:"tie_break_j2"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
} 
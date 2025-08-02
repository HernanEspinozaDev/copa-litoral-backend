package models

import (
	"database/sql"
	"time"
)

type Usuario struct {
	ID           int            `json:"id"`
	NombreUsuario string        `json:"nombre_usuario" validate:"required,min=3,max=50,no_sql_injection,safe_string"`
	Email        sql.NullString `json:"email"`
	Password     string         `json:"password,omitempty" validate:"required,min=6,max=100"`
	PasswordHash string         `json:"-"`
	Rol          string         `json:"rol" validate:"required,oneof=administrador jugador"`
	JugadorID    sql.NullInt32  `json:"jugador_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
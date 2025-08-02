package models

import "time"

type Categoria struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} 
package models

import (
	"database/sql"
	"time"
)

type Jugador struct {
	ID                    int            `json:"id"`
	Nombre                string         `json:"nombre"`
	Apellido              string         `json:"apellido"`
	TelefonoWSP           sql.NullString `json:"telefono_wsp"`
	ContactoVisibleEnWeb  bool           `json:"contacto_visible_en_web"`
	CategoriaID           sql.NullInt32  `json:"categoria_id"`
	Club                  sql.NullString `json:"club"`
	EstadoParticipacion   string         `json:"estado_participacion"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
} 
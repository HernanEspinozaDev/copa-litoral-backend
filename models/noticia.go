package models

import (
	"database/sql"
	"time"
)

type Noticia struct {
	ID                  int            `json:"id"`
	Titulo              string         `json:"titulo"`
	Slug                string         `json:"slug"`
	Contenido           string         `json:"contenido"`
	FechaPublicacion    time.Time      `json:"fecha_publicacion"`
	AutorID             sql.NullInt32  `json:"autor_id"`
	ImagenDestacadaURL  sql.NullString `json:"imagen_destacada_url"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
} 
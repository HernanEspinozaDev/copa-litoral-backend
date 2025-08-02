package models

import (
	"database/sql"
	"time"
)

type EstadoPartido string

const (
	EstadoAgendado    EstadoPartido = "agendado"
	EstadoEnJuego     EstadoPartido = "en_juego"
	EstadoFinalizado  EstadoPartido = "finalizado"
	EstadoCancelado   EstadoPartido = "cancelado"
)

type Partido struct {
	ID                  int            `json:"id"`
	TorneoID            int            `json:"torneo_id"`
	CategoriaID         int            `json:"categoria_id"`
	Jugador1ID          int            `json:"jugador1_id"`
	Jugador2ID          int            `json:"jugador2_id"`
	Fase                string         `json:"fase"`
	FechaAgendada       sql.NullTime   `json:"fecha_agendada"`
	HoraAgendada        sql.NullTime   `json:"hora_agendada"`
	PropuestaFechaJ1    sql.NullTime   `json:"propuesta_fecha_j1"`
	PropuestaHoraJ1     sql.NullTime   `json:"propuesta_hora_j1"`
	PropuestaFechaJ2    sql.NullTime   `json:"propuesta_fecha_j2"`
	PropuestaHoraJ2     sql.NullTime   `json:"propuesta_hora_j2"`
	PropuestaAceptadaJ1 bool           `json:"propuesta_aceptada_j1"`
	PropuestaAceptadaJ2 bool           `json:"propuesta_aceptada_j2"`
	Estado              EstadoPartido  `json:"estado"`
	ResultadoSetsJ1     sql.NullInt32  `json:"resultado_sets_j1"`
	ResultadoSetsJ2     sql.NullInt32  `json:"resultado_sets_j2"`
	GanadorID           sql.NullInt32  `json:"ganador_id"`
	PerdedorID          sql.NullInt32  `json:"perdedor_id"`
	ResultadoAprobado   bool           `json:"resultado_aprobado"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	
	// Campos adicionales para facilitar la respuesta JSON
	Jugador1Nombre      string         `json:"jugador1_nombre,omitempty"`
	Jugador2Nombre      string         `json:"jugador2_nombre,omitempty"`
	CategoriaNombre     string         `json:"categoria_nombre,omitempty"`
} 
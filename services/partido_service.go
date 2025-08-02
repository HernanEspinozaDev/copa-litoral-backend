package services

import (
	"database/sql"
	"errors"
	"time"

	"copa-litoral-backend/database"
	"copa-litoral-backend/models"
)

type PartidoService interface {
	GetAllPartidos(categoriaID int) ([]models.Partido, error)
	GetPartidoByID(id int) (*models.Partido, error)
	CreatePartido(partido *models.Partido) error
	UpdatePartido(id int, partido *models.Partido) error
	DeletePartido(id int) error
	ProposeMatchTime(partidoID int, jugadorID int, fecha time.Time, hora time.Time) error
	AcceptMatchTime(partidoID int, jugadorID int) error
	ReportMatchResult(partidoID int, setsGanadosJ1 int, setsGanadosJ2 int, ganadorID int, sets []models.SetPartido) error
	ApproveResult(partidoID int) error
}

type partidoServiceImpl struct{}

func NewPartidoService() PartidoService {
	return &partidoServiceImpl{}
}

func (s *partidoServiceImpl) GetAllPartidos(categoriaID int) ([]models.Partido, error) {
	var query string
	var args []interface{}

	if categoriaID > 0 {
		query = `
			SELECT p.id, p.torneo_id, p.categoria_id, p.jugador1_id, p.jugador2_id, 
			       p.fase, p.fecha_agendada, p.hora_agendada, p.propuesta_fecha_j1, 
			       p.propuesta_hora_j1, p.propuesta_fecha_j2, p.propuesta_hora_j2,
			       p.propuesta_aceptada_j1, p.propuesta_aceptada_j2, p.estado,
			       p.resultado_sets_j1, p.resultado_sets_j2, p.ganador_id, p.perdedor_id,
			       p.resultado_aprobado, p.created_at, p.updated_at,
			       j1.nombre || ' ' || j1.apellido as jugador1_nombre,
			       j2.nombre || ' ' || j2.apellido as jugador2_nombre,
			       c.nombre as categoria_nombre
			FROM partidos p
			LEFT JOIN jugadores j1 ON p.jugador1_id = j1.id
			LEFT JOIN jugadores j2 ON p.jugador2_id = j2.id
			LEFT JOIN categorias c ON p.categoria_id = c.id
			WHERE p.categoria_id = $1
			ORDER BY p.created_at DESC`
		args = append(args, categoriaID)
	} else {
		query = `
			SELECT p.id, p.torneo_id, p.categoria_id, p.jugador1_id, p.jugador2_id, 
			       p.fase, p.fecha_agendada, p.hora_agendada, p.propuesta_fecha_j1, 
			       p.propuesta_hora_j1, p.propuesta_fecha_j2, p.propuesta_hora_j2,
			       p.propuesta_aceptada_j1, p.propuesta_aceptada_j2, p.estado,
			       p.resultado_sets_j1, p.resultado_sets_j2, p.ganador_id, p.perdedor_id,
			       p.resultado_aprobado, p.created_at, p.updated_at,
			       j1.nombre || ' ' || j1.apellido as jugador1_nombre,
			       j2.nombre || ' ' || j2.apellido as jugador2_nombre,
			       c.nombre as categoria_nombre
			FROM partidos p
			LEFT JOIN jugadores j1 ON p.jugador1_id = j1.id
			LEFT JOIN jugadores j2 ON p.jugador2_id = j2.id
			LEFT JOIN categorias c ON p.categoria_id = c.id
			ORDER BY p.created_at DESC`
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var partidos []models.Partido
	for rows.Next() {
		var p models.Partido
		err := rows.Scan(
			&p.ID, &p.TorneoID, &p.CategoriaID, &p.Jugador1ID, &p.Jugador2ID,
			&p.Fase, &p.FechaAgendada, &p.HoraAgendada, &p.PropuestaFechaJ1,
			&p.PropuestaHoraJ1, &p.PropuestaFechaJ2, &p.PropuestaHoraJ2,
			&p.PropuestaAceptadaJ1, &p.PropuestaAceptadaJ2, &p.Estado,
			&p.ResultadoSetsJ1, &p.ResultadoSetsJ2, &p.GanadorID, &p.PerdedorID,
			&p.ResultadoAprobado, &p.CreatedAt, &p.UpdatedAt,
			&p.Jugador1Nombre, &p.Jugador2Nombre, &p.CategoriaNombre,
		)
		if err != nil {
			return nil, err
		}
		partidos = append(partidos, p)
	}

	return partidos, nil
}

func (s *partidoServiceImpl) GetPartidoByID(id int) (*models.Partido, error) {
	query := `
		SELECT p.id, p.torneo_id, p.categoria_id, p.jugador1_id, p.jugador2_id, 
		       p.fase, p.fecha_agendada, p.hora_agendada, p.propuesta_fecha_j1, 
		       p.propuesta_hora_j1, p.propuesta_fecha_j2, p.propuesta_hora_j2,
		       p.propuesta_aceptada_j1, p.propuesta_aceptada_j2, p.estado,
		       p.resultado_sets_j1, p.resultado_sets_j2, p.ganador_id, p.perdedor_id,
		       p.resultado_aprobado, p.created_at, p.updated_at,
		       j1.nombre || ' ' || j1.apellido as jugador1_nombre,
		       j2.nombre || ' ' || j2.apellido as jugador2_nombre,
		       c.nombre as categoria_nombre
		FROM partidos p
		LEFT JOIN jugadores j1 ON p.jugador1_id = j1.id
		LEFT JOIN jugadores j2 ON p.jugador2_id = j2.id
		LEFT JOIN categorias c ON p.categoria_id = c.id
		WHERE p.id = $1`

	var partido models.Partido
	err := database.DB.QueryRow(query, id).Scan(
		&partido.ID, &partido.TorneoID, &partido.CategoriaID, &partido.Jugador1ID, &partido.Jugador2ID,
		&partido.Fase, &partido.FechaAgendada, &partido.HoraAgendada, &partido.PropuestaFechaJ1,
		&partido.PropuestaHoraJ1, &partido.PropuestaFechaJ2, &partido.PropuestaHoraJ2,
		&partido.PropuestaAceptadaJ1, &partido.PropuestaAceptadaJ2, &partido.Estado,
		&partido.ResultadoSetsJ1, &partido.ResultadoSetsJ2, &partido.GanadorID, &partido.PerdedorID,
		&partido.ResultadoAprobado, &partido.CreatedAt, &partido.UpdatedAt,
		&partido.Jugador1Nombre, &partido.Jugador2Nombre, &partido.CategoriaNombre,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("partido no encontrado")
		}
		return nil, err
	}

	return &partido, nil
}

func (s *partidoServiceImpl) CreatePartido(partido *models.Partido) error {
	query := `
		INSERT INTO partidos (torneo_id, categoria_id, jugador1_id, jugador2_id, fase, 
		                     fecha_agendada, hora_agendada, estado, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return database.DB.QueryRow(query,
		partido.TorneoID, partido.CategoriaID, partido.Jugador1ID, partido.Jugador2ID,
		partido.Fase, partido.FechaAgendada, partido.HoraAgendada, partido.Estado,
	).Scan(&partido.ID, &partido.CreatedAt, &partido.UpdatedAt)
}

func (s *partidoServiceImpl) UpdatePartido(id int, partido *models.Partido) error {
	query := `
		UPDATE partidos 
		SET torneo_id = $1, categoria_id = $2, jugador1_id = $3, jugador2_id = $4,
		    fase = $5, fecha_agendada = $6, hora_agendada = $7, estado = $8,
		    updated_at = NOW()
		WHERE id = $9`

	result, err := database.DB.Exec(query,
		partido.TorneoID, partido.CategoriaID, partido.Jugador1ID, partido.Jugador2ID,
		partido.Fase, partido.FechaAgendada, partido.HoraAgendada, partido.Estado, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("partido no encontrado")
	}

	return nil
}

func (s *partidoServiceImpl) DeletePartido(id int) error {
	query := `DELETE FROM partidos WHERE id = $1`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("partido no encontrado")
	}

	return nil
}

func (s *partidoServiceImpl) ProposeMatchTime(partidoID int, jugadorID int, fecha time.Time, hora time.Time) error {
	// Verificar que el jugador es parte del partido
	partido, err := s.GetPartidoByID(partidoID)
	if err != nil {
		return err
	}

	if partido.Jugador1ID != jugadorID && partido.Jugador2ID != jugadorID {
		return errors.New("el jugador no es parte de este partido")
	}

	// Determinar qué jugador está proponiendo
	var query string
	if partido.Jugador1ID == jugadorID {
		query = `
			UPDATE partidos 
			SET propuesta_fecha_j1 = $1, propuesta_hora_j1 = $2, updated_at = NOW()
			WHERE id = $3`
	} else {
		query = `
			UPDATE partidos 
			SET propuesta_fecha_j2 = $1, propuesta_hora_j2 = $2, updated_at = NOW()
			WHERE id = $3`
	}

	result, err := database.DB.Exec(query, fecha, hora, partidoID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("partido no encontrado")
	}

	return nil
}

func (s *partidoServiceImpl) AcceptMatchTime(partidoID int, jugadorID int) error {
	// Verificar que el jugador es parte del partido
	partido, err := s.GetPartidoByID(partidoID)
	if err != nil {
		return err
	}

	if partido.Jugador1ID != jugadorID && partido.Jugador2ID != jugadorID {
		return errors.New("el jugador no es parte de este partido")
	}

	// Determinar qué jugador está aceptando
	var query string
	if partido.Jugador1ID == jugadorID {
		query = `
			UPDATE partidos 
			SET propuesta_aceptada_j1 = true, updated_at = NOW()
			WHERE id = $1`
	} else {
		query = `
			UPDATE partidos 
			SET propuesta_aceptada_j2 = true, updated_at = NOW()
			WHERE id = $1`
	}

	result, err := database.DB.Exec(query, partidoID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("partido no encontrado")
	}

	return nil
}

func (s *partidoServiceImpl) ReportMatchResult(partidoID int, setsGanadosJ1 int, setsGanadosJ2 int, ganadorID int, sets []models.SetPartido) error {
	// Iniciar transacción
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Actualizar el partido con el resultado
	perdedorID := 0
	if ganadorID == 0 {
		return errors.New("ganador_id es requerido")
	}

	// Determinar el perdedor
	partido, err := s.GetPartidoByID(partidoID)
	if err != nil {
		return err
	}

	if ganadorID == partido.Jugador1ID {
		perdedorID = partido.Jugador2ID
	} else if ganadorID == partido.Jugador2ID {
		perdedorID = partido.Jugador1ID
	} else {
		return errors.New("ganador_id no corresponde a ningún jugador del partido")
	}

	// Actualizar partido
	updateQuery := `
		UPDATE partidos 
		SET resultado_sets_j1 = $1, resultado_sets_j2 = $2, ganador_id = $3, 
		    perdedor_id = $4, estado = 'finalizado', updated_at = NOW()
		WHERE id = $5`

	_, err = tx.Exec(updateQuery, setsGanadosJ1, setsGanadosJ2, ganadorID, perdedorID, partidoID)
	if err != nil {
		return err
	}

	// Insertar los sets del partido
	for _, set := range sets {
		set.PartidoID = partidoID
		insertSetQuery := `
			INSERT INTO sets_partido (partido_id, numero_set, score_jugador1, score_jugador2,
			                         tie_break_j1, tie_break_j2, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

		_, err = tx.Exec(insertSetQuery,
			set.PartidoID, set.NumeroSet, set.ScoreJugador1, set.ScoreJugador2,
			set.TieBreakJ1, set.TieBreakJ2,
		)
		if err != nil {
			return err
		}
	}

	// Confirmar transacción
	return tx.Commit()
}

func (s *partidoServiceImpl) ApproveResult(partidoID int) error {
	query := `
		UPDATE partidos 
		SET resultado_aprobado = true, updated_at = NOW()
		WHERE id = $1`

	result, err := database.DB.Exec(query, partidoID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("partido no encontrado")
	}

	return nil
} 
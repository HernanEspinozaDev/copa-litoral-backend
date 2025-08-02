package services

import (
	"database/sql"
	"errors"

	"copa-litoral-backend/database"
	"copa-litoral-backend/models"
)

type JugadorService interface {
	GetAllJugadores() ([]models.Jugador, error)
	GetJugadorByID(id int) (*models.Jugador, error)
	CreateJugador(jugador *models.Jugador) error
	UpdateJugador(id int, jugador *models.Jugador) error
	DeleteJugador(id int) error
}

type jugadorServiceImpl struct{}

func NewJugadorService() JugadorService {
	return &jugadorServiceImpl{}
}

func (s *jugadorServiceImpl) GetAllJugadores() ([]models.Jugador, error) {
	query := `
		SELECT id, nombre, apellido, telefono_wsp, contacto_visible_en_web, 
		       categoria_id, club, estado_participacion, created_at, updated_at
		FROM jugadores
		ORDER BY nombre, apellido`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jugadores []models.Jugador
	for rows.Next() {
		var j models.Jugador
		err := rows.Scan(
			&j.ID, &j.Nombre, &j.Apellido, &j.TelefonoWSP, &j.ContactoVisibleEnWeb,
			&j.CategoriaID, &j.Club, &j.EstadoParticipacion, &j.CreatedAt, &j.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jugadores = append(jugadores, j)
	}

	return jugadores, nil
}

func (s *jugadorServiceImpl) GetJugadorByID(id int) (*models.Jugador, error) {
	query := `
		SELECT id, nombre, apellido, telefono_wsp, contacto_visible_en_web, 
		       categoria_id, club, estado_participacion, created_at, updated_at
		FROM jugadores
		WHERE id = $1`

	var jugador models.Jugador
	err := database.DB.QueryRow(query, id).Scan(
		&jugador.ID, &jugador.Nombre, &jugador.Apellido, &jugador.TelefonoWSP, &jugador.ContactoVisibleEnWeb,
		&jugador.CategoriaID, &jugador.Club, &jugador.EstadoParticipacion, &jugador.CreatedAt, &jugador.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("jugador no encontrado")
		}
		return nil, err
	}

	return &jugador, nil
}

func (s *jugadorServiceImpl) CreateJugador(jugador *models.Jugador) error {
	query := `
		INSERT INTO jugadores (nombre, apellido, telefono_wsp, contacto_visible_en_web, 
		                      categoria_id, club, estado_participacion, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return database.DB.QueryRow(query,
		jugador.Nombre, jugador.Apellido, jugador.TelefonoWSP, jugador.ContactoVisibleEnWeb,
		jugador.CategoriaID, jugador.Club, jugador.EstadoParticipacion,
	).Scan(&jugador.ID, &jugador.CreatedAt, &jugador.UpdatedAt)
}

func (s *jugadorServiceImpl) UpdateJugador(id int, jugador *models.Jugador) error {
	query := `
		UPDATE jugadores 
		SET nombre = $1, apellido = $2, telefono_wsp = $3, contacto_visible_en_web = $4,
		    categoria_id = $5, club = $6, estado_participacion = $7, updated_at = NOW()
		WHERE id = $8`

	result, err := database.DB.Exec(query,
		jugador.Nombre, jugador.Apellido, jugador.TelefonoWSP, jugador.ContactoVisibleEnWeb,
		jugador.CategoriaID, jugador.Club, jugador.EstadoParticipacion, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("jugador no encontrado")
	}

	return nil
}

func (s *jugadorServiceImpl) DeleteJugador(id int) error {
	query := `DELETE FROM jugadores WHERE id = $1`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("jugador no encontrado")
	}

	return nil
} 
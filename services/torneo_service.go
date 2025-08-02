package services

import (
	"database/sql"
	"errors"

	"copa-litoral-backend/database"
	"copa-litoral-backend/models"
)

type TorneoService interface {
	GetAllTorneos() ([]models.Torneo, error)
	GetTorneoByID(id int) (*models.Torneo, error)
	CreateTorneo(torneo *models.Torneo) error
	UpdateTorneo(id int, torneo *models.Torneo) error
	DeleteTorneo(id int) error
}

type torneoServiceImpl struct{}

func NewTorneoService() TorneoService {
	return &torneoServiceImpl{}
}

func (s *torneoServiceImpl) GetAllTorneos() ([]models.Torneo, error) {
	query := `
		SELECT id, nombre, anio, fecha_inicio, fecha_fin, foto_url, frase_destacada, 
		       activo, created_at, updated_at
		FROM torneos
		ORDER BY anio DESC, nombre`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var torneos []models.Torneo
	for rows.Next() {
		var t models.Torneo
		err := rows.Scan(
			&t.ID, &t.Nombre, &t.Anio, &t.FechaInicio, &t.FechaFin, &t.FotoURL,
			&t.FraseDestacada, &t.Activo, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		torneos = append(torneos, t)
	}

	return torneos, nil
}

func (s *torneoServiceImpl) GetTorneoByID(id int) (*models.Torneo, error) {
	query := `
		SELECT id, nombre, anio, fecha_inicio, fecha_fin, foto_url, frase_destacada, 
		       activo, created_at, updated_at
		FROM torneos
		WHERE id = $1`

	var torneo models.Torneo
	err := database.DB.QueryRow(query, id).Scan(
		&torneo.ID, &torneo.Nombre, &torneo.Anio, &torneo.FechaInicio, &torneo.FechaFin,
		&torneo.FotoURL, &torneo.FraseDestacada, &torneo.Activo, &torneo.CreatedAt, &torneo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("torneo no encontrado")
		}
		return nil, err
	}

	return &torneo, nil
}

func (s *torneoServiceImpl) CreateTorneo(torneo *models.Torneo) error {
	query := `
		INSERT INTO torneos (nombre, anio, fecha_inicio, fecha_fin, foto_url, 
		                    frase_destacada, activo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return database.DB.QueryRow(query,
		torneo.Nombre, torneo.Anio, torneo.FechaInicio, torneo.FechaFin,
		torneo.FotoURL, torneo.FraseDestacada, torneo.Activo,
	).Scan(&torneo.ID, &torneo.CreatedAt, &torneo.UpdatedAt)
}

func (s *torneoServiceImpl) UpdateTorneo(id int, torneo *models.Torneo) error {
	query := `
		UPDATE torneos 
		SET nombre = $1, anio = $2, fecha_inicio = $3, fecha_fin = $4,
		    foto_url = $5, frase_destacada = $6, activo = $7, updated_at = NOW()
		WHERE id = $8`

	result, err := database.DB.Exec(query,
		torneo.Nombre, torneo.Anio, torneo.FechaInicio, torneo.FechaFin,
		torneo.FotoURL, torneo.FraseDestacada, torneo.Activo, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("torneo no encontrado")
	}

	return nil
}

func (s *torneoServiceImpl) DeleteTorneo(id int) error {
	query := `DELETE FROM torneos WHERE id = $1`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("torneo no encontrado")
	}

	return nil
} 
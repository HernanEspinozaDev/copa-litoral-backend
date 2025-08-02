package services

import (
	"database/sql"
	"errors"

	"copa-litoral-backend/database"
	"copa-litoral-backend/models"
)

type CategoriaService interface {
	GetAllCategorias() ([]models.Categoria, error)
	GetCategoriaByID(id int) (*models.Categoria, error)
	CreateCategoria(categoria *models.Categoria) error
	UpdateCategoria(id int, categoria *models.Categoria) error
	DeleteCategoria(id int) error
}

type categoriaServiceImpl struct{}

func NewCategoriaService() CategoriaService {
	return &categoriaServiceImpl{}
}

func (s *categoriaServiceImpl) GetAllCategorias() ([]models.Categoria, error) {
	query := `
		SELECT id, nombre, created_at, updated_at
		FROM categorias
		ORDER BY nombre`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categorias []models.Categoria
	for rows.Next() {
		var c models.Categoria
		err := rows.Scan(&c.ID, &c.Nombre, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categorias = append(categorias, c)
	}

	return categorias, nil
}

func (s *categoriaServiceImpl) GetCategoriaByID(id int) (*models.Categoria, error) {
	query := `
		SELECT id, nombre, created_at, updated_at
		FROM categorias
		WHERE id = $1`

	var categoria models.Categoria
	err := database.DB.QueryRow(query, id).Scan(
		&categoria.ID, &categoria.Nombre, &categoria.CreatedAt, &categoria.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("categoría no encontrada")
		}
		return nil, err
	}

	return &categoria, nil
}

func (s *categoriaServiceImpl) CreateCategoria(categoria *models.Categoria) error {
	query := `
		INSERT INTO categorias (nombre, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return database.DB.QueryRow(query, categoria.Nombre).Scan(
		&categoria.ID, &categoria.CreatedAt, &categoria.UpdatedAt,
	)
}

func (s *categoriaServiceImpl) UpdateCategoria(id int, categoria *models.Categoria) error {
	query := `
		UPDATE categorias 
		SET nombre = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := database.DB.Exec(query, categoria.Nombre, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("categoría no encontrada")
	}

	return nil
}

func (s *categoriaServiceImpl) DeleteCategoria(id int) error {
	query := `DELETE FROM categorias WHERE id = $1`

	result, err := database.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("categoría no encontrada")
	}

	return nil
} 
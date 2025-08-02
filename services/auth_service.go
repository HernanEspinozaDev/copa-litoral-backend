package services

import (
	"database/sql"
	"errors"

	"copa-litoral-backend/config"
	"copa-litoral-backend/database"
	"copa-litoral-backend/models"
	"copa-litoral-backend/utils"
)

type AuthService interface {
	RegisterUser(user *models.Usuario) error
	LoginUser(username, password string) (string, error)
}

type authServiceImpl struct{
	config *config.Config
}

func NewAuthService(cfg *config.Config) AuthService {
	return &authServiceImpl{
		config: cfg,
	}
}

func (s *authServiceImpl) RegisterUser(user *models.Usuario) error {
	// Hashear la contraseña antes de guardar
	hashedPassword, err := utils.HashPassword(user.PasswordHash)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword

	// Verificar que el nombre de usuario no exista
	var existingID int
	err = database.DB.QueryRow("SELECT id FROM usuarios WHERE nombre_usuario = $1", user.NombreUsuario).Scan(&existingID)
	if err == nil {
		return errors.New("el nombre de usuario ya existe")
	} else if err != sql.ErrNoRows {
		return err
	}

	// Insertar el nuevo usuario
	query := `
		INSERT INTO usuarios (nombre_usuario, email, password_hash, rol, jugador_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return database.DB.QueryRow(query,
		user.NombreUsuario, user.Email, user.PasswordHash, user.Rol, user.JugadorID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (s *authServiceImpl) LoginUser(username, password string) (string, error) {
	// Buscar el usuario por nombre de usuario
	var user models.Usuario
	query := `
		SELECT id, nombre_usuario, email, password_hash, rol, jugador_id, created_at, updated_at
		FROM usuarios
		WHERE nombre_usuario = $1`

	err := database.DB.QueryRow(query, username).Scan(
		&user.ID, &user.NombreUsuario, &user.Email, &user.PasswordHash, &user.Rol, &user.JugadorID,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("credenciales inválidas")
		}
		return "", err
	}

	// Verificar la contraseña
	err = utils.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	// Generar JWT
	token, err := utils.GenerateJWT(user.ID, user.Rol, s.config.JWTSecret)
	if err != nil {
		return "", err
	}

	return token, nil
} 
package handlers

import (
	"net/http"

	"copa-litoral-backend/config"
	"copa-litoral-backend/models"
	"copa-litoral-backend/services"
	"copa-litoral-backend/utils"
)

type AuthHandler struct {
	authService services.AuthService
	config      *config.Config
}

func NewAuthHandler(authService services.AuthService, config *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		config:      config,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parsear y validar JSON
	var usuario models.Usuario
	if err := utils.ParseAndValidateJSON(r, &usuario); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos inválidos: "+err.Error())
		return
	}

	// Sanitizar inputs
	usuario.NombreUsuario = utils.SanitizeString(usuario.NombreUsuario)
	
	// Validar que el password no esté vacío (viene del JSON)
	if usuario.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Password es requerido")
		return
	}

	// Mover password a PasswordHash para el servicio
	usuario.PasswordHash = usuario.Password
	usuario.Password = "" // Limpiar password del struct
	
	// Establecer rol por defecto si no se especifica
	if usuario.Rol == "" {
		usuario.Rol = "jugador"
	}

	// Registrar usuario
	err := h.authService.RegisterUser(&usuario)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Usuario creado exitosamente"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NombreUsuario string `json:"nombre_usuario" validate:"required,min=3,max=50,no_sql_injection,safe_string"`
		Password      string `json:"password" validate:"required,min=6,max=100"`
	}

	// Parsear y validar JSON
	if err := utils.ParseAndValidateJSON(r, &request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos inválidos: "+err.Error())
		return
	}

	// Sanitizar inputs
	request.NombreUsuario = utils.SanitizeString(request.NombreUsuario)
	request.Password = utils.SanitizeString(request.Password)

	token, err := h.authService.LoginUser(request.NombreUsuario, request.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	response := map[string]string{
		"token": token,
		"message": "Login exitoso",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
} 
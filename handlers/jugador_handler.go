package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"copa-litoral-backend/models"
	"copa-litoral-backend/services"
	"copa-litoral-backend/utils"

	"github.com/gorilla/mux"
)

type JugadorHandler struct {
	jugadorService services.JugadorService
}

func NewJugadorHandler(jugadorService services.JugadorService) *JugadorHandler {
	return &JugadorHandler{
		jugadorService: jugadorService,
	}
}

func (h *JugadorHandler) GetJugadores(w http.ResponseWriter, r *http.Request) {
	jugadores, err := h.jugadorService.GetAllJugadores()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, jugadores)
}

func (h *JugadorHandler) GetJugador(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	jugador, err := h.jugadorService.GetJugadorByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, jugador)
}

func (h *JugadorHandler) CreateJugador(w http.ResponseWriter, r *http.Request) {
	var jugador models.Jugador
	if err := json.NewDecoder(r.Body).Decode(&jugador); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.jugadorService.CreateJugador(&jugador); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, jugador)
}

func (h *JugadorHandler) UpdateJugador(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var jugador models.Jugador
	if err := json.NewDecoder(r.Body).Decode(&jugador); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.jugadorService.UpdateJugador(id, &jugador); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Jugador actualizado exitosamente"})
}

func (h *JugadorHandler) DeleteJugador(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.jugadorService.DeleteJugador(id); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Jugador eliminado exitosamente"})
} 
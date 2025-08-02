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

type TorneoHandler struct {
	torneoService services.TorneoService
}

func NewTorneoHandler(torneoService services.TorneoService) *TorneoHandler {
	return &TorneoHandler{
		torneoService: torneoService,
	}
}

func (h *TorneoHandler) GetTorneos(w http.ResponseWriter, r *http.Request) {
	torneos, err := h.torneoService.GetAllTorneos()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, torneos)
}

func (h *TorneoHandler) GetTorneo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	torneo, err := h.torneoService.GetTorneoByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, torneo)
}

func (h *TorneoHandler) CreateTorneo(w http.ResponseWriter, r *http.Request) {
	var torneo models.Torneo
	if err := json.NewDecoder(r.Body).Decode(&torneo); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.torneoService.CreateTorneo(&torneo); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, torneo)
}

func (h *TorneoHandler) UpdateTorneo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var torneo models.Torneo
	if err := json.NewDecoder(r.Body).Decode(&torneo); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.torneoService.UpdateTorneo(id, &torneo); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Torneo actualizado exitosamente"})
}

func (h *TorneoHandler) DeleteTorneo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.torneoService.DeleteTorneo(id); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Torneo eliminado exitosamente"})
} 
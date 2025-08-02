package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"copa-litoral-backend/models"
	"copa-litoral-backend/services"
	"copa-litoral-backend/utils"

	"github.com/gorilla/mux"
)

type PartidoHandler struct {
	partidoService services.PartidoService
}

func NewPartidoHandler(partidoService services.PartidoService) *PartidoHandler {
	return &PartidoHandler{
		partidoService: partidoService,
	}
}

func (h *PartidoHandler) GetPartidos(w http.ResponseWriter, r *http.Request) {
	categoriaIDStr := r.URL.Query().Get("categoria_id")
	categoriaID := 0

	if categoriaIDStr != "" {
		var err error
		categoriaID, err = strconv.Atoi(categoriaIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "categoria_id inválido")
			return
		}
	}

	partidos, err := h.partidoService.GetAllPartidos(categoriaID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, partidos)
}

func (h *PartidoHandler) GetPartido(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	partido, err := h.partidoService.GetPartidoByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, partido)
}

func (h *PartidoHandler) CreatePartido(w http.ResponseWriter, r *http.Request) {
	var partido models.Partido
	if err := json.NewDecoder(r.Body).Decode(&partido); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.partidoService.CreatePartido(&partido); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, partido)
}

func (h *PartidoHandler) UpdatePartido(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var partido models.Partido
	if err := json.NewDecoder(r.Body).Decode(&partido); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.partidoService.UpdatePartido(id, &partido); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Partido actualizado exitosamente"})
}

func (h *PartidoHandler) DeletePartido(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.partidoService.DeletePartido(id); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Partido eliminado exitosamente"})
}

func (h *PartidoHandler) ProposeMatchTime(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partidoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID de partido inválido")
		return
	}

	// Obtener el jugador ID del contexto JWT (esto se implementará en el middleware)
	jugadorID := 1 // Placeholder - se obtendrá del contexto JWT

	var request struct {
		Fecha string `json:"fecha"`
		Hora  string `json:"hora"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	fecha, err := time.Parse("2006-01-02", request.Fecha)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Formato de fecha inválido (YYYY-MM-DD)")
		return
	}

	hora, err := time.Parse("15:04", request.Hora)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Formato de hora inválido (HH:MM)")
		return
	}

	if err := h.partidoService.ProposeMatchTime(partidoID, jugadorID, fecha, hora); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Propuesta de horario enviada exitosamente"})
}

func (h *PartidoHandler) AcceptMatchTime(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partidoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID de partido inválido")
		return
	}

	// Obtener el jugador ID del contexto JWT
	jugadorID := 1 // Placeholder - se obtendrá del contexto JWT

	if err := h.partidoService.AcceptMatchTime(partidoID, jugadorID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Horario aceptado exitosamente"})
}

func (h *PartidoHandler) ReportMatchResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partidoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID de partido inválido")
		return
	}

	var request struct {
		SetsGanadosJ1 int              `json:"sets_ganados_j1"`
		SetsGanadosJ2 int              `json:"sets_ganados_j2"`
		GanadorID     int              `json:"ganador_id"`
		Sets          []models.SetPartido `json:"sets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.partidoService.ReportMatchResult(partidoID, request.SetsGanadosJ1, request.SetsGanadosJ2, request.GanadorID, request.Sets); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Resultado reportado exitosamente"})
}

func (h *PartidoHandler) ApproveResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partidoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID de partido inválido")
		return
	}

	if err := h.partidoService.ApproveResult(partidoID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Resultado aprobado exitosamente"})
} 
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

type CategoriaHandler struct {
	categoriaService services.CategoriaService
}

func NewCategoriaHandler(categoriaService services.CategoriaService) *CategoriaHandler {
	return &CategoriaHandler{
		categoriaService: categoriaService,
	}
}

func (h *CategoriaHandler) GetCategorias(w http.ResponseWriter, r *http.Request) {
	categorias, err := h.categoriaService.GetAllCategorias()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, categorias)
}

func (h *CategoriaHandler) GetCategoria(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	categoria, err := h.categoriaService.GetCategoriaByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, categoria)
}

func (h *CategoriaHandler) CreateCategoria(w http.ResponseWriter, r *http.Request) {
	var categoria models.Categoria
	if err := json.NewDecoder(r.Body).Decode(&categoria); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.categoriaService.CreateCategoria(&categoria); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, categoria)
}

func (h *CategoriaHandler) UpdateCategoria(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var categoria models.Categoria
	if err := json.NewDecoder(r.Body).Decode(&categoria); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Datos JSON inválidos")
		return
	}

	if err := h.categoriaService.UpdateCategoria(id, &categoria); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Categoría actualizada exitosamente"})
}

func (h *CategoriaHandler) DeleteCategoria(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.categoriaService.DeleteCategoria(id); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Categoría eliminada exitosamente"})
} 
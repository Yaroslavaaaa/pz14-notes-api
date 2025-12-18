package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/notes-api/internal/core"
	"example.com/notes-api/internal/repo"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Repo *repo.NoteRepoPG
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

/*
====================
CREATE NOTE
====================
*/

// CreateNote godoc
// @Summary      Создать заметку
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        input  body     core.NoteCreate  true  "Данные новой заметки"
// @Success      201    {object} core.Note
// @Failure      400    {object} map[string]string
// @Failure      500    {object} map[string]string
// @Router       /notes [post]
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var req core.NoteCreate

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	id, err := h.Repo.Create(r.Context(), req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	note, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve created note")
		return
	}

	respondWithJSON(w, http.StatusCreated, note)
}

/*
====================
GET NOTE BY ID
====================
*/

// GetNote godoc
// @Summary      Получить заметку
// @Tags         notes
// @Param        id   path   int  true  "ID"
// @Success      200  {object} core.Note
// @Failure      400  {object} map[string]string
// @Failure      500  {object} map[string]string
// @Router       /notes/{id} [get]
func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	note, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get note")
		return
	}

	respondWithJSON(w, http.StatusOK, note)
}

/*
====================
LIST NOTES
====================
*/

// ListNotes godoc
// @Summary      Список заметок
// @Tags         notes
// @Success      200  {array} core.Note
// @Router       /notes [get]
func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.Repo.GetAll(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list notes")
		return
	}
	respondWithJSON(w, http.StatusOK, notes)
}

/*
====================
PATCH NOTE
====================
*/

// PatchNote godoc
// @Summary      Обновить заметку (частично)
// @Tags         notes
// @Accept       json
// @Param        id     path   int              true  "ID"
// @Param        input  body   core.NoteUpdate  true  "Поля для обновления"
// @Success      200    {object} core.Note
// @Failure      400    {object} map[string]string
// @Failure      500    {object} map[string]string
// @Router       /notes/{id} [patch]
func (h *Handler) PatchNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	var update core.NoteUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if update.Title == nil && update.Content == nil {
		respondWithError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	if update.Title != nil && strings.TrimSpace(*update.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Title cannot be empty")
		return
	}

	if err := h.Repo.Update(r.Context(), id, update); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update note")
		return
	}

	note, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve updated note")
		return
	}

	respondWithJSON(w, http.StatusOK, note)
}

/*
====================
DELETE NOTE
====================
*/

// DeleteNote godoc
// @Summary      Удалить заметку
// @Tags         notes
// @Param        id  path  int  true  "ID"
// @Success      204  "No Content"
// @Failure      400  {object} map[string]string
// @Failure      500  {object} map[string]string
// @Router       /notes/{id} [delete]
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/*
====================
HELPERS
====================
*/

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}

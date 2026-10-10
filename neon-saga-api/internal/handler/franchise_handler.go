package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
	"github.com/Lix-09991/neon-saga-api/internal/service"
)

type FranchiseHandler struct {
	svc      *service.FranchiseService
	validate *validator.Validate
}

func NewFranchiseHandler(svc *service.FranchiseService) *FranchiseHandler {
	return &FranchiseHandler{svc: svc, validate: validator.New()}
}

type franchiseRequest struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"       validate:"required,min=1,max=255"`
	Description *string `json:"description"`
	Cover       *string `json:"cover"`
}

type franchiseItemsRequest struct {
	Items []franchiseItem `json:"items" validate:"dive"`
}

type franchiseItem struct {
	AnimeID  uuid.UUID `json:"anime_id" validate:"required"`
	Position int       `json:"position" validate:"required,gte=1"`
	Note     *string   `json:"note"`
}

func (h *FranchiseHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, total, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка списка")
		return
	}
	if items == nil {
		items = []models.Franchise{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *FranchiseHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	f, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrFranchiseNotFound) {
			writeError(w, http.StatusNotFound, "не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка")
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *FranchiseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req franchiseRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}
	f, err := h.svc.Create(r.Context(), service.CreateFranchiseInput{
		Slug: req.Slug, Title: req.Title, Description: req.Description, Cover: req.Cover,
	})
	if err != nil {
		if errors.Is(err, repository.ErrFranchiseSlugTaken) {
			writeError(w, http.StatusConflict, "slug занят")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка создания")
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (h *FranchiseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	var req franchiseRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}
	f, err := h.svc.Update(r.Context(), id, service.CreateFranchiseInput{
		Title: req.Title, Description: req.Description, Cover: req.Cover,
	})
	if err != nil {
		if errors.Is(err, service.ErrFranchiseNotFound) {
			writeError(w, http.StatusNotFound, "не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка обновления")
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *FranchiseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка удаления")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FranchiseHandler) SetItems(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	var req franchiseItemsRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}
	items := make([]models.FranchiseItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, models.FranchiseItem{
			AnimeID: it.AnimeID, Position: it.Position, Note: it.Note,
		})
	}
	if err := h.svc.ReplaceItems(r.Context(), id, items); err != nil {
		if errors.Is(err, service.ErrFranchiseNotFound) {
			writeError(w, http.StatusNotFound, "не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

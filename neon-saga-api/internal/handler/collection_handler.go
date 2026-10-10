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

type CollectionHandler struct {
	svc      *service.CollectionService
	validate *validator.Validate
}

func NewCollectionHandler(svc *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{svc: svc, validate: validator.New()}
}

type collectionRequest struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"       validate:"required,min=1,max=255"`
	Description *string `json:"description"`
	Cover       *string `json:"cover"`
	IsFeatured  bool    `json:"is_featured"`
	Position    int     `json:"position"`
}

type collectionItemsRequest struct {
	Items []collectionItem `json:"items" validate:"dive"`
}

type collectionItem struct {
	AnimeID  uuid.UUID `json:"anime_id" validate:"required"`
	Position int       `json:"position" validate:"required,gte=1"`
	Note     *string   `json:"note"`
}

func (h *CollectionHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var featured *bool
	if v := r.URL.Query().Get("featured"); v == "true" {
		b := true
		featured = &b
	} else if v == "false" {
		b := false
		featured = &b
	}

	items, total, err := h.svc.List(r.Context(), repository.CollectionFilter{
		Featured: featured, Limit: limit, Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка списка")
		return
	}
	if items == nil {
		items = []models.Collection{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *CollectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req collectionRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}
	c, err := h.svc.Create(r.Context(), service.CreateCollectionInput{
		Slug: req.Slug, Title: req.Title, Description: req.Description,
		Cover: req.Cover, IsFeatured: req.IsFeatured, Position: req.Position,
	})
	if err != nil {
		if errors.Is(err, repository.ErrCollectionSlugTaken) {
			writeError(w, http.StatusConflict, "slug занят")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка создания")
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *CollectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	var req collectionRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}
	c, err := h.svc.Update(r.Context(), id, service.CreateCollectionInput{
		Title: req.Title, Description: req.Description, Cover: req.Cover,
		IsFeatured: req.IsFeatured, Position: req.Position,
	})
	if err != nil {
		if errors.Is(err, service.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка обновления")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func (h *CollectionHandler) SetItems(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	var req collectionItemsRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}
	items := make([]models.CollectionItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, models.CollectionItem{
			AnimeID: it.AnimeID, Position: it.Position, Note: it.Note,
		})
	}
	if err := h.svc.ReplaceItems(r.Context(), id, items); err != nil {
		if errors.Is(err, service.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

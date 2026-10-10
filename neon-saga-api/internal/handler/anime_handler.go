package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
	"github.com/Lix-09991/neon-saga-api/internal/service"
)

type AnimeHandler struct {
	animeService *service.AnimeService
	validate     *validator.Validate
}

func NewAnimeHandler(s *service.AnimeService) *AnimeHandler {
	return &AnimeHandler{animeService: s, validate: validator.New()}
}

// ============================================================
// DTO
// ============================================================
type animeRequest struct {
	Title       string     `json:"title"       validate:"required,min=1,max=255"`
	TitleEn     *string    `json:"title_en"`
	Description string     `json:"description" validate:"required"`
	Cover       *string    `json:"cover"`
	Banner      *string    `json:"banner"`
	Year        int        `json:"year"        validate:"required,gte=1900,lte=2100"`
	Country     string     `json:"country"`
	Type        string     `json:"type"`
	Status      string     `json:"status"      validate:"omitempty,oneof=ONGOING COMPLETED ANNOUNCED"`
	Genres      []string   `json:"genres"`
	FranchiseID *uuid.UUID `json:"franchise_id"`
}

// ============================================================
// HANDLERS
// ============================================================

// GET /api/v1/anime
func (h *AnimeHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 24
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset < 0 {
		offset = 0
	}
	yearFrom, _ := strconv.Atoi(q.Get("year_from"))
	yearTo, _ := strconv.Atoi(q.Get("year_to"))
	minRating, _ := strconv.ParseFloat(q.Get("min_rating"), 64)

	var franchiseID *uuid.UUID
	if v := q.Get("franchise_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			franchiseID = &id
		}
	}

	filter := repository.ListFilter{
		Search:      q.Get("search"),
		Genres:      splitComma(q.Get("genres")),
		GenresMode:  q.Get("genres_mode"),
		Countries:   splitComma(q.Get("countries")),
		Types:       splitComma(q.Get("types")),
		Statuses:    splitComma(q.Get("statuses")),
		YearFrom:    yearFrom,
		YearTo:      yearTo,
		MinRating:   minRating,
		FranchiseID: franchiseID,
		Sort:        q.Get("sort"),
		Order:       q.Get("order"),
		Limit:       limit,
		Offset:      offset,
	}

	items, total, err := h.animeService.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения списка")
		return
	}
	if items == nil {
		items = []models.Anime{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GET /api/v1/anime/facets
func (h *AnimeHandler) Facets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	yearFrom, _ := strconv.Atoi(q.Get("year_from"))
	yearTo, _ := strconv.Atoi(q.Get("year_to"))
	minRating, _ := strconv.ParseFloat(q.Get("min_rating"), 64)

	filter := repository.ListFilter{
		Search:     q.Get("search"),
		Genres:     splitComma(q.Get("genres")),
		GenresMode: q.Get("genres_mode"),
		Countries:  splitComma(q.Get("countries")),
		Types:      splitComma(q.Get("types")),
		Statuses:   splitComma(q.Get("statuses")),
		YearFrom:   yearFrom,
		YearTo:     yearTo,
		MinRating:  minRating,
	}

	facets, err := h.animeService.Facets(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка facets")
		return
	}
	writeJSON(w, http.StatusOK, facets)
}

// GET /api/v1/anime/{id}
func (h *AnimeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}

	anime, episodes, err := h.animeService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrAnimeNotFound) {
			writeError(w, http.StatusNotFound, "аниме не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка получения аниме")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"anime":    anime,
		"episodes": episodes,
	})
}

// POST /api/v1/anime (admin)
func (h *AnimeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req animeRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}

	a, err := h.animeService.Create(r.Context(), service.CreateAnimeInput{
		Title:       req.Title,
		TitleEn:     req.TitleEn,
		Description: req.Description,
		Cover:       req.Cover,
		Banner:      req.Banner,
		Year:        req.Year,
		Country:     req.Country,
		Type:        models.AnimeType(req.Type),
		Status:      models.AnimeStatus(req.Status),
		Genres:      req.Genres,
		FranchiseID: req.FranchiseID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrSlugTaken) {
			writeError(w, http.StatusConflict, "slug уже занят")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка создания аниме")
		return
	}

	writeJSON(w, http.StatusCreated, a)
}

// PATCH /api/v1/anime/{id} (admin)
func (h *AnimeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}

	var req animeRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}

	a, err := h.animeService.Update(r.Context(), id, service.UpdateAnimeInput{
		Title:       req.Title,
		TitleEn:     req.TitleEn,
		Description: req.Description,
		Cover:       req.Cover,
		Banner:      req.Banner,
		Year:        req.Year,
		Country:     req.Country,
		Type:        models.AnimeType(req.Type),
		Status:      models.AnimeStatus(req.Status),
		Genres:      req.Genres,
		FranchiseID: req.FranchiseID,
	})
	if err != nil {
		if errors.Is(err, service.ErrAnimeNotFound) {
			writeError(w, http.StatusNotFound, "аниме не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка обновления")
		return
	}

	writeJSON(w, http.StatusOK, a)
}

// DELETE /api/v1/anime/{id} (admin)
func (h *AnimeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}

	if err := h.animeService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrAnimeNotFound) {
			writeError(w, http.StatusNotFound, "аниме не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка удаления")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// HELPERS
// ============================================================
func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

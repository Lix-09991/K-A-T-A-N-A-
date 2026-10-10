package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/service"
)

type EpisodeHandler struct {
	epService *service.EpisodeService
	validate  *validator.Validate
}

func NewEpisodeHandler(s *service.EpisodeService) *EpisodeHandler {
	return &EpisodeHandler{epService: s, validate: validator.New()}
}

// ============================================================
// DTO
// ============================================================
type episodeRequest struct {
	Number      int     `json:"number"      validate:"required,gte=1"`
	Title       string  `json:"title"       validate:"required,min=1,max=255"`
	Description *string `json:"description"`
	Duration    int     `json:"duration"    validate:"gte=0"`
	VideoURL    string  `json:"video_url"   validate:"required"`
	Thumbnail   *string `json:"thumbnail"`
}

type markersRequest struct {
	Markers []markerItem `json:"markers" validate:"dive"`
}

type markerItem struct {
	Type      string  `json:"type"       validate:"required,oneof=op ed skip custom"`
	StartTime int     `json:"start_time" validate:"gte=0"`
	EndTime   *int    `json:"end_time"`
	Label     *string `json:"label"`
}

// ============================================================
// HANDLERS
// ============================================================

// GET /api/v1/anime/{animeID}/episodes
func (h *EpisodeHandler) ListByAnime(w http.ResponseWriter, r *http.Request) {
	animeID, err := uuid.Parse(chi.URLParam(r, "animeID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный animeID")
		return
	}

	eps, err := h.epService.ListByAnime(r.Context(), animeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения эпизодов")
		return
	}
	if eps == nil {
		eps = []models.Episode{}
	}
	writeJSON(w, http.StatusOK, eps)
}

// GET /api/v1/episodes/{id}
func (h *EpisodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}

	ep, markers, err := h.epService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrEpisodeNotFound) {
			writeError(w, http.StatusNotFound, "эпизод не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка получения эпизода")
		return
	}
	if markers == nil {
		markers = []models.EpisodeMarker{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"episode": ep,
		"markers": markers,
	})
}

// POST /api/v1/anime/{animeID}/episodes (admin)
func (h *EpisodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	animeID, err := uuid.Parse(chi.URLParam(r, "animeID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный animeID")
		return
	}

	var req episodeRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}

	ep, err := h.epService.Create(r.Context(), service.CreateEpisodeInput{
		AnimeID:     animeID,
		Number:      req.Number,
		Title:       req.Title,
		Description: req.Description,
		Duration:    req.Duration,
		VideoURL:    req.VideoURL,
		Thumbnail:   req.Thumbnail,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка создания эпизода")
		return
	}

	writeJSON(w, http.StatusCreated, ep)
}

// DELETE /api/v1/episodes/{id} (admin)
func (h *EpisodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}
	if err := h.epService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrEpisodeNotFound) {
			writeError(w, http.StatusNotFound, "эпизод не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка удаления")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PUT /api/v1/episodes/{id}/markers (admin)
func (h *EpisodeHandler) SetMarkers(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}

	var req markersRequest
	if !decodeAndValidate(w, r, h.validate, &req) {
		return
	}

	markers := make([]models.EpisodeMarker, 0, len(req.Markers))
	for _, m := range req.Markers {
		markers = append(markers, models.EpisodeMarker{
			Type:      models.MarkerType(m.Type),
			StartTime: m.StartTime,
			EndTime:   m.EndTime,
			Label:     m.Label,
		})
	}

	if err := h.epService.SetMarkers(r.Context(), id, markers); err != nil {
		if errors.Is(err, service.ErrEpisodeNotFound) {
			writeError(w, http.StatusNotFound, "эпизод не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка сохранения маркеров")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"markers": markers})
}

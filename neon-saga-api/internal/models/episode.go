package models

import (
	"time"

	"github.com/google/uuid"
)

type Episode struct {
	ID          uuid.UUID  `json:"id"`
	AnimeID     uuid.UUID  `json:"anime_id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Duration    int        `json:"duration"`
	VideoURL    string     `json:"video_url"`
	Thumbnail   *string    `json:"thumbnail,omitempty"`
	AiredAt     *time.Time `json:"aired_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ============================================================
// MARKERS — таймкоды OP/ED для автопропуска
// ============================================================
type MarkerType string

const (
	MarkerOP     MarkerType = "op"
	MarkerED     MarkerType = "ed"
	MarkerSkip   MarkerType = "skip"
	MarkerCustom MarkerType = "custom"
)

type EpisodeMarker struct {
	ID        uuid.UUID  `json:"id"`
	EpisodeID uuid.UUID  `json:"episode_id"`
	Type      MarkerType `json:"type"`
	StartTime int        `json:"start_time"`
	EndTime   *int       `json:"end_time,omitempty"`
	Label     *string    `json:"label,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

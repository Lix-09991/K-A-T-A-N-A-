package models

import (
	"time"

	"github.com/google/uuid"
)

type AnimeStatus string

const (
	StatusOngoing   AnimeStatus = "ONGOING"
	StatusCompleted AnimeStatus = "COMPLETED"
	StatusAnnounced AnimeStatus = "ANNOUNCED"
)

type Anime struct {
	ID          uuid.UUID   `json:"id"`
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	TitleEn     *string     `json:"title_en,omitempty"`
	Description string      `json:"description"`
	Cover       *string     `json:"cover,omitempty"`
	Banner      *string     `json:"banner,omitempty"`
	Year        int         `json:"year"`
	Rating      float32     `json:"rating"`
	Status      AnimeStatus `json:"status"`
	Genres      []string    `json:"genres"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

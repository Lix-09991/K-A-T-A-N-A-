package models

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Cover       *string   `json:"cover,omitempty"`
	IsFeatured  bool      `json:"is_featured"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Заполняется при загрузке
	Items      []CollectionItem `json:"items,omitempty"`
	AnimeCount int              `json:"anime_count,omitempty"`
}

type CollectionItem struct {
	ID           uuid.UUID `json:"id"`
	CollectionID uuid.UUID `json:"collection_id"`
	AnimeID      uuid.UUID `json:"anime_id"`
	Position     int       `json:"position"`
	Note         *string   `json:"note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// Заполняется при JOIN
	Anime *Anime `json:"anime,omitempty"`
}

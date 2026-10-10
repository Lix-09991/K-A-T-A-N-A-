package models

import (
	"time"

	"github.com/google/uuid"
)

type Franchise struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Cover       *string   `json:"cover,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Заполняется при загрузке с items
	Items []FranchiseItem `json:"items,omitempty"`
}

type FranchiseItem struct {
	ID          uuid.UUID `json:"id"`
	FranchiseID uuid.UUID `json:"franchise_id"`
	AnimeID     uuid.UUID `json:"anime_id"`
	Position    int       `json:"position"`
	Note        *string   `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	// Заполняется при JOIN
	Anime *Anime `json:"anime,omitempty"`
}

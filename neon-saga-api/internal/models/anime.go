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

// AnimeType — формат (TV/MOVIE/OVA/ONA/SPECIAL). Произвольная строка.
type AnimeType string

const (
	TypeTV      AnimeType = "TV"
	TypeMovie   AnimeType = "MOVIE"
	TypeOVA     AnimeType = "OVA"
	TypeONA     AnimeType = "ONA"
	TypeSpecial AnimeType = "SPECIAL"
)

type Anime struct {
	ID            uuid.UUID   `json:"id"`
	Slug          string      `json:"slug"`
	Title         string      `json:"title"`
	TitleEn       *string     `json:"title_en,omitempty"`
	Description   string      `json:"description"`
	Cover         *string     `json:"cover,omitempty"`
	Banner        *string     `json:"banner,omitempty"`
	Year          int         `json:"year"`
	Country       string      `json:"country"`
	Type          AnimeType   `json:"type"`
	Rating        float32     `json:"rating"`
	Status        AnimeStatus `json:"status"`
	Genres        []string    `json:"genres"`
	Views         int64       `json:"views"`
	CommentsCount int         `json:"comments_count"`
	FranchiseID   *uuid.UUID  `json:"franchise_id,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	NextEpisodeAt *time.Time  `json:"next_episode_at,omitempty"`
}

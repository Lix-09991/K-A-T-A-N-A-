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

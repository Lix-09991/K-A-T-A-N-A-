package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Lix-09991/neon-saga-api/internal/models"
)

var (
	ErrEpisodeNotFound    = errors.New("эпизод не найден")
	ErrEpisodeNumberTaken = errors.New("эпизод с таким номером уже существует")
)

type EpisodeRepository struct {
	pool *pgxpool.Pool
}

func NewEpisodeRepository(pool *pgxpool.Pool) *EpisodeRepository {
	return &EpisodeRepository{pool: pool}
}

const episodeColumns = `
	id, anime_id, number, title, description, duration,
	video_url, thumbnail, aired_at, created_at
`

// ============================================================
// CREATE
// ============================================================
func (r *EpisodeRepository) Create(ctx context.Context, e *models.Episode) error {
	const q = `
		INSERT INTO episodes (
			anime_id, number, title, description, duration, video_url, thumbnail, aired_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, q,
		e.AnimeID, e.Number, e.Title, e.Description,
		e.Duration, e.VideoURL, e.Thumbnail, e.AiredAt,
	).Scan(&e.ID, &e.CreatedAt)

	if err != nil {
		if isUniqueViolation(err, "episodes_anime_id_number_key") {
			return ErrEpisodeNumberTaken
		}
		return fmt.Errorf("создание эпизода: %w", err)
	}
	return nil
}

// ============================================================
// READ
// ============================================================
func (r *EpisodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Episode, error) {
	q := `SELECT ` + episodeColumns + ` FROM episodes WHERE id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, id))
}

func (r *EpisodeRepository) ListByAnime(ctx context.Context, animeID uuid.UUID) ([]models.Episode, error) {
	q := `SELECT ` + episodeColumns + ` FROM episodes WHERE anime_id = $1 ORDER BY number ASC`

	rows, err := r.pool.Query(ctx, q, animeID)
	if err != nil {
		return nil, fmt.Errorf("список эпизодов: %w", err)
	}
	defer rows.Close()

	result := make([]models.Episode, 0)
	for rows.Next() {
		var e models.Episode
		if err := rows.Scan(
			&e.ID, &e.AnimeID, &e.Number, &e.Title, &e.Description,
			&e.Duration, &e.VideoURL, &e.Thumbnail, &e.AiredAt, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan эпизода: %w", err)
		}
		result = append(result, e)
	}
	return result, nil
}

// ============================================================
// UPDATE / DELETE
// ============================================================
func (r *EpisodeRepository) Update(ctx context.Context, e *models.Episode) error {
	const q = `
		UPDATE episodes SET
			number = $2, title = $3, description = $4,
			duration = $5, video_url = $6, thumbnail = $7, aired_at = $8
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, q,
		e.ID, e.Number, e.Title, e.Description,
		e.Duration, e.VideoURL, e.Thumbnail, e.AiredAt,
	)
	if err != nil {
		if isUniqueViolation(err, "episodes_anime_id_number_key") {
			return ErrEpisodeNumberTaken
		}
		return fmt.Errorf("обновление эпизода: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrEpisodeNotFound
	}
	return nil
}

func (r *EpisodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM episodes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("удаление эпизода: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrEpisodeNotFound
	}
	return nil
}

// ============================================================
// SCAN
// ============================================================
func (r *EpisodeRepository) scanOne(row pgx.Row) (*models.Episode, error) {
	var e models.Episode
	err := row.Scan(
		&e.ID, &e.AnimeID, &e.Number, &e.Title, &e.Description,
		&e.Duration, &e.VideoURL, &e.Thumbnail, &e.AiredAt, &e.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEpisodeNotFound
		}
		return nil, fmt.Errorf("scan эпизода: %w", err)
	}
	return &e, nil
}

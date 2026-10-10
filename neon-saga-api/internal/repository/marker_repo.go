package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Lix-09991/neon-saga-api/internal/models"
)

var ErrMarkerNotFound = errors.New("маркер не найден")

type MarkerRepository struct {
	pool *pgxpool.Pool
}

func NewMarkerRepository(pool *pgxpool.Pool) *MarkerRepository {
	return &MarkerRepository{pool: pool}
}

// ReplaceForEpisode — атомарно заменяет ВСЕ маркеры для эпизода.
func (r *MarkerRepository) ReplaceForEpisode(
	ctx context.Context,
	episodeID uuid.UUID,
	markers []models.EpisodeMarker,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("начало транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM episode_markers WHERE episode_id = $1`, episodeID); err != nil {
		return fmt.Errorf("удаление старых маркеров: %w", err)
	}

	const q = `
		INSERT INTO episode_markers (episode_id, type, start_time, end_time, label)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	for i := range markers {
		m := &markers[i]
		m.EpisodeID = episodeID

		err := tx.QueryRow(ctx, q,
			m.EpisodeID, string(m.Type), m.StartTime, m.EndTime, m.Label,
		).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return fmt.Errorf("вставка маркера: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *MarkerRepository) ListByEpisode(ctx context.Context, episodeID uuid.UUID) ([]models.EpisodeMarker, error) {
	const q = `
		SELECT id, episode_id, type, start_time, end_time, label, created_at, updated_at
		FROM episode_markers
		WHERE episode_id = $1
		ORDER BY start_time ASC
	`
	rows, err := r.pool.Query(ctx, q, episodeID)
	if err != nil {
		return nil, fmt.Errorf("список маркеров: %w", err)
	}
	defer rows.Close()

	result := make([]models.EpisodeMarker, 0)
	for rows.Next() {
		var m models.EpisodeMarker
		var typeStr string
		if err := rows.Scan(
			&m.ID, &m.EpisodeID, &typeStr, &m.StartTime,
			&m.EndTime, &m.Label, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan маркера: %w", err)
		}
		m.Type = models.MarkerType(typeStr)
		result = append(result, m)
	}
	return result, nil
}

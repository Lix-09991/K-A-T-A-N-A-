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
	ErrFranchiseNotFound  = errors.New("франшиза не найдена")
	ErrFranchiseSlugTaken = errors.New("slug франшизы занят")
)

type FranchiseRepository struct {
	pool *pgxpool.Pool
}

func NewFranchiseRepository(pool *pgxpool.Pool) *FranchiseRepository {
	return &FranchiseRepository{pool: pool}
}

// ============================================================
// CREATE
// ============================================================
func (r *FranchiseRepository) Create(ctx context.Context, f *models.Franchise) error {
	const q = `
		INSERT INTO franchises (slug, title, description, cover)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, q, f.Slug, f.Title, f.Description, f.Cover).
		Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err, "franchises_slug_key") {
			return ErrFranchiseSlugTaken
		}
		return fmt.Errorf("создание франшизы: %w", err)
	}
	return nil
}

// ============================================================
// READ
// ============================================================
func (r *FranchiseRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Franchise, error) {
	const q = `
		SELECT id, slug, title, description, cover, created_at, updated_at
		FROM franchises WHERE id = $1
	`
	var f models.Franchise
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&f.ID, &f.Slug, &f.Title, &f.Description, &f.Cover,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFranchiseNotFound
		}
		return nil, fmt.Errorf("scan франшизы: %w", err)
	}

	items, err := r.loadItems(ctx, f.ID)
	if err != nil {
		return nil, err
	}
	f.Items = items
	return &f, nil
}

func (r *FranchiseRepository) List(ctx context.Context, limit, offset int) ([]models.Franchise, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM franchises`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count франшиз: %w", err)
	}

	const q = `
		SELECT id, slug, title, description, cover, created_at, updated_at
		FROM franchises
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("список франшиз: %w", err)
	}
	defer rows.Close()

	result := make([]models.Franchise, 0, limit)
	for rows.Next() {
		var f models.Franchise
		if err := rows.Scan(
			&f.ID, &f.Slug, &f.Title, &f.Description, &f.Cover,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan франшизы: %w", err)
		}
		result = append(result, f)
	}
	return result, total, nil
}

// ============================================================
// UPDATE / DELETE
// ============================================================
func (r *FranchiseRepository) Update(ctx context.Context, f *models.Franchise) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE franchises SET title = $2, description = $3, cover = $4
		WHERE id = $1
	`, f.ID, f.Title, f.Description, f.Cover)
	if err != nil {
		return fmt.Errorf("обновление франшизы: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrFranchiseNotFound
	}
	return nil
}

func (r *FranchiseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM franchises WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("удаление франшизы: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrFranchiseNotFound
	}
	return nil
}

// ============================================================
// ITEMS (порядок просмотра)
// ============================================================
// ReplaceItems — атомарно заменяет ВСЕ items франшизы.
// Также синхронизирует anime.franchise_id.
func (r *FranchiseRepository) ReplaceItems(
	ctx context.Context,
	franchiseID uuid.UUID,
	items []models.FranchiseItem,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("начало транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Удаляем старые items
	if _, err := tx.Exec(ctx, `DELETE FROM franchise_items WHERE franchise_id = $1`, franchiseID); err != nil {
		return fmt.Errorf("удаление старых items: %w", err)
	}

	// 2. Сбрасываем привязку всех аниме к этой франшизе
	if _, err := tx.Exec(ctx, `UPDATE anime SET franchise_id = NULL WHERE franchise_id = $1`, franchiseID); err != nil {
		return fmt.Errorf("сброс franchise_id: %w", err)
	}

	// 3. Вставляем новые items + обновляем anime.franchise_id
	const q = `
		INSERT INTO franchise_items (franchise_id, anime_id, position, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	for i := range items {
		it := &items[i]
		it.FranchiseID = franchiseID

		if err := tx.QueryRow(ctx, q,
			it.FranchiseID, it.AnimeID, it.Position, it.Note,
		).Scan(&it.ID, &it.CreatedAt); err != nil {
			return fmt.Errorf("вставка item: %w", err)
		}

		if _, err := tx.Exec(ctx,
			`UPDATE anime SET franchise_id = $1 WHERE id = $2`,
			franchiseID, it.AnimeID,
		); err != nil {
			return fmt.Errorf("обновление anime.franchise_id: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit транзакции: %w", err)
	}
	return nil
}

// ============================================================
// loadItems — загрузка items с полным Anime через JOIN
// ============================================================
func (r *FranchiseRepository) loadItems(
	ctx context.Context,
	franchiseID uuid.UUID,
) ([]models.FranchiseItem, error) {
	const q = `
		SELECT fi.id, fi.franchise_id, fi.anime_id, fi.position, fi.note, fi.created_at,
		       a.id, a.slug, a.title, a.title_en, a.description, a.cover, a.banner,
		       a.year, a.country, a.type, a.rating, a.status, a.genres,
		       a.views, a.comments_count, a.franchise_id, a.created_at, a.updated_at
		FROM franchise_items fi
		JOIN anime a ON a.id = fi.anime_id
		WHERE fi.franchise_id = $1
		ORDER BY fi.position ASC
	`
	rows, err := r.pool.Query(ctx, q, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("загрузка items: %w", err)
	}
	defer rows.Close()

	result := make([]models.FranchiseItem, 0)
	for rows.Next() {
		var (
			it        models.FranchiseItem
			a         models.Anime
			typeStr   string
			statusStr string
		)
		if err := rows.Scan(
			&it.ID, &it.FranchiseID, &it.AnimeID, &it.Position, &it.Note, &it.CreatedAt,
			&a.ID, &a.Slug, &a.Title, &a.TitleEn, &a.Description, &a.Cover, &a.Banner,
			&a.Year, &a.Country, &typeStr, &a.Rating, &statusStr, &a.Genres,
			&a.Views, &a.CommentsCount, &a.FranchiseID, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		a.Type = models.AnimeType(typeStr)
		a.Status = models.AnimeStatus(statusStr)
		it.Anime = &a
		result = append(result, it)
	}
	return result, nil
}

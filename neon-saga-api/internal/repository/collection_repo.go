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
	ErrCollectionNotFound  = errors.New("подборка не найдена")
	ErrCollectionSlugTaken = errors.New("slug подборки занят")
)

type CollectionRepository struct {
	pool *pgxpool.Pool
}

func NewCollectionRepository(pool *pgxpool.Pool) *CollectionRepository {
	return &CollectionRepository{pool: pool}
}

func (r *CollectionRepository) Create(ctx context.Context, c *models.Collection) error {
	const q = `
		INSERT INTO collections (slug, title, description, cover, is_featured, position)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, q,
		c.Slug, c.Title, c.Description, c.Cover, c.IsFeatured, c.Position,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err, "collections_slug_key") {
			return ErrCollectionSlugTaken
		}
		return fmt.Errorf("создание подборки: %w", err)
	}
	return nil
}

func (r *CollectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	const q = `
		SELECT id, slug, title, description, cover, is_featured, position, created_at, updated_at
		FROM collections WHERE id = $1
	`
	var c models.Collection
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&c.ID, &c.Slug, &c.Title, &c.Description, &c.Cover,
		&c.IsFeatured, &c.Position, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCollectionNotFound
		}
		return nil, err
	}

	items, err := r.loadItems(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.Items = items
	c.AnimeCount = len(items)
	return &c, nil
}

type CollectionFilter struct {
	Featured *bool
	Limit    int
	Offset   int
}

func (r *CollectionRepository) List(ctx context.Context, f CollectionFilter) ([]models.Collection, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	where := ""
	args := []any{}
	if f.Featured != nil {
		args = append(args, *f.Featured)
		where = fmt.Sprintf("WHERE is_featured = $%d", len(args))
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM collections "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`
		SELECT c.id, c.slug, c.title, c.description, c.cover, c.is_featured, c.position,
		       c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM collection_items ci WHERE ci.collection_id = c.id) AS anime_count
		FROM collections c
		%s
		ORDER BY c.position ASC, c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []models.Collection
	for rows.Next() {
		var c models.Collection
		if err := rows.Scan(
			&c.ID, &c.Slug, &c.Title, &c.Description, &c.Cover,
			&c.IsFeatured, &c.Position, &c.CreatedAt, &c.UpdatedAt, &c.AnimeCount,
		); err != nil {
			return nil, 0, err
		}
		result = append(result, c)
	}
	return result, total, nil
}

func (r *CollectionRepository) Update(ctx context.Context, c *models.Collection) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE collections
		SET title=$2, description=$3, cover=$4, is_featured=$5, position=$6
		WHERE id=$1
	`, c.ID, c.Title, c.Description, c.Cover, c.IsFeatured, c.Position)
	if err != nil {
		return fmt.Errorf("обновление подборки: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCollectionNotFound
	}
	return nil
}

func (r *CollectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM collections WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCollectionNotFound
	}
	return nil
}

func (r *CollectionRepository) ReplaceItems(ctx context.Context, collectionID uuid.UUID, items []models.CollectionItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM collection_items WHERE collection_id = $1`, collectionID); err != nil {
		return err
	}

	const q = `
		INSERT INTO collection_items (collection_id, anime_id, position, note)
		VALUES ($1,$2,$3,$4)
		RETURNING id, created_at
	`
	for i := range items {
		it := &items[i]
		it.CollectionID = collectionID
		if err := tx.QueryRow(ctx, q, it.CollectionID, it.AnimeID, it.Position, it.Note).
			Scan(&it.ID, &it.CreatedAt); err != nil {
			return fmt.Errorf("вставка item: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *CollectionRepository) loadItems(ctx context.Context, collectionID uuid.UUID) ([]models.CollectionItem, error) {
	const q = `
		SELECT ci.id, ci.collection_id, ci.anime_id, ci.position, ci.note, ci.created_at,
		       a.id, a.slug, a.title, a.title_en, a.description, a.cover, a.banner,
		       a.year, a.country, a.type, a.rating, a.status, a.genres,
		       a.views, a.comments_count, a.franchise_id, a.created_at, a.updated_at
		FROM collection_items ci
		JOIN anime a ON a.id = ci.anime_id
		WHERE ci.collection_id = $1
		ORDER BY ci.position ASC
	`
	rows, err := r.pool.Query(ctx, q, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.CollectionItem
	for rows.Next() {
		var it models.CollectionItem
		var a models.Anime
		var typeStr, statusStr string
		if err := rows.Scan(
			&it.ID, &it.CollectionID, &it.AnimeID, &it.Position, &it.Note, &it.CreatedAt,
			&a.ID, &a.Slug, &a.Title, &a.TitleEn, &a.Description, &a.Cover, &a.Banner,
			&a.Year, &a.Country, &typeStr, &a.Rating, &statusStr, &a.Genres,
			&a.Views, &a.CommentsCount, &a.FranchiseID, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		a.Type = models.AnimeType(typeStr)
		a.Status = models.AnimeStatus(statusStr)
		it.Anime = &a
		result = append(result, it)
	}
	return result, nil
}

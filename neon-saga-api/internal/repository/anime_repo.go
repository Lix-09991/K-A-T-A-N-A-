package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/Lix-09991/neon-saga-api/internal/models"
)

var (
	ErrAnimeNotFound = errors.New("аниме не найдено")
	ErrSlugTaken     = errors.New("slug уже занят")
)

type AnimeRepository struct {
	pool *pgxpool.Pool
}

func NewAnimeRepository(pool *pgxpool.Pool) *AnimeRepository {
	return &AnimeRepository{pool: pool}
}

// ============================================================
// ФИЛЬТРЫ
// ============================================================
type ListFilter struct {
	Search      string
	Genres      []string
	GenresMode  string
	Countries   []string
	Types       []string
	Statuses    []string
	YearFrom    int
	YearTo      int
	MinRating   float64
	FranchiseID *uuid.UUID
	Limit       int
	Offset      int
	Sort        string
	Order       string
}

var allowedSorts = map[string]string{
	"created_at":     "created_at",
	"updated_at":     "updated_at",
	"views":          "views",
	"rating":         "rating",
	"year":           "year",
	"comments_count": "comments_count",
	"title":          "title",
}

func (r *AnimeRepository) buildWhere(f ListFilter, exclude string) (string, []any) {
	var parts []string
	var args []any
	add := func(cond string, val any) {
		args = append(args, val)
		parts = append(parts, fmt.Sprintf(cond, len(args)))
	}

	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		n := len(args)
		parts = append(parts, fmt.Sprintf("(title ILIKE $%d OR COALESCE(title_en,'') ILIKE $%d)", n, n))
	}

	if exclude != "genres" && len(f.Genres) > 0 {
		if strings.EqualFold(f.GenresMode, "and") {
			args = append(args, f.Genres)
			parts = append(parts, fmt.Sprintf("genres @> $%d", len(args)))
		} else {
			args = append(args, f.Genres)
			parts = append(parts, fmt.Sprintf("genres && $%d", len(args)))
		}
	}

	if exclude != "countries" && len(f.Countries) > 0 {
		add("country = ANY($%d)", f.Countries)
	}
	if exclude != "types" && len(f.Types) > 0 {
		add("type = ANY($%d)", f.Types)
	}
	if exclude != "statuses" && len(f.Statuses) > 0 {
		add("status = ANY($%d)", f.Statuses)
	}
	if exclude != "years" {
		if f.YearFrom > 0 {
			add("year >= $%d", f.YearFrom)
		}
		if f.YearTo > 0 {
			add("year <= $%d", f.YearTo)
		}
	}
	if exclude != "rating" && f.MinRating > 0 {
		add("rating >= $%d", f.MinRating)
	}
	if f.FranchiseID != nil {
		add("franchise_id = $%d", *f.FranchiseID)
	}

	if len(parts) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// ============================================================
// CREATE
// ============================================================
func (r *AnimeRepository) Create(ctx context.Context, a *models.Anime) error {
	const q = `
		INSERT INTO anime (
			slug, title, title_en, description, cover, banner,
			year, country, type, status, genres, franchise_id, next_episode_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, rating, views, comments_count, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, q,
		a.Slug, a.Title, a.TitleEn, a.Description, a.Cover, a.Banner,
		a.Year, a.Country, string(a.Type), string(a.Status), a.Genres, a.FranchiseID, a.NextEpisodeAt,
	).Scan(&a.ID, &a.Rating, &a.Views, &a.CommentsCount, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err, "anime_slug_key") {
			return ErrSlugTaken
		}
		return fmt.Errorf("создание аниме: %w", err)
	}
	return nil
}

// ============================================================
// READ
// ============================================================
const animeColumns = `
	id, slug, title, title_en, description, cover, banner,
	year, country, type, rating, status, genres,
	views, comments_count, franchise_id, created_at, updated_at, next_episode_at
`

func (r *AnimeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Anime, error) {
	q := `SELECT ` + animeColumns + ` FROM anime WHERE id = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, id))
}

func (r *AnimeRepository) FindBySlug(ctx context.Context, slug string) (*models.Anime, error) {
	q := `SELECT ` + animeColumns + ` FROM anime WHERE slug = $1`
	return r.scanOne(r.pool.QueryRow(ctx, q, slug))
}

func (r *AnimeRepository) List(ctx context.Context, f ListFilter) ([]models.Anime, int, error) {
	where, args := r.buildWhere(f, "")

	var total int
	countQ := "SELECT COUNT(*) FROM anime " + where
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count аниме: %w", err)
	}

	sortCol, ok := allowedSorts[f.Sort]
	if !ok {
		sortCol = "created_at"
	}
	order := "DESC"
	if strings.EqualFold(f.Order, "asc") {
		order = "ASC"
	}

	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 24
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	args = append(args, f.Limit, f.Offset)
	limitN := len(args) - 1
	offsetN := len(args)

	listQ := fmt.Sprintf(`
		SELECT %s FROM anime
		%s
		ORDER BY %s %s, id DESC
		LIMIT $%d OFFSET $%d
	`, animeColumns, where, sortCol, order, limitN, offsetN)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("список аниме: %w", err)
	}
	defer rows.Close()

	result := make([]models.Anime, 0, f.Limit)
	for rows.Next() {
		a, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *a)
	}
	return result, total, nil
}

// ============================================================
// FACETS
// ============================================================
type FacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
type YearFacet struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}
type Facets struct {
	Total     int          `json:"total"`
	Genres    []FacetValue `json:"genres"`
	Countries []FacetValue `json:"countries"`
	Types     []FacetValue `json:"types"`
	Statuses  []FacetValue `json:"statuses"`
	Years     []YearFacet  `json:"years"`
}

func (r *AnimeRepository) Facets(ctx context.Context, f ListFilter) (*Facets, error) {
	result := &Facets{
		Genres:    []FacetValue{},
		Countries: []FacetValue{},
		Types:     []FacetValue{},
		Statuses:  []FacetValue{},
		Years:     []YearFacet{},
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		where, args := r.buildWhere(f, "")
		return r.pool.QueryRow(gctx, "SELECT COUNT(*) FROM anime "+where, args...).Scan(&result.Total)
	})

	g.Go(func() error {
		where, args := r.buildWhere(f, "genres")
		q := `SELECT g, COUNT(*) FROM anime, unnest(genres) AS g ` + where + ` GROUP BY g ORDER BY COUNT(*) DESC LIMIT 50`
		rows, err := r.pool.Query(gctx, q, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v FacetValue
			if err := rows.Scan(&v.Value, &v.Count); err != nil {
				return err
			}
			result.Genres = append(result.Genres, v)
		}
		return nil
	})

	g.Go(func() error {
		where, args := r.buildWhere(f, "countries")
		return r.scanFacet(gctx, `SELECT country AS v, COUNT(*) FROM anime `+where+` GROUP BY v ORDER BY COUNT(*) DESC`, args, &result.Countries)
	})

	g.Go(func() error {
		where, args := r.buildWhere(f, "types")
		return r.scanFacet(gctx, `SELECT type AS v, COUNT(*) FROM anime `+where+` GROUP BY v ORDER BY COUNT(*) DESC`, args, &result.Types)
	})

	g.Go(func() error {
		where, args := r.buildWhere(f, "statuses")
		return r.scanFacet(gctx, `SELECT status AS v, COUNT(*) FROM anime `+where+` GROUP BY v ORDER BY COUNT(*) DESC`, args, &result.Statuses)
	})

	g.Go(func() error {
		where, args := r.buildWhere(f, "years")
		q := `SELECT year, COUNT(*) FROM anime ` + where + ` GROUP BY year ORDER BY year DESC LIMIT 30`
		rows, err := r.pool.Query(gctx, q, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var y YearFacet
			if err := rows.Scan(&y.Year, &y.Count); err != nil {
				return err
			}
			result.Years = append(result.Years, y)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("facets: %w", err)
	}
	return result, nil
}

func (r *AnimeRepository) scanFacet(ctx context.Context, q string, args []any, dst *[]FacetValue) error {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var v FacetValue
		if err := rows.Scan(&v.Value, &v.Count); err != nil {
			return err
		}
		*dst = append(*dst, v)
	}
	return nil
}

// ============================================================
// UPDATE / DELETE
// ============================================================
func (r *AnimeRepository) Update(ctx context.Context, a *models.Anime) error {
	const q = `
		UPDATE anime SET
			title = $2, title_en = $3, description = $4,
			cover = $5, banner = $6, year = $7, country = $8,
			type = $9, status = $10, genres = $11, franchise_id = $12, next_episode_at = $13
		WHERE id = $1
		RETURNING updated_at
	`
	err := r.pool.QueryRow(ctx, q,
		a.ID, a.Title, a.TitleEn, a.Description, a.Cover, a.Banner,
		a.Year, a.Country, string(a.Type), string(a.Status), a.Genres, a.FranchiseID, a.NextEpisodeAt,
	).Scan(&a.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAnimeNotFound
		}
		return fmt.Errorf("обновление аниме: %w", err)
	}
	return nil
}

func (r *AnimeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM anime WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("удаление аниме: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAnimeNotFound
	}
	return nil
}

func (r *AnimeRepository) IncrementViews(ctx context.Context, animeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE anime SET views = views + 1 WHERE id = $1`, animeID)
	return err
}

func (r *AnimeRepository) IncrementViewsByEpisode(ctx context.Context, episodeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE anime SET views = views + 1
		WHERE id = (SELECT anime_id FROM episodes WHERE id = $1)
	`, episodeID)
	return err
}

// ============================================================
// SCAN helpers
// ============================================================
func (r *AnimeRepository) scanOne(row pgx.Row) (*models.Anime, error) {
	var a models.Anime
	var typeStr, statusStr string
	err := row.Scan(
		&a.ID, &a.Slug, &a.Title, &a.TitleEn, &a.Description,
		&a.Cover, &a.Banner, &a.Year, &a.Country, &typeStr, &a.Rating,
		&statusStr, &a.Genres, &a.Views, &a.CommentsCount, &a.FranchiseID,
		&a.CreatedAt, &a.UpdatedAt, &a.NextEpisodeAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAnimeNotFound
		}
		return nil, fmt.Errorf("scan аниме: %w", err)
	}
	a.Type = models.AnimeType(typeStr)
	a.Status = models.AnimeStatus(statusStr)
	return &a, nil
}

func (r *AnimeRepository) scanRow(rows pgx.Rows) (*models.Anime, error) {
	var a models.Anime
	var typeStr, statusStr string
	err := rows.Scan(
		&a.ID, &a.Slug, &a.Title, &a.TitleEn, &a.Description,
		&a.Cover, &a.Banner, &a.Year, &a.Country, &typeStr, &a.Rating,
		&statusStr, &a.Genres, &a.Views, &a.CommentsCount, &a.FranchiseID,
		&a.CreatedAt, &a.UpdatedAt, &a.NextEpisodeAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan row аниме: %w", err)
	}
	a.Type = models.AnimeType(typeStr)
	a.Status = models.AnimeStatus(statusStr)
	return &a, nil
}

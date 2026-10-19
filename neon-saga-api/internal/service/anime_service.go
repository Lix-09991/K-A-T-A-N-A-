package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
)

var ErrAnimeNotFound = errors.New("аниме не найдено")

type AnimeService struct {
	animeRepo *repository.AnimeRepository
	epRepo    *repository.EpisodeRepository
}

func NewAnimeService(animeRepo *repository.AnimeRepository, epRepo *repository.EpisodeRepository) *AnimeService {
	return &AnimeService{animeRepo: animeRepo, epRepo: epRepo}
}

type CreateAnimeInput struct {
	Slug          string
	Title         string
	TitleEn       *string
	Description   string
	Cover         *string
	Banner        *string
	Year          int
	Country       string
	Type          models.AnimeType
	Status        models.AnimeStatus
	Genres        []string
	FranchiseID   *uuid.UUID
	NextEpisodeAt *time.Time
}

type UpdateAnimeInput = CreateAnimeInput

func (s *AnimeService) Create(ctx context.Context, in CreateAnimeInput) (*models.Anime, error) {
	if in.Slug == "" {
		in.Slug = slugify(in.Title)
	}
	if in.Country == "" {
		in.Country = "JP"
	}
	if in.Type == "" {
		in.Type = models.TypeTV
	}
	if in.Status == "" {
		in.Status = models.StatusOngoing
	}

	a := &models.Anime{
		Slug:          in.Slug,
		Title:         in.Title,
		TitleEn:       in.TitleEn,
		Description:   in.Description,
		Cover:         in.Cover,
		Banner:        in.Banner,
		Year:          in.Year,
		Country:       in.Country,
		Type:          in.Type,
		Status:        in.Status,
		Genres:        in.Genres,
		FranchiseID:   in.FranchiseID,
		NextEpisodeAt: in.NextEpisodeAt,
	}
	if a.Genres == nil {
		a.Genres = []string{}
	}
	if err := s.animeRepo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AnimeService) GetByID(ctx context.Context, id uuid.UUID) (*models.Anime, []models.Episode, error) {
	a, err := s.animeRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAnimeNotFound) {
			return nil, nil, ErrAnimeNotFound
		}
		return nil, nil, err
	}
	eps, err := s.epRepo.ListByAnime(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		_ = s.animeRepo.IncrementViews(context.Background(), id)
	}()
	return a, eps, nil
}

func (s *AnimeService) List(ctx context.Context, f repository.ListFilter) ([]models.Anime, int, error) {
	return s.animeRepo.List(ctx, f)
}

func (s *AnimeService) Facets(ctx context.Context, f repository.ListFilter) (*repository.Facets, error) {
	return s.animeRepo.Facets(ctx, f)
}

func (s *AnimeService) Update(ctx context.Context, id uuid.UUID, in UpdateAnimeInput) (*models.Anime, error) {
	a, err := s.animeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Title = in.Title
	a.TitleEn = in.TitleEn
	a.Description = in.Description
	a.Cover = in.Cover
	a.Banner = in.Banner
	a.Year = in.Year
	a.Country = in.Country
	a.Type = in.Type
	a.Status = in.Status
	a.Genres = in.Genres
	a.FranchiseID = in.FranchiseID
	a.NextEpisodeAt = in.NextEpisodeAt

	if err := s.animeRepo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AnimeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.animeRepo.Delete(ctx, id)
}

// ============================================================
// SLUGIFY
// ============================================================
var (
	translitMap = map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	}
	nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
)

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if v, ok := translitMap[r]; ok {
			b.WriteString(v)
		} else {
			b.WriteRune(r)
		}
	}
	return strings.Trim(nonAlnum.ReplaceAllString(b.String(), "-"), "-")
}

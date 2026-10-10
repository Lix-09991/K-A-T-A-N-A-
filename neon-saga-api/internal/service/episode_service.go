package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
)

var ErrEpisodeNotFound = errors.New("эпизод не найден")

type EpisodeService struct {
	epRepo     *repository.EpisodeRepository
	markerRepo *repository.MarkerRepository
	animeRepo  *repository.AnimeRepository
}

func NewEpisodeService(
	epRepo *repository.EpisodeRepository,
	markerRepo *repository.MarkerRepository,
	animeRepo *repository.AnimeRepository,
) *EpisodeService {
	return &EpisodeService{
		epRepo:     epRepo,
		markerRepo: markerRepo,
		animeRepo:  animeRepo,
	}
}

type CreateEpisodeInput struct {
	AnimeID     uuid.UUID
	Number      int
	Title       string
	Description *string
	Duration    int
	VideoURL    string
	Thumbnail   *string
}

func (s *EpisodeService) Create(ctx context.Context, in CreateEpisodeInput) (*models.Episode, error) {
	// Проверяем, что аниме существует
	if _, err := s.animeRepo.FindByID(ctx, in.AnimeID); err != nil {
		return nil, err
	}

	e := &models.Episode{
		AnimeID:     in.AnimeID,
		Number:      in.Number,
		Title:       in.Title,
		Description: in.Description,
		Duration:    in.Duration,
		VideoURL:    in.VideoURL,
		Thumbnail:   in.Thumbnail,
	}
	if err := s.epRepo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EpisodeService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Episode, []models.EpisodeMarker, error) {
	e, err := s.epRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrEpisodeNotFound) {
			return nil, nil, ErrEpisodeNotFound
		}
		return nil, nil, err
	}

	markers, err := s.markerRepo.ListByEpisode(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Async increment views через аниме
	go func(animeID uuid.UUID) {
		_ = s.animeRepo.IncrementViews(context.Background(), animeID)
	}(e.AnimeID)

	return e, markers, nil
}

func (s *EpisodeService) ListByAnime(ctx context.Context, animeID uuid.UUID) ([]models.Episode, error) {
	return s.epRepo.ListByAnime(ctx, animeID)
}

func (s *EpisodeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.epRepo.Delete(ctx, id)
}

func (s *EpisodeService) SetMarkers(
	ctx context.Context,
	episodeID uuid.UUID,
	markers []models.EpisodeMarker,
) error {
	if _, err := s.epRepo.FindByID(ctx, episodeID); err != nil {
		if errors.Is(err, repository.ErrEpisodeNotFound) {
			return ErrEpisodeNotFound
		}
		return err
	}
	return s.markerRepo.ReplaceForEpisode(ctx, episodeID, markers)
}

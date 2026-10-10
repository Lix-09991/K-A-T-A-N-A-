package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
)

var ErrCollectionNotFound = errors.New("подборка не найдена")

type CollectionService struct {
	repo *repository.CollectionRepository
}

func NewCollectionService(repo *repository.CollectionRepository) *CollectionService {
	return &CollectionService{repo: repo}
}

type CreateCollectionInput struct {
	Slug        string
	Title       string
	Description *string
	Cover       *string
	IsFeatured  bool
	Position    int
}

func (s *CollectionService) Create(ctx context.Context, in CreateCollectionInput) (*models.Collection, error) {
	if in.Slug == "" {
		in.Slug = slugify(in.Title)
	}
	c := &models.Collection{
		Slug:        in.Slug,
		Title:       in.Title,
		Description: in.Description,
		Cover:       in.Cover,
		IsFeatured:  in.IsFeatured,
		Position:    in.Position,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CollectionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCollectionNotFound) {
			return nil, ErrCollectionNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *CollectionService) List(ctx context.Context, f repository.CollectionFilter) ([]models.Collection, int, error) {
	return s.repo.List(ctx, f)
}

func (s *CollectionService) Update(ctx context.Context, id uuid.UUID, in CreateCollectionInput) (*models.Collection, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Title = in.Title
	c.Description = in.Description
	c.Cover = in.Cover
	c.IsFeatured = in.IsFeatured
	c.Position = in.Position
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CollectionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *CollectionService) ReplaceItems(ctx context.Context, id uuid.UUID, items []models.CollectionItem) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.ReplaceItems(ctx, id, items)
}

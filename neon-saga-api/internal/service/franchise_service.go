package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
	"github.com/Lix-09991/neon-saga-api/internal/repository"
)

var ErrFranchiseNotFound = errors.New("франшиза не найдена")

type FranchiseService struct {
	repo *repository.FranchiseRepository
}

func NewFranchiseService(repo *repository.FranchiseRepository) *FranchiseService {
	return &FranchiseService{repo: repo}
}

type CreateFranchiseInput struct {
	Slug        string
	Title       string
	Description *string
	Cover       *string
}

func (s *FranchiseService) Create(ctx context.Context, in CreateFranchiseInput) (*models.Franchise, error) {
	if in.Slug == "" {
		in.Slug = slugify(in.Title)
	}
	f := &models.Franchise{
		Slug: in.Slug, Title: in.Title, Description: in.Description, Cover: in.Cover,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FranchiseService) GetByID(ctx context.Context, id uuid.UUID) (*models.Franchise, error) {
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrFranchiseNotFound) {
			return nil, ErrFranchiseNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *FranchiseService) List(ctx context.Context, limit, offset int) ([]models.Franchise, int, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *FranchiseService) Update(ctx context.Context, id uuid.UUID, in CreateFranchiseInput) (*models.Franchise, error) {
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	f.Title = in.Title
	f.Description = in.Description
	f.Cover = in.Cover
	if err := s.repo.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FranchiseService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *FranchiseService) ReplaceItems(ctx context.Context, id uuid.UUID, items []models.FranchiseItem) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.ReplaceItems(ctx, id, items)
}

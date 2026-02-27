package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type attractionService struct {
	repo domain.AttractionRepository
}

func NewAttractionService(repo domain.AttractionRepository) domain.AttractionService {
	return &attractionService{repo: repo}
}

func (s *attractionService) GetAttractionByID(ctx context.Context, id string) (*domain.Attraction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *attractionService) GetAttractionByName(ctx context.Context, name string) ([]*domain.Attraction, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *attractionService) ListAttractions(ctx context.Context, filter domain.AttractionFilter) ([]*domain.Attraction, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *attractionService) GetAttractionCategories(ctx context.Context) ([]*domain.AttractionCategory, error) {
	return s.repo.GetCategories(ctx)
}

func (s *attractionService) GetAttractionTypes(ctx context.Context) ([]*domain.AttractionType, error) {
	return s.repo.GetTypes(ctx)
}

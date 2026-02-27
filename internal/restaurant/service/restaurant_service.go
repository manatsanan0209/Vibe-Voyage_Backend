package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type restaurantService struct {
	repo domain.RestaurantRepository
}

func NewRestaurantService(repo domain.RestaurantRepository) domain.RestaurantService {
	return &restaurantService{repo: repo}
}

func (s *restaurantService) GetRestaurantByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *restaurantService) ListRestaurants(ctx context.Context, filter domain.RestaurantFilter) ([]*domain.Restaurant, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *restaurantService) GetFoodTypes(ctx context.Context) ([]*domain.FoodType, error) {
	return s.repo.GetFoodTypes(ctx)
}

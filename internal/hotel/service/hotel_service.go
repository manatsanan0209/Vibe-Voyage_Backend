package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type hotelService struct {
	repo domain.HotelRepository
}

func NewHotelService(repo domain.HotelRepository) domain.HotelService {
	return &hotelService{repo: repo}
}

func (s *hotelService) GetHotelByID(ctx context.Context, id string) (*domain.Hotel, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *hotelService) GetHotelByName(ctx context.Context, name string) ([]*domain.Hotel, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *hotelService) ListHotels(ctx context.Context, filter domain.HotelFilter) ([]*domain.Hotel, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *hotelService) GetAccommodationTypes(ctx context.Context) ([]*domain.AccommodationType, error) {
	return s.repo.GetAccommodationTypes(ctx)
}

func (s *hotelService) GetPriceRanges(ctx context.Context) ([]*domain.PriceRange, error) {
	return s.repo.GetPriceRanges(ctx)
}

package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type tripSuggestionService struct {
	repo domain.TripSuggestionRepository
}

func NewTripSuggestionService(repo domain.TripSuggestionRepository) domain.TripSuggestionService {
	return &tripSuggestionService{repo: repo}
}

func (s *tripSuggestionService) GetFeed(ctx context.Context, page, limit int, userID uint) ([]domain.PublishedTripWithMeta, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.GetPublishedTrips(ctx, domain.GetPublishedTripsOptions{
		Page:   page,
		Limit:  limit,
		UserID: userID,
	})
}

func (s *tripSuggestionService) GetPublishedTripByTripID(ctx context.Context, tripID uint) (*domain.PublishedTrip, error) {
	return s.repo.GetPublishedTripByTripID(ctx, tripID)
}

func (s *tripSuggestionService) GetDetail(ctx context.Context, publishedTripID, userID uint) (*domain.PublishedTripWithMeta, error) {
	// Increment view count non-fatally before fetching detail
	_ = s.repo.IncrementViewCount(ctx, publishedTripID)
	return s.repo.GetPublishedTripByID(ctx, publishedTripID, userID)
}

func (s *tripSuggestionService) PublishTrip(ctx context.Context, tripID, userID uint, title, description string) (*domain.PublishedTrip, error) {
	return s.repo.PublishTrip(ctx, tripID, userID, title, description)
}

func (s *tripSuggestionService) UnpublishTrip(ctx context.Context, tripID, userID uint) error {
	return s.repo.UnpublishTrip(ctx, tripID, userID)
}

func (s *tripSuggestionService) ToggleLike(ctx context.Context, publishedTripID, userID uint) (bool, error) {
	return s.repo.ToggleLike(ctx, publishedTripID, userID)
}

func (s *tripSuggestionService) ToggleBookmark(ctx context.Context, publishedTripID, userID uint) (bool, error) {
	return s.repo.ToggleBookmark(ctx, publishedTripID, userID)
}

func (s *tripSuggestionService) GetBookmarks(ctx context.Context, userID uint) ([]domain.PublishedTripWithMeta, error) {
	return s.repo.GetBookmarkedTrips(ctx, userID)
}

func (s *tripSuggestionService) UseAsTemplate(ctx context.Context, publishedTripID, userID uint, input domain.UseAsTemplateInput) (*domain.CreateTripResult, error) {
	return s.repo.UseAsTemplate(ctx, publishedTripID, userID, input)
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type userLifestyleService struct {
	repo                 domain.UserLifestyleRepository
	recommendationClient domain.RecommendationClient
}

func NewUserLifestyleService(repo domain.UserLifestyleRepository, recommendationClient domain.RecommendationClient) domain.UserLifestyleService {
	return &userLifestyleService{repo: repo, recommendationClient: recommendationClient}
}

func (s *userLifestyleService) GetLifestyle(ctx context.Context, userID, roomID uint) (*domain.UserLifestyle, error) {
	lifestyle, err := s.repo.GetByUserAndRoom(ctx, userID, roomID)
	if err != nil {
		return nil, err
	}
	if lifestyle == nil {
		return nil, domain.ErrLifestyleNotFound
	}
	return lifestyle, nil
}

func (s *userLifestyleService) AnalyzeLifestyle(ctx context.Context, lifestyleID uint) ([]domain.RecommendedPlace, error) {
	lifestyle, err := s.repo.GetByID(ctx, lifestyleID)
	if err != nil {
		return nil, err
	}
	if lifestyle == nil {
		return nil, domain.ErrLifestyleNotFound
	}

	trip, err := s.repo.GetTripByRoomID(ctx, lifestyle.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip for room %d: %w", lifestyle.RoomID, err)
	}

	var travelVibes []string
	if err := json.Unmarshal([]byte(lifestyle.TravelVibes), &travelVibes); err != nil || travelVibes == nil {
		travelVibes = []string{}
	}

	var voyagePriorities []string
	if err := json.Unmarshal([]byte(lifestyle.VoyagePriorities), &voyagePriorities); err != nil || voyagePriorities == nil {
		voyagePriorities = []string{}
	}

	var foodVibes []string
	if err := json.Unmarshal([]byte(lifestyle.FoodVibes), &foodVibes); err != nil || foodVibes == nil {
		foodVibes = []string{}
	}

	log.Printf("destination: %s", trip.DestinationName)

	places, structuredJSON, err := s.recommendationClient.Recommend(ctx, domain.RecommendationRequest{
		DestinationName:  trip.DestinationName,
		DestinationID:    trip.DestinationID,
		TravelVibes:      travelVibes,
		VoyagePriorities: voyagePriorities,
		FoodVibes:        foodVibes,
		AdditionalNotes:  lifestyle.AdditionalNotes,
	})
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStructuredLifestyle(ctx, lifestyle.LifestyleID, structuredJSON); err != nil {
		return nil, fmt.Errorf("failed to update structured_lifestyle: %w", err)
	}

	return places, nil
}

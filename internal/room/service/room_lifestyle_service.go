package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (s *roomService) AddRoomLifestyle(ctx context.Context, roomID, userID uint, input domain.CreateRoomLifestyleInput) (*domain.UserLifestyle, error) {
	role, err := s.getMemberRole(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if role == domain.RoleSpectator {
		return nil, errors.New("only members with edit access can add lifestyle")
	}

	existing, err := s.lifestyleRepo.GetByUserAndRoom(ctx, userID, roomID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("lifestyle already exists for this user in this room")
	}

	if input.PreferredDestinations == nil {
		input.PreferredDestinations = []domain.PreferredDestination{}
	}
	if input.TravelVibes == nil {
		input.TravelVibes = []string{}
	}
	if input.VoyagePriorities == nil {
		input.VoyagePriorities = []string{}
	}
	if input.FoodVibes == nil {
		input.FoodVibes = []string{}
	}

	preferredDestinationsJSON, err := json.Marshal(input.PreferredDestinations)
	if err != nil {
		return nil, errors.New("failed to marshal preferred_destinations")
	}
	travelVibesJSON, err := json.Marshal(input.TravelVibes)
	if err != nil {
		return nil, errors.New("failed to marshal travel_vibes")
	}
	voyagePrioritiesJSON, err := json.Marshal(input.VoyagePriorities)
	if err != nil {
		return nil, errors.New("failed to marshal voyage_priorities")
	}
	foodVibesJSON, err := json.Marshal(input.FoodVibes)
	if err != nil {
		return nil, errors.New("failed to marshal food_vibes")
	}

	lifestyle := &domain.UserLifestyle{
		UserID:                userID,
		RoomID:                roomID,
		PreferredDestinations: string(preferredDestinationsJSON),
		TravelVibes:           string(travelVibesJSON),
		VoyagePriorities:      string(voyagePrioritiesJSON),
		FoodVibes:             string(foodVibesJSON),
		AdditionalNotes:       input.AdditionalNotes,
	}

	if err := s.lifestyleRepo.Create(ctx, lifestyle); err != nil {
		return nil, err
	}

	s.enqueueLifestyleAnalyze(lifestyle.LifestyleID, userID)

	return lifestyle, nil
}

func (s *roomService) enqueueLifestyleAnalyze(lifestyleID, userID uint) {
	if s.userLifestyleSvc == nil {
		return
	}

	go func() {
		s.analyzeSemaphore <- struct{}{}
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[RoomLifestyle] async analyze panic (lifestyle_id=%d): %v", lifestyleID, r)
			}
			<-s.analyzeSemaphore
		}()

		asyncCtx, cancel := context.WithTimeout(context.Background(), s.analyzeTimeout)
		defer cancel()

		if _, err := s.userLifestyleSvc.AnalyzeLifestyle(asyncCtx, lifestyleID); err != nil {
			log.Printf("[RoomLifestyle] async analyze failed (lifestyle_id=%d): %v", lifestyleID, err)
			return
		}

		log.Printf("[RoomLifestyle] async analyze completed (lifestyle_id=%d)", lifestyleID)

		if err := s.notifSvc.Notify(context.Background(), userID, domain.NotifTypeLifestyleAnalyzed, "Lifestyle analysis complete", "Your lifestyle analysis has been completed.", nil, nil); err != nil {
			log.Printf("[Notification] lifestyle_analyzed failed (user_id=%d): %v", userID, err)
		}
	}()
}

func (s *roomService) getMemberRole(ctx context.Context, roomID, userID uint) (int, error) {
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return 0, err
	}

	for _, member := range members {
		if member.UserID == userID {
			return member.Role, nil
		}
	}

	return 0, errors.New("you are not a member of this room")
}

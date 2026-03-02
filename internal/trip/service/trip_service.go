package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type tripService struct {
	db           *gorm.DB
	lifestyleSvc domain.UserLifestyleService
}

func NewTripService(db *gorm.DB, lifestyleSvc domain.UserLifestyleService) domain.TripService {
	return &tripService{db: db, lifestyleSvc: lifestyleSvc}
}

func (s *tripService) GetTripSchedule(ctx context.Context, tripID uint) (*domain.GetTripScheduleResult, error) {
	var trip domain.Trips
	if err := s.db.WithContext(ctx).First(&trip, tripID).Error; err != nil {
		return nil, err
	}

	var schedules []domain.TripSchedule
	if err := s.db.WithContext(ctx).
		Where("trip_id = ?", tripID).
		Order("day_number ASC, sequence_order ASC").
		Find(&schedules).Error; err != nil {
		return nil, err
	}

	var suggestions []domain.TripSchedule
	dayMap := make(map[int][]domain.TripSchedule)

	for _, item := range schedules {
		if item.DayNumber == 0 && item.SequenceOrder == 0 {
			suggestions = append(suggestions, item)
		} else {
			dayMap[item.DayNumber] = append(dayMap[item.DayNumber], item)
		}
	}

	days := make([]domain.DaySchedule, 0, len(dayMap))
	for dayNum, items := range dayMap {
		days = append(days, domain.DaySchedule{
			DayNumber: dayNum,
			Items:     items,
		})
	}

	return &domain.GetTripScheduleResult{
		Trip:        &trip,
		Suggestions: suggestions,
		Days:        days,
	}, nil
}

func (s *tripService) CreateTrip(ctx context.Context, userID uint, input domain.CreateTripInput) (*domain.CreateTripResult, error) {
	if input.RoomName == "" {
		return nil, errors.New("room_name is required")
	}
	if input.DestinationName == "" || input.DestinationID == "" {
		return nil, errors.New("destination_name and destination_id are required")
	}
	if input.StartDate.IsZero() || input.EndDate.IsZero() {
		return nil, errors.New("start_date and end_date are required")
	}
	if !input.EndDate.After(input.StartDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	var result domain.CreateTripResult

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

	preferredDestJSON, err := json.Marshal(input.PreferredDestinations)
	if err != nil {
		return nil, errors.New("failed to marshal preferred_destinations")
	}
	travelVibesJSON, err := json.Marshal(input.TravelVibes)
	if err != nil {
		return nil, errors.New("failed to marshal travel_vibes")
	}
	prioritiesJSON, err := json.Marshal(input.VoyagePriorities)
	if err != nil {
		return nil, errors.New("failed to marshal voyage_priorities")
	}
	foodVibesJSON, err := json.Marshal(input.FoodVibes)
	if err != nil {
		return nil, errors.New("failed to marshal food_vibes")
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		room := &domain.Room{
			OwnerID:   userID,
			RoomName:  input.RoomName,
			RoomImage: input.RoomImage,
		}
		if err := tx.Create(room).Error; err != nil {
			return err
		}

		trip := &domain.Trips{
			RoomID:          room.RoomID,
			DestinationName: input.DestinationName,
			DestinationID:   input.DestinationID,
			StartDate:       input.StartDate,
			EndDate:         input.EndDate,
		}
		if err := tx.Create(trip).Error; err != nil {
			return err
		}

		member := &domain.RoomMember{
			RoomID: room.RoomID,
			UserID: userID,
			Role:   domain.RoleOwner,
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		lifestyle := &domain.UserLifestyle{
			UserID:                userID,
			RoomID:                room.RoomID,
			PreferredDestinations: string(preferredDestJSON),
			TravelVibes:           string(travelVibesJSON),
			VoyagePriorities:      string(prioritiesJSON),
			FoodVibes:             string(foodVibesJSON),
			AdditionalNotes:       input.AdditionalNotes,
		}
		if err := tx.Create(lifestyle).Error; err != nil {
			return err
		}

		result = domain.CreateTripResult{
			Room:      room,
			Trip:      trip,
			Member:    member,
			Lifestyle: lifestyle,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Analyze lifestyle and save recommendations as suggestions
	places, err := s.lifestyleSvc.AnalyzeLifestyle(ctx, result.Lifestyle.LifestyleID)
	if err != nil {
		log.Printf("[CreateTrip] AnalyzeLifestyle failed (lifestyle_id=%d): %v", result.Lifestyle.LifestyleID, err)
	} else {
		suggestions := make([]domain.TripSchedule, 0, len(places))
		for _, p := range places {
			suggestions = append(suggestions, domain.TripSchedule{
				TripID:        result.Trip.TripID,
				DayNumber:     0,
				SequenceOrder: 0,
				PlaceName:     p.Name,
				PlaceID:       "",
				Latitude:      p.Latitude,
				Longitude:     p.Longitude,
				Type:          p.Category,
			})
		}
		if len(suggestions) > 0 {
			if err := s.db.WithContext(ctx).Create(&suggestions).Error; err != nil {
				log.Printf("[CreateTrip] failed to save suggestions: %v", err)
			} else {
				result.Suggestions = suggestions
			}
		}
	}

	return &result, nil
}

func (s *tripService) CreateTripSchedule(ctx context.Context, inputs []domain.CreateTripScheduleInput) ([]domain.TripSchedule, error) {
	if len(inputs) == 0 {
		return nil, errors.New("items must not be empty")
	}

	schedules := make([]domain.TripSchedule, 0, len(inputs))
	for _, inp := range inputs {
		startTime, err := time.Parse("15:04", inp.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time %q: must be HH:MM", inp.StartTime)
		}
		endTime, err := time.Parse("15:04", inp.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time %q: must be HH:MM", inp.EndTime)
		}

		schedules = append(schedules, domain.TripSchedule{
			TripID:        inp.TripID,
			DayNumber:     inp.DayNumber,
			SequenceOrder: inp.SequenceOrder,
			PlaceName:     inp.PlaceName,
			PlaceID:       inp.PlaceID,
			Latitude:      inp.Latitude,
			Longitude:     inp.Longitude,
			StartTime:     startTime,
			EndTime:       endTime,
			Type:          inp.Type,
		})
	}

	if err := s.db.WithContext(ctx).Create(&schedules).Error; err != nil {
		return nil, err
	}

	return schedules, nil
}

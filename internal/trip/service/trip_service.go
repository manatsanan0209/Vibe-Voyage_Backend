package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type tripService struct {
	repo         domain.TripRepository
	lifestyleSvc domain.UserLifestyleService
}

func NewTripService(repo domain.TripRepository, lifestyleSvc domain.UserLifestyleService) domain.TripService {
	return &tripService{repo: repo, lifestyleSvc: lifestyleSvc}
}

func (s *tripService) GetTripSchedule(ctx context.Context, tripID uint) (*domain.GetTripScheduleResult, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	schedules, err := s.repo.GetSchedulesByTripID(ctx, tripID)
	if err != nil {
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
		Trip:        trip,
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

	createdResult, err := s.repo.CreateTripBundle(
		ctx,
		userID,
		input,
		string(preferredDestJSON),
		string(travelVibesJSON),
		string(prioritiesJSON),
		string(foodVibesJSON),
	)
	if err != nil {
		return nil, err
	}
	result = *createdResult

	// Analyze lifestyle and save recommendations as suggestions
	places, err := s.lifestyleSvc.AnalyzeLifestyle(ctx, result.Lifestyle.LifestyleID)
	if err != nil {
		log.Printf("[CreateTrip] AnalyzeLifestyle failed (lifestyle_id=%d): %v", result.Lifestyle.LifestyleID, err)
	} else {
		aiSuggestions := make([]domain.TripSchedule, 0, len(places))
		for _, p := range places {
			aiSuggestions = append(aiSuggestions, domain.TripSchedule{
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
		if len(aiSuggestions) > 0 {
			scheduled := schedulePlaces(aiSuggestions, result.Trip.StartDate, result.Trip.EndDate)
			if err := s.repo.CreateSchedules(ctx, scheduled); err != nil {
				log.Printf("[CreateTrip] failed to save suggestions: %v", err)
			} else {
				result.Suggestions = append(result.Suggestions, scheduled...)
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

	if err := s.repo.CreateSchedules(ctx, schedules); err != nil {
		return nil, err
	}

	return schedules, nil
}

// haversine returns the great-circle distance in km between two lat/lng points.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// nearestNeighborOrder reorders places so each next place is the closest
// unvisited one from the current position (greedy nearest-neighbor).
func nearestNeighborOrder(places []domain.TripSchedule) []domain.TripSchedule {
	if len(places) == 0 {
		return places
	}
	visited := make([]bool, len(places))
	result := make([]domain.TripSchedule, 0, len(places))
	current := 0
	visited[current] = true
	result = append(result, places[current])

	for len(result) < len(places) {
		minDist := math.MaxFloat64
		nearest := -1
		for j := range places {
			if visited[j] {
				continue
			}
			d := haversine(
				places[current].Latitude, places[current].Longitude,
				places[j].Latitude, places[j].Longitude,
			)
			if d < minDist {
				minDist = d
				nearest = j
			}
		}
		if nearest == -1 {
			break
		}
		visited[nearest] = true
		result = append(result, places[nearest])
		current = nearest
	}
	return result
}

// schedulePlaces assigns DayNumber and SequenceOrder to places by:
// 1. Ordering them with nearest-neighbor (geographically close places together)
// 2. Distributing evenly across the trip days (max 4 places per day)
func schedulePlaces(places []domain.TripSchedule, startDate, endDate time.Time) []domain.TripSchedule {
	if len(places) == 0 {
		return places
	}

	const maxPlacesPerDay = 4

	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	if totalDays < 1 {
		totalDays = 1
	}

	ordered := nearestNeighborOrder(places)

	// Truncate to at most 4 places per day
	maxPlaces := totalDays * maxPlacesPerDay
	if len(ordered) > maxPlaces {
		ordered = ordered[:maxPlaces]
	}

	// ceiling division: spread extras into earlier days
	placesPerDay := (len(ordered) + totalDays - 1) / totalDays
	if placesPerDay < 1 {
		placesPerDay = 1
	}
	if placesPerDay > maxPlacesPerDay {
		placesPerDay = maxPlacesPerDay
	}

	for i := range ordered {
		day := i/placesPerDay + 1
		if day > totalDays {
			day = totalDays
		}
		seq := i%placesPerDay + 1
		ordered[i].DayNumber = day
		ordered[i].SequenceOrder = seq
	}

	return ordered
}

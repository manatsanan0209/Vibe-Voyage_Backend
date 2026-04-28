package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type tripService struct {
	repo             domain.TripRepository
	restaurantRepo   domain.RestaurantRepository
	attractionRepo   domain.AttractionRepository
	lifestyleSvc     domain.UserLifestyleService
	roomSvc          domain.RoomService
	notifSvc         domain.NotificationService
	analyzeSemaphore chan struct{}
	analyzeTimeout   time.Duration
	rescheduleLocks  sync.Map
}

func NewTripService(
	repo domain.TripRepository,
	restaurantRepo domain.RestaurantRepository,
	attractionRepo domain.AttractionRepository,
	lifestyleSvc domain.UserLifestyleService,
	roomSvc domain.RoomService,
	notifSvc domain.NotificationService,
) domain.TripService {
	return &tripService{
		repo:             repo,
		restaurantRepo:   restaurantRepo,
		attractionRepo:   attractionRepo,
		lifestyleSvc:     lifestyleSvc,
		roomSvc:          roomSvc,
		notifSvc:         notifSvc,
		analyzeSemaphore: make(chan struct{}, 5),
		analyzeTimeout:   45 * time.Second,
	}
}

var foodVibeToFoodTypes = map[string][]string{
	"thai_food":         {"Thai Food", "Thai northern food", "Thai-Isan", "Noodle", "Seafood"},
	"asian_food":        {"Asian Fusion", "Chinese", "Japanese", "Indonesian", "Indian", "Vietnam Food", "Vietnamese food", "Noodle"},
	"thai_local_food":   {"Thai Food", "Thai northern food", "Thai-Isan", "Noodle"},
	"halal_muslim_food": {"Halal and Kosher", "Muslim Food"},
	"western_food":      {"American", "Barbecue", "Cajun/Creole/Soul", "European", "Fast Food/Snack", "French/Bistro", "Greek", "Italian"},
	"cafe_dessert":      {"Barcirerie/Cafe", "Coffee Shop", "Desserts", "Icecream"},
}

func mapFoodVibesToFoodTypes(foodVibes []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, v := range foodVibes {
		for _, t := range foodVibeToFoodTypes[v] {
			if !seen[t] {
				seen[t] = true
				result = append(result, t)
			}
		}
	}
	return result
}

func (s *tripService) JoinTripByInviteCode(ctx context.Context, userID uint, inviteCode string) (*domain.JoinTripByInviteCodeResult, error) {
	member, err := s.roomSvc.JoinByInviteCode(ctx, userID, inviteCode)
	if err != nil {
		return nil, err
	}

	trip, err := s.repo.GetByRoomID(ctx, member.RoomID)
	if err != nil {
		return nil, err
	}

	return &domain.JoinTripByInviteCodeResult{
		Trip:   trip,
		Member: member,
	}, nil
}

func (s *tripService) GetTripSchedule(ctx context.Context, userID, tripID uint) (*domain.GetTripScheduleResult, error) {
	allowed, err := s.repo.IsUserInTripRoom(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

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

	s.enqueueAnalyzeAndSaveSuggestions(
		result.Lifestyle.LifestyleID,
		result.Trip.TripID,
		userID,
		result.Trip.StartDate,
		result.Trip.EndDate,
		input.FoodVibes,
		result.Trip.DestinationID,
	)

	return &result, nil
}

func (s *tripService) enqueueAnalyzeAndSaveSuggestions(
	lifestyleID, tripID, userID uint,
	startDate, endDate time.Time,
	foodVibes []string,
	destinationID string,
) {
	go func() {
		s.analyzeSemaphore <- struct{}{}
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[CreateTrip] async analyze panic (lifestyle_id=%d): %v", lifestyleID, r)
			}
			<-s.analyzeSemaphore
		}()

		asyncCtx, cancel := context.WithTimeout(context.Background(), s.analyzeTimeout)
		defer cancel()

		places, err := s.lifestyleSvc.AnalyzeLifestyle(asyncCtx, lifestyleID)
		if err != nil {
			log.Printf("[CreateTrip] async analyze failed (lifestyle_id=%d): %v", lifestyleID, err)
			return
		}

		aiSuggestions := make([]domain.TripSchedule, 0, len(places))
		for _, p := range places {
			placeID := s.resolveAttractionID(asyncCtx, p.Name, p.Latitude, p.Longitude)
			aiSuggestions = append(aiSuggestions, domain.TripSchedule{
				TripID:        tripID,
				DayNumber:     0,
				SequenceOrder: 0,
				PlaceName:     p.Name,
				PlaceID:       placeID,
				Latitude:      p.Latitude,
				Longitude:     p.Longitude,
				Type:          "attraction",
			})
		}

		if len(aiSuggestions) == 0 {
			log.Printf("[CreateTrip] async analyze completed with no suggestions (lifestyle_id=%d)", lifestyleID)
			return
		}

		scheduled, unscheduled := schedulePlaces(aiSuggestions, startDate, endDate)
		for i := range scheduled {
			scheduled[i].SequenceOrder = shiftAttractionSeq(scheduled[i].SequenceOrder)
		}

		restaurants := s.buildMealSchedules(asyncCtx, tripID, scheduled, foodVibes, destinationID)
		allSchedules := append(scheduled, restaurants...)
		allSchedules = append(allSchedules, unscheduled...)

		if err := s.repo.CreateSchedules(asyncCtx, allSchedules); err != nil {
			log.Printf("[CreateTrip] async save suggestions failed (trip_id=%d, lifestyle_id=%d): %v", tripID, lifestyleID, err)
			return
		}

		log.Printf("[CreateTrip] async suggestions saved (trip_id=%d, lifestyle_id=%d, attractions=%d, meals=%d)", tripID, lifestyleID, len(scheduled), len(restaurants))

		refType := "trip"
		if err := s.notifSvc.Notify(context.Background(), userID, domain.NotifTypeTripCreated, "Trip is ready!", "Your trip schedule has been created.", &tripID, &refType); err != nil {
			log.Printf("[Notification] trip_created failed (user_id=%d): %v", userID, err)
		}
	}()
}

// shiftAttractionSeq reserves seq 1/4/7 for meals by shifting attractions:
// 1 -> 2, 2 -> 3, 3 -> 5, 4 -> 6.
func shiftAttractionSeq(oldSeq int) int {
	if oldSeq <= 2 {
		return oldSeq + 1
	}
	return oldSeq + 2
}

// buildMealSchedules picks breakfast/lunch/dinner restaurants per day, anchored
// on attractions already scheduled. Restaurants are filtered by food_vibes when
// possible; falls back to nearest-by-province if no match is found. Restaurants
// used on earlier days are excluded from later days when candidates allow it.
func (s *tripService) buildMealSchedules(
	ctx context.Context,
	tripID uint,
	attractions []domain.TripSchedule,
	foodVibes []string,
	destinationID string,
) []domain.TripSchedule {
	if len(attractions) == 0 {
		return nil
	}

	foodTypes := mapFoodVibesToFoodTypes(foodVibes)

	candidates, err := s.restaurantRepo.ListNearbyByFoodTypes(ctx, destinationID, foodTypes, 200)
	if err != nil {
		log.Printf("[CreateTrip] fetch restaurants failed (trip_id=%d): %v", tripID, err)
		return nil
	}
	if len(candidates) == 0 && len(foodTypes) > 0 {
		candidates, err = s.restaurantRepo.ListNearbyByFoodTypes(ctx, destinationID, nil, 200)
		if err != nil {
			log.Printf("[CreateTrip] fetch fallback restaurants failed (trip_id=%d): %v", tripID, err)
			return nil
		}
	}
	if len(candidates) == 0 {
		log.Printf("[CreateTrip] no restaurants available for province %q (trip_id=%d)", destinationID, tripID)
		return nil
	}

	byDay := map[int][]domain.TripSchedule{}
	dayOrder := []int{}
	for _, a := range attractions {
		if _, ok := byDay[a.DayNumber]; !ok {
			dayOrder = append(dayOrder, a.DayNumber)
		}
		byDay[a.DayNumber] = append(byDay[a.DayNumber], a)
	}

	type meal struct {
		seq       int
		anchorIdx int
	}
	mealSlots := []meal{
		{seq: 1, anchorIdx: 0},
		{seq: 4, anchorIdx: 1},
		{seq: 7, anchorIdx: 3},
	}

	excludeIDs := map[string]bool{}
	result := []domain.TripSchedule{}

	for _, day := range dayOrder {
		dayAttrs := byDay[day]
		// Sort within a day by seq so anchor lookups are stable.
		sortBySeq(dayAttrs)

		for _, m := range mealSlots {
			anchorIdx := m.anchorIdx
			// Fallback: if the preferred anchor is missing (fewer attractions), use the last available.
			if anchorIdx >= len(dayAttrs) {
				if len(dayAttrs) == 0 {
					continue
				}
				anchorIdx = len(dayAttrs) - 1
			}
			anchor := dayAttrs[anchorIdx]

			picked := findNearestRestaurant(anchor, candidates, excludeIDs)
			if picked == nil {
				// All candidates exhausted — allow reuse instead of skipping meals.
				picked = findNearestRestaurant(anchor, candidates, nil)
			}
			if picked == nil {
				continue
			}
			excludeIDs[picked.ID] = true
			result = append(result, restaurantToSchedule(picked, tripID, day, m.seq))
		}
	}

	return result
}

// resolveAttractionID looks up the attraction ID from the DB by name,
// then picks the candidate with the shortest haversine distance to the given coords.
// Returns empty string if no match is found.
func (s *tripService) resolveAttractionID(ctx context.Context, name string, lat, lon float64) string {
	candidates, err := s.attractionRepo.GetByName(ctx, name)
	if err != nil || len(candidates) == 0 {
		return ""
	}
	best := candidates[0]
	minDist := haversine(lat, lon, best.Latitude, best.Longitude)
	for _, c := range candidates[1:] {
		if d := haversine(lat, lon, c.Latitude, c.Longitude); d < minDist {
			minDist = d
			best = c
		}
	}
	return best.ID
}

func sortBySeq(items []domain.TripSchedule) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j-1].SequenceOrder > items[j].SequenceOrder; j-- {
			items[j-1], items[j] = items[j], items[j-1]
		}
	}
}

func findNearestRestaurant(
	anchor domain.TripSchedule,
	candidates []*domain.Restaurant,
	excludeIDs map[string]bool,
) *domain.Restaurant {
	var best *domain.Restaurant
	minDist := math.MaxFloat64
	for _, c := range candidates {
		if c == nil {
			continue
		}
		if excludeIDs != nil && excludeIDs[c.ID] {
			continue
		}
		if c.Latitude == 0 && c.Longitude == 0 {
			continue
		}
		d := haversine(anchor.Latitude, anchor.Longitude, c.Latitude, c.Longitude)
		if d < minDist {
			minDist = d
			best = c
		}
	}
	return best
}

func restaurantToSchedule(
	r *domain.Restaurant,
	tripID uint,
	dayNumber, sequenceOrder int,
) domain.TripSchedule {
	name := r.NameTH
	if name == "" {
		name = r.NameEN
	}
	return domain.TripSchedule{
		TripID:        tripID,
		DayNumber:     dayNumber,
		SequenceOrder: sequenceOrder,
		PlaceName:     name,
		PlaceID:       r.ID,
		Latitude:      r.Latitude,
		Longitude:     r.Longitude,
		Type:          "restaurant",
	}
}

func (s *tripService) CreateTripSchedule(ctx context.Context, inputs []domain.CreateTripScheduleInput) ([]domain.TripSchedule, error) {
	if len(inputs) == 0 {
		return nil, errors.New("items must not be empty")
	}

	schedules := make([]domain.TripSchedule, 0, len(inputs))
	for _, inp := range inputs {
		if inp.PlaceID == "" {
			return nil, fmt.Errorf("place_id is required for item %q", inp.PlaceName)
		}
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

func (s *tripService) ReplaceTripSchedule(ctx context.Context, userID, tripID uint, inputs []domain.CreateTripScheduleInput) ([]domain.TripSchedule, error) {
	role, exists, err := s.repo.GetUserRoleInTripRoom(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrForbidden
	}

	if role != domain.RoleOwner && role != domain.RoleMember {
		return nil, domain.ErrForbidden
	}

	schedules := make([]domain.TripSchedule, 0, len(inputs))
	for _, inp := range inputs {
		if inp.PlaceID == "" {
			return nil, fmt.Errorf("place_id is required for item %q", inp.PlaceName)
		}
		startTime, err := time.Parse("15:04", inp.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time %q: must be HH:MM", inp.StartTime)
		}
		endTime, err := time.Parse("15:04", inp.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time %q: must be HH:MM", inp.EndTime)
		}

		schedules = append(schedules, domain.TripSchedule{
			TripScheduleID: inp.TripScheduleID,
			TripID:         tripID,
			DayNumber:      inp.DayNumber,
			SequenceOrder:  inp.SequenceOrder,
			PlaceName:      inp.PlaceName,
			PlaceID:        inp.PlaceID,
			Latitude:       inp.Latitude,
			Longitude:      inp.Longitude,
			StartTime:      startTime,
			EndTime:        endTime,
			Type:           inp.Type,
		})
	}

	if err := s.repo.ReplaceSchedulesByTripID(ctx, tripID, schedules); err != nil {
		return nil, err
	}

	go func() {
		trip, err := s.repo.GetByID(context.Background(), tripID)
		if err != nil {
			return
		}
		members, err := s.roomSvc.GetMembersByRoomID(context.Background(), trip.RoomID)
		if err != nil {
			return
		}
		refType := "trip"
		for _, m := range members {
			if m.UserID == userID {
				continue
			}
			if err := s.notifSvc.Notify(context.Background(), m.UserID, domain.NotifTypeScheduleUpdated, "Schedule updated", "The trip schedule has been updated.", &tripID, &refType); err != nil {
				log.Printf("[Notification] schedule_updated failed (user_id=%d): %v", m.UserID, err)
			}
		}
	}()

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
func schedulePlaces(places []domain.TripSchedule, startDate, endDate time.Time) (scheduled, unscheduled []domain.TripSchedule) {
	if len(places) == 0 {
		return places, nil
	}

	const maxPlacesPerDay = 4

	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	if totalDays < 1 {
		totalDays = 1
	}

	ordered := nearestNeighborOrder(places)

	maxPlaces := totalDays * maxPlacesPerDay
	if len(ordered) > maxPlaces {
		unscheduled = ordered[maxPlaces:]
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

	return ordered, unscheduled
}

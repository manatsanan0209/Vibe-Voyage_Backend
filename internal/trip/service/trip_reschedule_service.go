package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

const (
	ownerGainScore            = 1.00
	deferredBonusScore        = 0.20
	maxDistancePerHopKM       = 25.0
	maxDistanceFromCentroidKM = 35.0
)

type memberRescheduleState struct {
	UserID            uint
	Username          string
	Candidates        []rankedCandidate
	CategoryRank      map[string]int
	Score             float64
	TimesServed       int
	DeferredCount     int
	LastSelectedRound int
}

type rankedCandidate struct {
	PlaceID   string
	Name      string
	Category  string
	Latitude  float64
	Longitude float64
}

type geoPoint struct {
	Latitude  float64
	Longitude float64
}

type fairnessRunSnapshot struct {
	AlgorithmVersion string                             `json:"algorithm_version"`
	RoundCount       int                                `json:"round_count"`
	SelectedPlaceIDs []string                           `json:"selected_place_ids"`
	Members          []domain.RescheduleTripMemberScore `json:"members"`
	GeneratedAt      string                             `json:"generated_at"`
}

func (s *tripService) RescheduleTrip(ctx context.Context, userID, tripID uint) (*domain.RescheduleTripResult, error) {
	role, exists, err := s.repo.GetUserRoleInTripRoom(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	if !exists || role != domain.RoleOwner {
		return nil, domain.ErrForbidden
	}

	releaseLock, locked := s.tryAcquireRescheduleLock(tripID)
	if !locked {
		return nil, domain.ErrRescheduleConcurrentModification
	}
	defer releaseLock()

	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	existingSchedules, err := s.repo.GetSchedulesByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	members, err := s.roomSvc.GetMembersByRoomID(ctx, trip.RoomID)
	if err != nil {
		return nil, err
	}

	states := make(map[uint]*memberRescheduleState)
	notReadyMembers := make([]domain.RescheduleNotReadyMember, 0)
	namesToResolve := make([]string, 0)
	foodVibesSet := map[string]bool{}

	for _, member := range members {
		if member.Role != domain.RoleOwner && member.Role != domain.RoleMember {
			continue
		}

		lifestyle, err := s.lifestyleSvc.GetLifestyle(ctx, member.UserID, trip.RoomID)
		if err != nil {
			if errors.Is(err, domain.ErrLifestyleNotFound) {
				notReadyMembers = append(notReadyMembers, domain.RescheduleNotReadyMember{
					UserID:      member.UserID,
					Username:    member.User.Username,
					LifestyleID: nil,
				})
				continue
			}
			return nil, err
		}

		lifestyleID := lifestyle.LifestyleID
		if !domain.IsStructuredLifestyleValid(lifestyle.StructuredLifestyle) {
			notReadyMembers = append(notReadyMembers, domain.RescheduleNotReadyMember{
				UserID:      member.UserID,
				Username:    member.User.Username,
				LifestyleID: &lifestyleID,
			})
			continue
		}

		candidates, err := parseStructuredLifestylePlaces(*lifestyle.StructuredLifestyle)
		if err != nil {
			notReadyMembers = append(notReadyMembers, domain.RescheduleNotReadyMember{
				UserID:      member.UserID,
				Username:    member.User.Username,
				LifestyleID: &lifestyleID,
			})
			continue
		}

		candidates = dedupeCandidatesByPlaceID(candidates)
		for _, candidate := range candidates {
			if candidate.PlaceID == "" {
				namesToResolve = append(namesToResolve, candidate.Name)
			}
		}

		collectFoodVibes(foodVibesSet, lifestyle.FoodVibes)

		states[member.UserID] = &memberRescheduleState{
			UserID:            member.UserID,
			Username:          member.User.Username,
			Candidates:        candidates,
			CategoryRank:      map[string]int{},
			LastSelectedRound: -1,
		}
	}

	if len(notReadyMembers) > 0 {
		sort.Slice(notReadyMembers, func(i, j int) bool {
			return notReadyMembers[i].UserID < notReadyMembers[j].UserID
		})
		return nil, &domain.RescheduleAnalysisNotReadyError{NotReadyMembers: notReadyMembers}
	}

	attractionsByName := map[string][]domain.Attraction{}
	if len(namesToResolve) > 0 {
		attractionsByName, err = s.repo.GetAttractionsByNames(ctx, namesToResolve)
		if err != nil {
			return nil, err
		}
	}

	for _, state := range states {
		resolved := make([]rankedCandidate, 0, len(state.Candidates))
		for _, candidate := range state.Candidates {
			if candidate.PlaceID == "" {
				candidate.PlaceID = resolveAttractionIDFromCandidates(candidate, attractionsByName[strings.ToLower(strings.TrimSpace(candidate.Name))])
			}
			if candidate.PlaceID == "" {
				continue
			}
			resolved = append(resolved, candidate)
		}
		state.Candidates = resolved
		state.CategoryRank = buildCategoryRank(resolved)
	}

	if len(states) == 0 {
		return nil, errors.New("no analyzed lifestyles available for rescheduling")
	}

	orderedCandidates, rounds := runFairnessSelection(states)
	if len(orderedCandidates) == 0 {
		return nil, errors.New("unable to build schedule from current constraints")
	}

	attractions := make([]domain.TripSchedule, 0, len(orderedCandidates))
	selectedPlaceIDs := make([]string, 0, len(orderedCandidates))
	for _, item := range orderedCandidates {
		selectedPlaceIDs = append(selectedPlaceIDs, item.PlaceID)
		attractions = append(attractions, domain.TripSchedule{
			TripID:        tripID,
			DayNumber:     0,
			SequenceOrder: 0,
			PlaceName:     item.Name,
			PlaceID:       item.PlaceID,
			Latitude:      item.Latitude,
			Longitude:     item.Longitude,
			Type:          "attraction",
		})
	}

	scheduled, unscheduled := schedulePlaces(attractions, trip.StartDate, trip.EndDate)
	for i := range scheduled {
		scheduled[i].SequenceOrder = shiftAttractionSeq(scheduled[i].SequenceOrder)
	}

	foodVibes := setToSortedSlice(foodVibesSet)
	restaurants := s.buildMealSchedules(ctx, tripID, scheduled, foodVibes, trip.DestinationID)

	allSchedules := make([]domain.TripSchedule, 0, len(scheduled)+len(restaurants)+len(unscheduled))
	allSchedules = append(allSchedules, scheduled...)
	allSchedules = append(allSchedules, restaurants...)
	allSchedules = append(allSchedules, unscheduled...)

	preservedSuggestions := collectPreservedSuggestions(existingSchedules, allSchedules)
	allSchedules = append(allSchedules, preservedSuggestions...)

	scoreboard := buildScoreboard(states)
	snapshotBytes, err := json.Marshal(fairnessRunSnapshot{
		AlgorithmVersion: "fairness_v1",
		RoundCount:       rounds,
		SelectedPlaceIDs: selectedPlaceIDs,
		Members:          scoreboard,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build reschedule snapshot: %w", err)
	}

	if err := s.repo.ReplaceScheduleAndSnapshot(ctx, tripID, allSchedules, string(snapshotBytes)); err != nil {
		return nil, err
	}

	go s.notifyScheduleUpdatedMembers(members, userID, tripID)

	return &domain.RescheduleTripResult{
		TripID:           tripID,
		ScheduledCount:   len(scheduled) + len(restaurants),
		SuggestionsCount: len(unscheduled) + len(preservedSuggestions),
		RoundCount:       rounds,
		SelectedPlaceIDs: selectedPlaceIDs,
		Scoreboard:       scoreboard,
	}, nil
}

func runFairnessSelection(states map[uint]*memberRescheduleState) ([]rankedCandidate, int) {
	selected := make([]rankedCandidate, 0)
	selectedByPlaceID := map[string]bool{}
	centroid := calculateCandidatesCentroid(states)
	roundCount := 0
	failedTurns := 0

	for {
		activeIDs := activeMemberIDs(states, selectedByPlaceID)
		if len(activeIDs) == 0 {
			break
		}

		nextMemberID := pickNextMemberID(activeIDs, states)
		memberState := states[nextMemberID]
		roundCount++

		candidate, found := findBestCandidateForMember(memberState, selected, selectedByPlaceID, centroid)
		if !found {
			memberState.DeferredCount++
			failedTurns++
			if failedTurns >= len(activeIDs) {
				break
			}
			continue
		}

		failedTurns = 0
		selected = append(selected, candidate)
		selectedByPlaceID[candidate.PlaceID] = true

		memberState.Score += ownerGainScore
		memberState.TimesServed++
		memberState.DeferredCount = 0
		memberState.LastSelectedRound = roundCount

		sharedGains := map[uint]float64{}
		if candidate.Category != "" {
			normalizedCategory := normalizeCategory(candidate.Category)
			for userID, state := range states {
				if userID == memberState.UserID {
					continue
				}
				rank, ok := state.CategoryRank[normalizedCategory]
				if !ok {
					continue
				}
				gain := sharedGainByCategoryRank(rank)
				if gain > sharedGains[userID] {
					sharedGains[userID] = gain
				}
			}
		}

		for userID, gain := range sharedGains {
			states[userID].Score += gain
		}
	}

	return selected, roundCount
}

func activeMemberIDs(states map[uint]*memberRescheduleState, selectedByPlaceID map[string]bool) []uint {
	ids := make([]uint, 0, len(states))
	for userID, state := range states {
		if hasUnusedCandidate(state, selectedByPlaceID) {
			ids = append(ids, userID)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	return ids
}

func hasUnusedCandidate(state *memberRescheduleState, selectedByPlaceID map[string]bool) bool {
	for _, candidate := range state.Candidates {
		if candidate.PlaceID == "" {
			continue
		}
		if !selectedByPlaceID[candidate.PlaceID] {
			return true
		}
	}
	return false
}

func pickNextMemberID(activeIDs []uint, states map[uint]*memberRescheduleState) uint {
	best := activeIDs[0]
	for _, userID := range activeIDs[1:] {
		if shouldPrioritize(states[userID], states[best], userID, best) {
			best = userID
		}
	}
	return best
}

func shouldPrioritize(a, b *memberRescheduleState, aUserID, bUserID uint) bool {
	aEffective := a.Score + float64(a.DeferredCount)*deferredBonusScore
	bEffective := b.Score + float64(b.DeferredCount)*deferredBonusScore
	if aEffective != bEffective {
		return aEffective < bEffective
	}
	if a.TimesServed != b.TimesServed {
		return a.TimesServed < b.TimesServed
	}
	if a.LastSelectedRound != b.LastSelectedRound {
		return a.LastSelectedRound < b.LastSelectedRound
	}
	return aUserID < bUserID
}

func findBestCandidateForMember(
	state *memberRescheduleState,
	selected []rankedCandidate,
	selectedByPlaceID map[string]bool,
	centroid *geoPoint,
) (rankedCandidate, bool) {
	for _, candidate := range state.Candidates {
		if candidate.PlaceID == "" || selectedByPlaceID[candidate.PlaceID] {
			continue
		}
		if !passesDistanceConstraint(candidate, selected, centroid) {
			continue
		}
		return candidate, true
	}
	return rankedCandidate{}, false
}

func passesDistanceConstraint(candidate rankedCandidate, selected []rankedCandidate, centroid *geoPoint) bool {
	if !hasCoordinates(candidate.Latitude, candidate.Longitude) {
		return true
	}

	if len(selected) == 0 {
		if centroid == nil {
			return true
		}
		return haversine(candidate.Latitude, candidate.Longitude, centroid.Latitude, centroid.Longitude) <= maxDistanceFromCentroidKM
	}

	last := selected[len(selected)-1]
	if !hasCoordinates(last.Latitude, last.Longitude) {
		return true
	}

	return haversine(candidate.Latitude, candidate.Longitude, last.Latitude, last.Longitude) <= maxDistancePerHopKM
}

func hasCoordinates(lat, lon float64) bool {
	return lat != 0 || lon != 0
}

func calculateCandidatesCentroid(states map[uint]*memberRescheduleState) *geoPoint {
	var latSum, lonSum float64
	var count int
	for _, state := range states {
		for _, candidate := range state.Candidates {
			if !hasCoordinates(candidate.Latitude, candidate.Longitude) {
				continue
			}
			latSum += candidate.Latitude
			lonSum += candidate.Longitude
			count++
		}
	}
	if count == 0 {
		return nil
	}
	return &geoPoint{
		Latitude:  latSum / float64(count),
		Longitude: lonSum / float64(count),
	}
}

func sharedGainByCategoryRank(rank int) float64 {
	switch rank {
	case 1:
		return 0.50
	case 2:
		return 0.35
	case 3:
		return 0.25
	default:
		return 0.15
	}
}

func buildCategoryRank(candidates []rankedCandidate) map[string]int {
	categoryOrder := map[string]int{}
	nextRank := 1

	for _, item := range candidates {
		category := normalizeCategory(item.Category)
		if category == "" {
			continue
		}
		if _, exists := categoryOrder[category]; exists {
			continue
		}
		categoryOrder[category] = nextRank
		nextRank++
	}

	return categoryOrder
}

func normalizeCategory(category string) string {
	return strings.ToLower(strings.TrimSpace(category))
}

func dedupeCandidatesByPlaceID(candidates []rankedCandidate) []rankedCandidate {
	seen := map[string]bool{}
	result := make([]rankedCandidate, 0, len(candidates))

	for _, candidate := range candidates {
		dedupeKey := strings.TrimSpace(candidate.PlaceID)
		if dedupeKey == "" {
			dedupeKey = fallbackPlaceKey(candidate)
		}
		if seen[dedupeKey] {
			continue
		}
		seen[dedupeKey] = true
		result = append(result, candidate)
	}

	return result
}

func fallbackPlaceKey(candidate rankedCandidate) string {
	return strings.ToLower(strings.TrimSpace(candidate.Name)) + "|" +
		strconv.FormatFloat(candidate.Latitude, 'f', 6, 64) + "|" +
		strconv.FormatFloat(candidate.Longitude, 'f', 6, 64)
}

func parseStructuredLifestylePlaces(raw string) ([]rankedCandidate, error) {
	var payload interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}

	result := make([]rankedCandidate, 0)
	collectPlacesFromPayload(payload, &result)
	return result, nil
}

func collectPlacesFromPayload(node interface{}, result *[]rankedCandidate) {
	switch value := node.(type) {
	case []interface{}:
		for _, item := range value {
			collectPlacesFromPayload(item, result)
		}
	case map[string]interface{}:
		if candidate, ok := toRankedCandidate(value); ok {
			*result = append(*result, candidate)
			return
		}
		for _, item := range value {
			collectPlacesFromPayload(item, result)
		}
	}
}

func toRankedCandidate(node map[string]interface{}) (rankedCandidate, bool) {
	placeID := getStringValue(node, "place_id", "id", "destination_id")
	name := getStringValue(node, "name", "place_name", "destination_name", "title")
	category := getStringValue(node, "category", "type", "place_type")
	lat, hasLat := getNumberValue(node, "latitude", "lat")
	lon, hasLon := getNumberValue(node, "longitude", "lon", "lng")

	if name == "" {
		return rankedCandidate{}, false
	}
	if placeID == "" && category == "" && !(hasLat && hasLon) {
		return rankedCandidate{}, false
	}

	return rankedCandidate{
		PlaceID:   placeID,
		Name:      name,
		Category:  category,
		Latitude:  lat,
		Longitude: lon,
	}, true
}

func getStringValue(node map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := node[key]
		if !ok || value == nil {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func getNumberValue(node map[string]interface{}, keys ...string) (float64, bool) {
	for _, key := range keys {
		value, ok := node[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return typed, true
		case float32:
			return float64(typed), true
		case int:
			return float64(typed), true
		case int64:
			return float64(typed), true
		case json.Number:
			if parsed, err := typed.Float64(); err == nil {
				return parsed, true
			}
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func collectFoodVibes(out map[string]bool, raw string) {
	if raw == "" {
		return
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return
	}
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized != "" {
			out[normalized] = true
		}
	}
}

func (s *tripService) tryAcquireRescheduleLock(tripID uint) (func(), bool) {
	lockValue, _ := s.rescheduleLocks.LoadOrStore(tripID, make(chan struct{}, 1))
	lock := lockValue.(chan struct{})

	select {
	case lock <- struct{}{}:
		return func() {
			<-lock
		}, true
	default:
		return nil, false
	}
}

func resolveAttractionIDFromCandidates(candidate rankedCandidate, attractions []domain.Attraction) string {
	if len(attractions) == 0 {
		return ""
	}

	best := attractions[0]
	minDist := haversine(candidate.Latitude, candidate.Longitude, best.Latitude, best.Longitude)

	for _, attraction := range attractions[1:] {
		distance := haversine(candidate.Latitude, candidate.Longitude, attraction.Latitude, attraction.Longitude)
		if distance < minDist {
			minDist = distance
			best = attraction
		}
	}

	return best.ID
}

func collectPreservedSuggestions(existingSchedules, regeneratedSchedules []domain.TripSchedule) []domain.TripSchedule {
	keptPlaceIDs := make(map[string]bool, len(regeneratedSchedules))
	for _, item := range regeneratedSchedules {
		placeID := normalizeSchedulePlaceID(item.PlaceID)
		if placeID == "" {
			continue
		}
		keptPlaceIDs[placeID] = true
	}

	preserved := make([]domain.TripSchedule, 0)
	preservedSeen := map[string]bool{}

	for _, item := range existingSchedules {
		if !shouldPreserveScheduleItem(item, keptPlaceIDs) {
			continue
		}

		key := preserveDedupKey(item)
		if preservedSeen[key] {
			continue
		}
		preservedSeen[key] = true

		item.TripScheduleID = 0
		item.DayNumber = 0
		item.SequenceOrder = 0
		preserved = append(preserved, item)
	}

	return preserved
}

func shouldPreserveScheduleItem(item domain.TripSchedule, keptPlaceIDs map[string]bool) bool {
	if strings.EqualFold(strings.TrimSpace(item.Type), "restaurant") {
		return false
	}

	placeID := normalizeSchedulePlaceID(item.PlaceID)
	if placeID == "" {
		return false
	}

	return !keptPlaceIDs[placeID]
}

func normalizeSchedulePlaceID(placeID string) string {
	return strings.ToLower(strings.TrimSpace(placeID))
}

func preserveDedupKey(item domain.TripSchedule) string {
	return normalizeSchedulePlaceID(item.PlaceID) + "|" + strings.ToLower(strings.TrimSpace(item.Type))
}

func setToSortedSlice(items map[string]bool) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	for item := range items {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func buildScoreboard(states map[uint]*memberRescheduleState) []domain.RescheduleTripMemberScore {
	ids := make([]uint, 0, len(states))
	for userID := range states {
		ids = append(ids, userID)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	scoreboard := make([]domain.RescheduleTripMemberScore, 0, len(ids))
	for _, userID := range ids {
		state := states[userID]
		scoreboard = append(scoreboard, domain.RescheduleTripMemberScore{
			UserID:         state.UserID,
			Username:       state.Username,
			Score:          state.Score,
			EffectiveScore: state.Score + float64(state.DeferredCount)*deferredBonusScore,
			TimesServed:    state.TimesServed,
			DeferredCount:  state.DeferredCount,
		})
	}
	return scoreboard
}

func (s *tripService) notifyScheduleUpdatedMembers(members []domain.RoomMember, actorUserID, tripID uint) {
	refType := "trip"
	for _, member := range members {
		if member.UserID == actorUserID {
			continue
		}
		if err := s.notifSvc.Notify(context.Background(), member.UserID, domain.NotifTypeScheduleUpdated, "Schedule updated", "The trip schedule has been updated.", &tripID, &refType); err != nil {
			log.Printf("[Notification] schedule_updated failed (user_id=%d): %v", member.UserID, err)
		}
	}
}

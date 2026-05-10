package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
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

// ── Reschedule trace types (internal to service package) ──────────────────────

type rescheduleScoreUpdate struct {
	UserID   uint
	Username string
	Gained   float64
	Reason   string
	OldScore float64
	NewScore float64
}

type fairnessRoundRecord struct {
	Round                  int
	PickedMemberID         uint
	PickedMemberUsername   string
	EffectiveScoreBefore   float64
	IsDeferred             bool
	DeferReason            string
	SelectedPlace          *rankedCandidate
	DistanceFromPrevKm     float64
	DistanceFromCentroidKm float64
	ScoreUpdates           []rescheduleScoreUpdate
	MemberStatesAfter      []domain.RescheduleMemberStateTrace
}

func snapshotMemberStates(states map[uint]*memberRescheduleState) []domain.RescheduleMemberStateTrace {
	ids := make([]uint, 0, len(states))
	for id := range states {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	result := make([]domain.RescheduleMemberStateTrace, 0, len(states))
	for _, id := range ids {
		st := states[id]
		result = append(result, domain.RescheduleMemberStateTrace{
			UserID:         st.UserID,
			Username:       st.Username,
			Score:          st.Score,
			EffectiveScore: st.Score + float64(st.DeferredCount)*deferredBonusScore,
			TimesServed:    st.TimesServed,
			DeferredCount:  st.DeferredCount,
		})
	}
	return result
}

// runFairnessSelectionWithTrace mirrors runFairnessSelection exactly but also
// records a FairnessRoundRecord for every turn (including deferred turns).
func runFairnessSelectionWithTrace(states map[uint]*memberRescheduleState) ([]rankedCandidate, int, []fairnessRoundRecord) {
	selected := make([]rankedCandidate, 0)
	selectedByPlaceID := map[string]bool{}
	centroid := calculateCandidatesCentroid(states)
	roundCount := 0
	failedTurns := 0
	records := make([]fairnessRoundRecord, 0)

	for {
		activeIDs := activeMemberIDs(states, selectedByPlaceID)
		if len(activeIDs) == 0 {
			break
		}

		nextMemberID := pickNextMemberID(activeIDs, states)
		memberState := states[nextMemberID]
		roundCount++

		effectiveBefore := memberState.Score + float64(memberState.DeferredCount)*deferredBonusScore

		candidate, found := findBestCandidateForMember(memberState, selected, selectedByPlaceID, centroid)
		if !found {
			memberState.DeferredCount++
			failedTurns++
			records = append(records, fairnessRoundRecord{
				Round:                roundCount,
				PickedMemberID:       nextMemberID,
				PickedMemberUsername: memberState.Username,
				EffectiveScoreBefore: effectiveBefore,
				IsDeferred:           true,
				DeferReason:          "distance_constraint_failed",
				MemberStatesAfter:    snapshotMemberStates(states),
			})
			if failedTurns >= len(activeIDs) {
				break
			}
			continue
		}

		failedTurns = 0

		distFromPrev := 0.0
		if len(selected) > 0 && hasCoordinates(candidate.Latitude, candidate.Longitude) {
			if last := selected[len(selected)-1]; hasCoordinates(last.Latitude, last.Longitude) {
				distFromPrev = math.Round(haversine(candidate.Latitude, candidate.Longitude, last.Latitude, last.Longitude)*100) / 100
			}
		}
		distFromCentroid := 0.0
		if centroid != nil && hasCoordinates(candidate.Latitude, candidate.Longitude) {
			distFromCentroid = math.Round(haversine(candidate.Latitude, candidate.Longitude, centroid.Latitude, centroid.Longitude)*100) / 100
		}

		selected = append(selected, candidate)
		selectedByPlaceID[candidate.PlaceID] = true

		updates := []rescheduleScoreUpdate{}

		oldOwnerScore := memberState.Score
		memberState.Score += ownerGainScore
		memberState.TimesServed++
		memberState.DeferredCount = 0
		memberState.LastSelectedRound = roundCount
		updates = append(updates, rescheduleScoreUpdate{
			UserID: memberState.UserID, Username: memberState.Username,
			Gained: ownerGainScore, Reason: "owner_pick",
			OldScore: oldOwnerScore, NewScore: memberState.Score,
		})

		sharedGains := map[uint]float64{}
		if candidate.Category != "" {
			normalizedCategory := normalizeCategory(candidate.Category)
			for uid, st := range states {
				if uid == memberState.UserID {
					continue
				}
				rank, ok := st.CategoryRank[normalizedCategory]
				if !ok {
					continue
				}
				gain := sharedGainByCategoryRank(rank)
				if gain > sharedGains[uid] {
					sharedGains[uid] = gain
				}
			}
		}
		for uid, gain := range sharedGains {
			st := states[uid]
			old := st.Score
			st.Score += gain
			updates = append(updates, rescheduleScoreUpdate{
				UserID: st.UserID, Username: st.Username,
				Gained: gain, Reason: fmt.Sprintf("shared_category_rank_%d", st.CategoryRank[normalizeCategory(candidate.Category)]),
				OldScore: old, NewScore: st.Score,
			})
		}
		sort.Slice(updates, func(i, j int) bool { return updates[i].UserID < updates[j].UserID })

		cp := candidate
		records = append(records, fairnessRoundRecord{
			Round:                  roundCount,
			PickedMemberID:         nextMemberID,
			PickedMemberUsername:   memberState.Username,
			EffectiveScoreBefore:   effectiveBefore,
			IsDeferred:             false,
			SelectedPlace:          &cp,
			DistanceFromPrevKm:     distFromPrev,
			DistanceFromCentroidKm: distFromCentroid,
			ScoreUpdates:           updates,
			MemberStatesAfter:      snapshotMemberStates(states),
		})
	}

	return selected, roundCount, records
}

func (s *tripService) GetReschedulePlanTrace(ctx context.Context, userID, tripID uint) (*domain.ReschedulePlanTrace, error) {
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

	members, err := s.roomSvc.GetMembersByRoomID(ctx, trip.RoomID)
	if err != nil {
		return nil, err
	}

	states := make(map[uint]*memberRescheduleState)
	notReadyMembers := make([]domain.RescheduleNotReadyMember, 0)
	namesToResolve := make([]string, 0)

	for _, member := range members {
		if member.Role != domain.RoleOwner && member.Role != domain.RoleMember {
			continue
		}
		lifestyle, err := s.lifestyleSvc.GetLifestyle(ctx, member.UserID, trip.RoomID)
		if err != nil {
			if errors.Is(err, domain.ErrLifestyleNotFound) {
				notReadyMembers = append(notReadyMembers, domain.RescheduleNotReadyMember{
					UserID: member.UserID, Username: member.User.Username,
				})
				continue
			}
			return nil, err
		}
		lifestyleID := lifestyle.LifestyleID
		if !domain.IsStructuredLifestyleValid(lifestyle.StructuredLifestyle) {
			notReadyMembers = append(notReadyMembers, domain.RescheduleNotReadyMember{
				UserID: member.UserID, Username: member.User.Username, LifestyleID: &lifestyleID,
			})
			continue
		}
		candidates, err := parseStructuredLifestylePlaces(*lifestyle.StructuredLifestyle)
		if err != nil {
			notReadyMembers = append(notReadyMembers, domain.RescheduleNotReadyMember{
				UserID: member.UserID, Username: member.User.Username, LifestyleID: &lifestyleID,
			})
			continue
		}
		candidates = dedupeCandidatesByPlaceID(candidates)
		for _, c := range candidates {
			if c.PlaceID == "" {
				namesToResolve = append(namesToResolve, c.Name)
			}
		}
		states[member.UserID] = &memberRescheduleState{
			UserID: member.UserID, Username: member.User.Username,
			Candidates: candidates, CategoryRank: map[string]int{}, LastSelectedRound: -1,
		}
	}

	if len(notReadyMembers) > 0 {
		sort.Slice(notReadyMembers, func(i, j int) bool { return notReadyMembers[i].UserID < notReadyMembers[j].UserID })
		return nil, &domain.RescheduleAnalysisNotReadyError{NotReadyMembers: notReadyMembers}
	}

	if len(namesToResolve) > 0 {
		attractionsByName, err := s.repo.GetAttractionsByNames(ctx, namesToResolve)
		if err != nil {
			return nil, err
		}
		for _, state := range states {
			resolved := make([]rankedCandidate, 0, len(state.Candidates))
			for _, c := range state.Candidates {
				if c.PlaceID == "" {
					c.PlaceID = resolveAttractionIDFromCandidates(c, attractionsByName[strings.ToLower(strings.TrimSpace(c.Name))])
				}
				if c.PlaceID == "" {
					continue
				}
				resolved = append(resolved, c)
			}
			state.Candidates = resolved
			state.CategoryRank = buildCategoryRank(resolved)
		}
	} else {
		for _, state := range states {
			state.CategoryRank = buildCategoryRank(state.Candidates)
		}
	}

	// Step 1: Build member + candidates summary
	centroid := calculateCandidatesCentroid(states)
	memberTraces := make([]domain.RescheduleMemberCandidateTrace, 0, len(states))
	memberIDs := make([]uint, 0, len(states))
	for id := range states {
		memberIDs = append(memberIDs, id)
	}
	sort.Slice(memberIDs, func(i, j int) bool { return memberIDs[i] < memberIDs[j] })
	for _, id := range memberIDs {
		st := states[id]
		candidateTraces := make([]domain.RescheduleCandidateTrace, len(st.Candidates))
		for i, c := range st.Candidates {
			candidateTraces[i] = domain.RescheduleCandidateTrace{
				Name: c.Name, PlaceID: c.PlaceID, Category: c.Category,
				Latitude: c.Latitude, Longitude: c.Longitude,
			}
		}
		memberTraces = append(memberTraces, domain.RescheduleMemberCandidateTrace{
			UserID: st.UserID, Username: st.Username,
			Candidates:   candidateTraces,
			CategoryRank: st.CategoryRank,
		})
	}

	var centroidTrace *domain.GeoPointTrace
	if centroid != nil {
		centroidTrace = &domain.GeoPointTrace{Latitude: centroid.Latitude, Longitude: centroid.Longitude}
	}

	// Step 2: Fairness selection with trace
	orderedCandidates, totalRounds, roundRecords := runFairnessSelectionWithTrace(states)

	fairnessRounds := make([]domain.FairnessRoundTrace, len(roundRecords))
	for i, r := range roundRecords {
		var selPlace *domain.RescheduleCandidateTrace
		if r.SelectedPlace != nil {
			cp := domain.RescheduleCandidateTrace{
				Name: r.SelectedPlace.Name, PlaceID: r.SelectedPlace.PlaceID,
				Category: r.SelectedPlace.Category,
				Latitude: r.SelectedPlace.Latitude, Longitude: r.SelectedPlace.Longitude,
			}
			selPlace = &cp
		}
		scoreUpdates := make([]domain.RescheduleScoreUpdateTrace, len(r.ScoreUpdates))
		for j, u := range r.ScoreUpdates {
			scoreUpdates[j] = domain.RescheduleScoreUpdateTrace{
				UserID: u.UserID, Username: u.Username, Gained: u.Gained, Reason: u.Reason,
				OldScore: u.OldScore, NewScore: u.NewScore,
			}
		}
		fairnessRounds[i] = domain.FairnessRoundTrace{
			Round:                  r.Round,
			PickedMemberID:         r.PickedMemberID,
			PickedMemberUsername:   r.PickedMemberUsername,
			EffectiveScoreBefore:   r.EffectiveScoreBefore,
			IsDeferred:             r.IsDeferred,
			DeferReason:            r.DeferReason,
			SelectedPlace:          selPlace,
			DistanceFromPrevKm:     r.DistanceFromPrevKm,
			DistanceFromCentroidKm: r.DistanceFromCentroidKm,
			ScoreUpdates:           scoreUpdates,
			MemberStatesAfter:      r.MemberStatesAfter,
		}
	}

	fairnessOrdered := make([]domain.RescheduleCandidateTrace, len(orderedCandidates))
	for i, c := range orderedCandidates {
		fairnessOrdered[i] = domain.RescheduleCandidateTrace{
			Name: c.Name, PlaceID: c.PlaceID, Category: c.Category,
			Latitude: c.Latitude, Longitude: c.Longitude,
		}
	}

	// Step 3: Nearest-neighbor ordering
	attractions := make([]domain.TripSchedule, len(orderedCandidates))
	for i, c := range orderedCandidates {
		attractions[i] = domain.TripSchedule{
			PlaceName: c.Name, PlaceID: c.PlaceID,
			Latitude: c.Latitude, Longitude: c.Longitude,
		}
	}
	nnSteps, nnOrdered := nearestNeighborOrderWithTrace(attractions)
	nnOrderedPlaces := make([]domain.PlanTracePlace, len(nnOrdered))
	for i, a := range nnOrdered {
		nnOrderedPlaces[i] = domain.PlanTracePlace{
			Name: a.PlaceName, PlaceID: a.PlaceID,
			Latitude: a.Latitude, Longitude: a.Longitude,
		}
	}

	// Step 4: Day distribution
	const maxPlacesPerDay = 4
	totalDays := int(trip.EndDate.Sub(trip.StartDate).Hours()/24) + 1
	if totalDays < 1 {
		totalDays = 1
	}
	maxPlaces := totalDays * maxPlacesPerDay
	var unscheduledRaw []domain.TripSchedule
	if len(nnOrdered) > maxPlaces {
		unscheduledRaw = nnOrdered[maxPlaces:]
		nnOrdered = nnOrdered[:maxPlaces]
	}
	placesPerDay := (len(nnOrdered) + totalDays - 1) / totalDays
	if placesPerDay < 1 {
		placesPerDay = 1
	}
	if placesPerDay > maxPlacesPerDay {
		placesPerDay = maxPlacesPerDay
	}
	scheduledPlaces := make([]domain.ScheduledPlaceTrace, len(nnOrdered))
	for i, a := range nnOrdered {
		day := i/placesPerDay + 1
		if day > totalDays {
			day = totalDays
		}
		scheduledPlaces[i] = domain.ScheduledPlaceTrace{
			Name: a.PlaceName, PlaceID: a.PlaceID,
			Latitude: a.Latitude, Longitude: a.Longitude,
			DayNumber: day, SequenceOrder: i%placesPerDay + 1,
		}
	}
	unscheduledPlaces := make([]domain.PlanTracePlace, len(unscheduledRaw))
	for i, u := range unscheduledRaw {
		unscheduledPlaces[i] = domain.PlanTracePlace{
			Name: u.PlaceName, PlaceID: u.PlaceID,
			Latitude: u.Latitude, Longitude: u.Longitude,
		}
	}

	// Step 5: Meals — read from current DB schedules
	schedules, err := s.repo.GetSchedulesByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	var restaurants []domain.TripSchedule
	for _, sc := range schedules {
		if sc.Type == "restaurant" {
			restaurants = append(restaurants, sc)
		}
	}

	type mealDef struct {
		name      string
		seq       int
		anchorIdx int
	}
	mealDefs := []mealDef{
		{"breakfast", 1, 0},
		{"lunch", 4, 1},
		{"dinner", 7, 3},
	}
	attrByDay := map[int][]domain.ScheduledPlaceTrace{}
	for _, a := range scheduledPlaces {
		attrByDay[a.DayNumber] = append(attrByDay[a.DayNumber], a)
	}
	restByDaySeq := map[[2]int]domain.TripSchedule{}
	for _, r := range restaurants {
		restByDaySeq[[2]int{r.DayNumber, r.SequenceOrder}] = r
	}
	var mealDetails []domain.MealSelectionDetail
	for day := 1; day <= totalDays; day++ {
		dayAttrs := attrByDay[day]
		for _, m := range mealDefs {
			r, ok := restByDaySeq[[2]int{day, m.seq}]
			if !ok {
				continue
			}
			anchorIdx := m.anchorIdx
			if anchorIdx >= len(dayAttrs) {
				if len(dayAttrs) == 0 {
					continue
				}
				anchorIdx = len(dayAttrs) - 1
			}
			anchor := dayAttrs[anchorIdx]
			dist := math.Round(haversine(anchor.Latitude, anchor.Longitude, r.Latitude, r.Longitude)*100) / 100
			mealDetails = append(mealDetails, domain.MealSelectionDetail{
				MealType: m.name, DayNumber: day, SequenceOrder: m.seq,
				AnchorPlace: domain.PlanTracePlace{
					Name: anchor.Name, PlaceID: anchor.PlaceID,
					Latitude: anchor.Latitude, Longitude: anchor.Longitude,
				},
				SelectedPlace: domain.PlanTracePlace{
					Name: r.PlaceName, PlaceID: r.PlaceID,
					Latitude: r.Latitude, Longitude: r.Longitude,
				},
				DistanceKm: dist,
			})
		}
	}

	return &domain.ReschedulePlanTrace{
		TripID: trip.TripID, DestinationName: trip.DestinationName,
		StartDate: trip.StartDate.Format("2006-01-02"), EndDate: trip.EndDate.Format("2006-01-02"),
		TotalDays: totalDays, PlacesPerDay: placesPerDay,
		Step1Members:               memberTraces,
		Step1Centroid:              centroidTrace,
		Step2FairnessRounds:        fairnessRounds,
		Step2FairnessOrderedPlaces: fairnessOrdered,
		Step2TotalRounds:           totalRounds,
		Step3NearestNeighborSteps:  nnSteps,
		Step3OrderedPlaces:         nnOrderedPlaces,
		Step4ScheduledPlaces:       scheduledPlaces,
		Step4UnscheduledPlaces:     unscheduledPlaces,
		Step5MealSelections:        mealDetails,
	}, nil
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

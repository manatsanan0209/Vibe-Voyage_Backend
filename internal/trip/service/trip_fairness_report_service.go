package service

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (s *tripService) GetFairnessReport(ctx context.Context, userID, tripID uint) (*domain.FairnessReport, error) {
	ok, err := s.repo.IsUserInTripRoom(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(trip.StructuredLifeStyle) == "" {
		return nil, nil
	}

	var snapshot fairnessRunSnapshot
	if err := json.Unmarshal([]byte(trip.StructuredLifeStyle), &snapshot); err != nil {
		return nil, nil
	}
	if snapshot.AlgorithmVersion == "" {
		return nil, nil
	}

	totalPlaces := len(snapshot.SelectedPlaceIDs)

	timesServedValues := make([]int, len(snapshot.Members))
	scoreValues := make([]float64, len(snapshot.Members))
	for i, m := range snapshot.Members {
		timesServedValues[i] = m.TimesServed
		scoreValues[i] = m.Score
	}

	reportMembers := make([]domain.FairnessReportMember, len(snapshot.Members))
	for i, m := range snapshot.Members {
		scheduleShare := 0.0
		if totalPlaces > 0 {
			scheduleShare = roundTo3(float64(m.TimesServed) / float64(totalPlaces) * 100)
		}
		deferredRate := 0.0
		if snapshot.RoundCount > 0 {
			deferredRate = roundTo3(float64(m.DeferredCount) / float64(snapshot.RoundCount))
		}
		reportMembers[i] = domain.FairnessReportMember{
			UserID:         m.UserID,
			Username:       m.Username,
			TimesServed:    m.TimesServed,
			Score:          m.Score,
			EffectiveScore: m.EffectiveScore,
			DeferredCount:  m.DeferredCount,
			ScheduleShare:  scheduleShare,
			DeferredRate:   deferredRate,
		}
	}

	return &domain.FairnessReport{
		GeneratedAt:      snapshot.GeneratedAt,
		AlgorithmVersion: snapshot.AlgorithmVersion,
		RoundCount:       snapshot.RoundCount,
		TotalPlaces:      totalPlaces,
		GiniCoefficient:  roundTo3(giniCoefficient(timesServedValues)),
		FairnessRatio:    roundTo3(fairnessRatio(timesServedValues)),
		ScoreStdDev:      roundTo3(stdDev(scoreValues)),
		Members:          reportMembers,
	}, nil
}

func (s *tripService) GetAggregatedFairnessReport(ctx context.Context, userID uint) (*domain.AggregatedFairnessReport, error) {
	trips, err := s.repo.GetTripsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summaries := make([]domain.TripFairnessSummary, 0, len(trips))
	for _, trip := range trips {
		summary := parseTripFairnessSummary(trip)
		if summary == nil {
			continue
		}
		summaries = append(summaries, *summary)
	}

	n := len(summaries)
	if n == 0 {
		return &domain.AggregatedFairnessReport{
			TripCount: 0,
			Trips:     []domain.TripFairnessSummary{},
		}, nil
	}

	var sumGini, sumRatio, sumStdDev, sumPlaces float64
	for _, s := range summaries {
		sumGini += s.GiniCoefficient
		sumRatio += s.FairnessRatio
		sumStdDev += s.ScoreStdDev
		sumPlaces += float64(s.TotalPlaces)
	}
	fn := float64(n)

	return &domain.AggregatedFairnessReport{
		TripCount:        n,
		AvgGini:          roundTo3(sumGini / fn),
		AvgFairnessRatio: roundTo3(sumRatio / fn),
		AvgScoreStdDev:   roundTo3(sumStdDev / fn),
		AvgTotalPlaces:   roundTo3(sumPlaces / fn),
		Trips:            summaries,
	}, nil
}

func parseTripFairnessSummary(trip domain.Trips) *domain.TripFairnessSummary {
	if strings.TrimSpace(trip.StructuredLifeStyle) == "" {
		return nil
	}
	var snapshot fairnessRunSnapshot
	if err := json.Unmarshal([]byte(trip.StructuredLifeStyle), &snapshot); err != nil {
		return nil
	}
	if snapshot.AlgorithmVersion == "" {
		return nil
	}

	totalPlaces := len(snapshot.SelectedPlaceIDs)
	timesServed := make([]int, len(snapshot.Members))
	scores := make([]float64, len(snapshot.Members))
	for i, m := range snapshot.Members {
		timesServed[i] = m.TimesServed
		scores[i] = m.Score
	}

	return &domain.TripFairnessSummary{
		TripID:          trip.TripID,
		DestinationName: trip.DestinationName,
		GeneratedAt:     snapshot.GeneratedAt,
		RoundCount:      snapshot.RoundCount,
		TotalPlaces:     totalPlaces,
		GiniCoefficient: roundTo3(giniCoefficient(timesServed)),
		FairnessRatio:   roundTo3(fairnessRatio(timesServed)),
		ScoreStdDev:     roundTo3(stdDev(scores)),
	}
}

func giniCoefficient(values []int) float64 {
	n := len(values)
	if n <= 1 {
		return 0
	}
	sorted := make([]int, n)
	copy(sorted, values)
	sort.Ints(sorted)

	sum := 0
	for _, v := range sorted {
		sum += v
	}
	if sum == 0 {
		return 0
	}

	weightedSum := 0
	for i, v := range sorted {
		weightedSum += (i + 1) * v
	}
	return (2.0*float64(weightedSum))/(float64(n)*float64(sum)) - float64(n+1)/float64(n)
}

func fairnessRatio(values []int) float64 {
	if len(values) == 0 {
		return 1.0
	}
	minVal, maxVal := values[0], values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return 1.0
	}
	return float64(minVal) / float64(maxVal)
}

func stdDev(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(n)

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(n))
}

func roundTo3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

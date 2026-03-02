package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type userLifestyleService struct {
	repo domain.UserLifestyleRepository
	db   *gorm.DB
}

func NewUserLifestyleService(repo domain.UserLifestyleRepository, db *gorm.DB) domain.UserLifestyleService {
	return &userLifestyleService{repo: repo, db: db}
}

func (s *userLifestyleService) GetLifestyle(ctx context.Context, userID, roomID uint) (*domain.UserLifestyle, error) {
	lifestyle, err := s.repo.GetByUserAndRoom(ctx, userID, roomID)
	if err != nil {
		return nil, err
	}
	if lifestyle == nil {
		return nil, errors.New("lifestyle not found")
	}
	return lifestyle, nil
}

func (s *userLifestyleService) AnalyzeLifestyle(ctx context.Context, lifestyleID uint) ([]domain.RecommendedPlace, error) {
	lifestyle, err := s.repo.GetByID(ctx, lifestyleID)
	if err != nil {
		return nil, err
	}
	if lifestyle == nil {
		return nil, errors.New("lifestyle not found")
	}

	var trip domain.Trips
	if err := s.db.WithContext(ctx).
		Where("room_id = ?", lifestyle.RoomID).
		First(&trip).Error; err != nil {
		return nil, fmt.Errorf("failed to get trip for room %d: %w", lifestyle.RoomID, err)
	}

	var travelVibes []string
	if err := json.Unmarshal([]byte(lifestyle.TravelVibes), &travelVibes); err != nil {
		travelVibes = []string{}
	}

	var voyagePriorities []string
	if err := json.Unmarshal([]byte(lifestyle.VoyagePriorities), &voyagePriorities); err != nil {
		voyagePriorities = []string{}
	}

	payload := map[string]interface{}{
		"attraction_types": voyagePriorities,
		"destination":      trip.DestinationName,
		"lifestyle_text":   lifestyle.AdditionalNotes,
		"lifestyle_types":  travelVibes,
	}

	log.Printf("destination: %s", trip.DestinationName);

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	recommendURL := os.Getenv("RECOMMEND_API_URL")
	if recommendURL == "" {
		recommendURL = "http://localhost:8000/api/v1/recommend"
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, recommendURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call recommend API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read recommend API response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("recommend API returned status %d: %s", resp.StatusCode, string(body))
	}

	var places []domain.RecommendedPlace
	if err := json.Unmarshal(body, &places); err != nil {
		return nil, fmt.Errorf("failed to parse recommend API response: %w", err)
	}

	structuredJSON := string(body)
	lifestyle.StructuredLifestyle = &structuredJSON
	if err := s.repo.Update(ctx, lifestyle); err != nil {
		return nil, fmt.Errorf("failed to update structured_lifestyle: %w", err)
	}

	return places, nil
}

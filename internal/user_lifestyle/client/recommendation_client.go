package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type httpRecommendationClient struct {
	client *http.Client
}

func NewHTTPRecommendationClient() domain.RecommendationClient {
	return &httpRecommendationClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *httpRecommendationClient) Recommend(ctx context.Context, req domain.RecommendationRequest) ([]domain.RecommendedPlace, string, error) {
	payload := map[string]interface{}{
		"destination_name":  req.DestinationName,
		"destination_id":    req.DestinationID,
		"travel_vibes":      req.TravelVibes,
		"voyage_priorities": req.VoyagePriorities,
		"food_vibes":        req.FoodVibes,
		"additional_notes":  req.AdditionalNotes,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	recommendURL := os.Getenv("RECOMMEND_API_URL")
	if recommendURL == "" {
		recommendURL = "http://localhost:8000/api/v1/recommend"
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, recommendURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to call recommend API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read recommend API response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("recommend API returned status %d: %s", resp.StatusCode, string(body))
	}

	var places []domain.RecommendedPlace
	if err := json.Unmarshal(body, &places); err != nil {
		return nil, "", fmt.Errorf("failed to parse recommend API response: %w", err)
	}

	return places, string(body), nil
}

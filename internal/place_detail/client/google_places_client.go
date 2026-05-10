package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

const (
	defaultPlacesBaseURL       = "https://places.googleapis.com/v1"
	textSearchFieldMask        = "places.id,places.displayName,places.location"
	placeDetailsFieldMask      = "rating,userRatingCount,regularOpeningHours,photos,googleMapsUri,editorialSummary"
	defaultLocationBiasRadiusM = 1500.0
	defaultMaxMatchDistanceKM  = 5.0
)

type googlePlacesClient struct {
	httpClient         *http.Client
	apiKey             string
	baseURL            string
	languageCode       string
	regionCode         string
	locationBiasRadius float64
	maxMatchDistanceKM float64
}

func NewGooglePlacesClientFromEnv() domain.GooglePlacesClient {
	maxMatchDistance := defaultMaxMatchDistanceKM
	if raw := os.Getenv("GOOGLE_PLACES_MAX_MATCH_DISTANCE_KM"); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed > 0 {
			maxMatchDistance = parsed
		}
	}

	return &googlePlacesClient{
		httpClient:         &http.Client{Timeout: 15 * time.Second},
		apiKey:             os.Getenv("GOOGLE_PLACES_API_KEY"),
		baseURL:            defaultString(os.Getenv("GOOGLE_PLACES_BASE_URL"), defaultPlacesBaseURL),
		languageCode:       defaultString(os.Getenv("GOOGLE_PLACES_LANGUAGE_CODE"), "th"),
		regionCode:         defaultString(os.Getenv("GOOGLE_PLACES_REGION_CODE"), "TH"),
		locationBiasRadius: defaultLocationBiasRadiusM,
		maxMatchDistanceKM: maxMatchDistance,
	}
}

func NewGooglePlacesClientForTest(httpClient *http.Client, apiKey, baseURL string) domain.GooglePlacesClient {
	return &googlePlacesClient{
		httpClient:         httpClient,
		apiKey:             apiKey,
		baseURL:            strings.TrimRight(baseURL, "/"),
		languageCode:       "th",
		regionCode:         "TH",
		locationBiasRadius: defaultLocationBiasRadiusM,
		maxMatchDistanceKM: defaultMaxMatchDistanceKM,
	}
}

func (c *googlePlacesClient) FetchPlaceDetailByText(ctx context.Context, input domain.GooglePlaceDetailFetchInput) (*domain.GooglePlaceDetailFetchResult, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, &domain.GooglePlacesError{Code: "missing_api_key", Message: "GOOGLE_PLACES_API_KEY is not set"}
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, &domain.GooglePlacesError{Code: "empty_place_name", Message: "place name is empty"}
	}

	placeID, err := c.searchPlaceID(ctx, input)
	if err != nil {
		return nil, err
	}
	return c.fetchDetails(ctx, placeID)
}

func (c *googlePlacesClient) searchPlaceID(ctx context.Context, input domain.GooglePlaceDetailFetchInput) (string, error) {
	body := map[string]interface{}{
		"textQuery":      input.Name,
		"maxResultCount": 3,
		"languageCode":   c.languageCode,
		"regionCode":     c.regionCode,
	}
	if input.Latitude != 0 || input.Longitude != 0 {
		body["locationBias"] = map[string]interface{}{
			"circle": map[string]interface{}{
				"center": map[string]float64{
					"latitude":  input.Latitude,
					"longitude": input.Longitude,
				},
				"radius": c.locationBiasRadius,
			},
		}
	}
	if input.SourceType == domain.PlaceDetailSourceRestaurant {
		body["includedType"] = "restaurant"
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/places:searchText", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", textSearchFieldMask)

	respBody, statusCode, err := c.do(req)
	if err != nil {
		return "", err
	}
	if statusCode < 200 || statusCode >= 300 {
		return "", googleHTTPError("text_search_failed", statusCode, respBody)
	}

	var parsed textSearchResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", &domain.GooglePlacesError{Code: "text_search_parse_failed", Message: err.Error(), Transient: true}
	}
	if len(parsed.Places) == 0 {
		return "", &domain.GooglePlacesError{Code: "place_not_found", Message: "text search returned no places"}
	}

	best := parsed.Places[0]
	bestDistance := math.MaxFloat64
	hasInputLocation := input.Latitude != 0 || input.Longitude != 0
	if hasInputLocation {
		for _, place := range parsed.Places {
			d := haversineKM(input.Latitude, input.Longitude, place.Location.Latitude, place.Location.Longitude)
			if d < bestDistance {
				bestDistance = d
				best = place
			}
		}
		if bestDistance > c.maxMatchDistanceKM {
			return "", &domain.GooglePlacesError{
				Code:    "place_match_too_far",
				Message: fmt.Sprintf("nearest text search result is %.2fkm away", bestDistance),
			}
		}
	}

	if best.ID == "" {
		return "", &domain.GooglePlacesError{Code: "place_id_missing", Message: "text search result did not include id"}
	}
	return best.ID, nil
}

func (c *googlePlacesClient) fetchDetails(ctx context.Context, placeID string) (*domain.GooglePlaceDetailFetchResult, error) {
	escapedPlaceID := url.PathEscape(placeID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/places/"+escapedPlaceID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", placeDetailsFieldMask)

	respBody, statusCode, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, googleHTTPError("place_details_failed", statusCode, respBody)
	}

	var parsed placeDetailsResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, &domain.GooglePlacesError{Code: "place_details_parse_failed", Message: err.Error(), Transient: true}
	}

	detail := domain.PlaceDetail{
		Rating:           parsed.Rating,
		UserRatingCount:  parsed.UserRatingCount,
		GoogleMapsURI:    parsed.GoogleMapsURI,
		EditorialSummary: parsed.EditorialSummary.Text,
	}
	if parsed.RegularOpeningHours != nil {
		detail.OpeningHours = &domain.PlaceDetailOpeningHours{
			WeekdayText: parsed.RegularOpeningHours.WeekdayDescriptions,
			OpenNow:     parsed.RegularOpeningHours.OpenNow,
		}
	}

	var attributions []domain.GooglePhotoAttribution
	if len(parsed.Photos) > 0 && parsed.Photos[0].Name != "" {
		photoURL, attrs := c.fetchPhotoURL(ctx, parsed.Photos[0])
		detail.PhotoURL = photoURL
		attributions = attrs
	}

	return &domain.GooglePlaceDetailFetchResult{
		GooglePlaceID:     placeID,
		Detail:            detail,
		PhotoAttributions: attributions,
	}, nil
}

func (c *googlePlacesClient) fetchPhotoURL(ctx context.Context, photo googlePhoto) (string, []domain.GooglePhotoAttribution) {
	attributions := make([]domain.GooglePhotoAttribution, 0, len(photo.AuthorAttributions))
	for _, attr := range photo.AuthorAttributions {
		attributions = append(attributions, domain.GooglePhotoAttribution{
			DisplayName: attr.DisplayName,
			URI:         attr.URI,
			PhotoURI:    attr.PhotoURI,
		})
	}

	photoURL := c.baseURL + "/" + strings.TrimLeft(photo.Name, "/") + "/media?maxHeightPx=480&maxWidthPx=720&skipHttpRedirect=true&key=" + url.QueryEscape(c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, photoURL, nil)
	if err != nil {
		return "", attributions
	}

	respBody, statusCode, err := c.do(req)
	if err != nil || statusCode < 200 || statusCode >= 300 {
		return "", attributions
	}

	var parsed photoMediaResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", attributions
	}
	return parsed.PhotoURI, attributions
}

func (c *googlePlacesClient) do(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, &domain.GooglePlacesError{Code: "google_request_failed", Message: err.Error(), Transient: true}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, &domain.GooglePlacesError{Code: "google_response_read_failed", Message: err.Error(), Transient: true}
	}
	return body, resp.StatusCode, nil
}

func googleHTTPError(code string, statusCode int, body []byte) error {
	transient := statusCode == http.StatusTooManyRequests || statusCode >= 500
	if statusCode == http.StatusBadRequest || statusCode == http.StatusNotFound {
		transient = false
	}
	return &domain.GooglePlacesError{
		Code:      fmt.Sprintf("%s_%d", code, statusCode),
		Message:   sanitizeGoogleErrorBody(body),
		Transient: transient,
	}
}

func sanitizeGoogleErrorBody(body []byte) string {
	message := strings.TrimSpace(string(body))
	if len(message) > 500 {
		return message[:500]
	}
	return message
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimRight(value, "/")
}

func haversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	const radiusKM = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return radiusKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

type textSearchResponse struct {
	Places []textSearchPlace `json:"places"`
}

type textSearchPlace struct {
	ID       string         `json:"id"`
	Location googleLocation `json:"location"`
}

type googleLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type placeDetailsResponse struct {
	Rating              *float64            `json:"rating"`
	UserRatingCount     *int                `json:"userRatingCount"`
	RegularOpeningHours *googleOpeningHours `json:"regularOpeningHours"`
	Photos              []googlePhoto       `json:"photos"`
	GoogleMapsURI       string              `json:"googleMapsUri"`
	EditorialSummary    googleLocalizedText `json:"editorialSummary"`
}

type googleOpeningHours struct {
	OpenNow             bool     `json:"openNow"`
	WeekdayDescriptions []string `json:"weekdayDescriptions"`
}

type googlePhoto struct {
	Name               string                    `json:"name"`
	AuthorAttributions []googleAuthorAttribution `json:"authorAttributions"`
}

type googleAuthorAttribution struct {
	DisplayName string `json:"displayName"`
	URI         string `json:"uri"`
	PhotoURI    string `json:"photoUri"`
}

type googleLocalizedText struct {
	Text         string `json:"text"`
	LanguageCode string `json:"languageCode"`
}

type photoMediaResponse struct {
	PhotoURI string `json:"photoUri"`
}

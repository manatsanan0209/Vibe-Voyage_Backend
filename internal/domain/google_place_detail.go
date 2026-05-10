package domain

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	PlaceDetailStatusCached      = "cached"
	PlaceDetailStatusPending     = "pending"
	PlaceDetailStatusUnavailable = "unavailable"

	PlaceDetailSourceAttraction = "attraction"
	PlaceDetailSourceRestaurant = "restaurant"

	PlaceDetailErrorMissingAPIKey = "missing_api_key"

	maxGooglePlaceDetailSourceKeyLength = 500
)

type PlaceDetailOpeningHours struct {
	WeekdayText []string `json:"weekday_text"`
	OpenNow     bool     `json:"open_now"`
}

type PlaceDetail struct {
	Rating           *float64                 `json:"rating,omitempty"`
	UserRatingCount  *int                     `json:"user_rating_count,omitempty"`
	OpeningHours     *PlaceDetailOpeningHours `json:"opening_hours,omitempty"`
	PhotoURL         string                   `json:"photo_url,omitempty"`
	GoogleMapsURI    string                   `json:"google_maps_uri,omitempty"`
	EditorialSummary string                   `json:"editorial_summary,omitempty"`
}

type PlaceDetailAttachment struct {
	Status string
	Detail *PlaceDetail
}

type GooglePlaceDetailSourceKey struct {
	SourceType      string
	SourcePlaceID   string
	SourceName      string
	SourceLatitude  float64
	SourceLongitude float64
}

func NewGooglePlaceDetailSourceKey(item TripSchedule) GooglePlaceDetailSourceKey {
	return GooglePlaceDetailSourceKey{
		SourceType:      strings.ToLower(strings.TrimSpace(item.Type)),
		SourcePlaceID:   strings.TrimSpace(item.PlaceID),
		SourceName:      strings.TrimSpace(item.PlaceName),
		SourceLatitude:  item.Latitude,
		SourceLongitude: item.Longitude,
	}
}

func (k GooglePlaceDetailSourceKey) SupportsGooglePlaceDetail() bool {
	return k.SourceType == PlaceDetailSourceAttraction || k.SourceType == PlaceDetailSourceRestaurant
}

func (k GooglePlaceDetailSourceKey) NormalizedName() string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(k.SourceName))), " ")
}

func (k GooglePlaceDetailSourceKey) RoundedLatitude() float64 {
	return math.Round(k.SourceLatitude*10000) / 10000
}

func (k GooglePlaceDetailSourceKey) RoundedLongitude() float64 {
	return math.Round(k.SourceLongitude*10000) / 10000
}

func (k GooglePlaceDetailSourceKey) CacheKey() string {
	sourceType := strings.ToLower(strings.TrimSpace(k.SourceType))
	sourcePlaceID := strings.TrimSpace(k.SourcePlaceID)
	if sourcePlaceID != "" {
		return truncateSourceKey(fmt.Sprintf("%s|id:%s", sourceType, sourcePlaceID))
	}
	return truncateSourceKey(fmt.Sprintf(
		"%s|fallback:%s|%.4f|%.4f",
		sourceType,
		k.NormalizedName(),
		k.RoundedLatitude(),
		k.RoundedLongitude(),
	))
}

func truncateSourceKey(value string) string {
	if len(value) <= maxGooglePlaceDetailSourceKeyLength {
		return value
	}
	runes := []rune(value)
	if len(runes) <= maxGooglePlaceDetailSourceKeyLength {
		return value
	}
	return string(runes[:maxGooglePlaceDetailSourceKeyLength])
}

type GooglePlaceDetail struct {
	GooglePlaceDetailID   uint           `json:"google_place_detail_id" gorm:"primaryKey;autoIncrement"`
	SourceType            string         `json:"source_type" gorm:"size:32;not null;uniqueIndex:idx_google_place_detail_source"`
	SourceKey             string         `json:"source_key" gorm:"size:512;not null;uniqueIndex:idx_google_place_detail_source"`
	SourcePlaceID         string         `json:"source_place_id" gorm:"size:255;index"`
	SourceName            string         `json:"source_name" gorm:"not null"`
	NormalizedName        string         `json:"normalized_name" gorm:"size:512;not null;index"`
	SourceLatitude        float64        `json:"source_latitude"`
	SourceLongitude       float64        `json:"source_longitude"`
	RoundedLatitude       float64        `json:"rounded_latitude"`
	RoundedLongitude      float64        `json:"rounded_longitude"`
	GooglePlaceID         string         `json:"google_place_id" gorm:"size:255;index"`
	Rating                *float64       `json:"rating"`
	UserRatingCount       *int           `json:"user_rating_count"`
	OpeningHoursJSON      string         `json:"opening_hours_json" gorm:"type:text"`
	PhotoURL              string         `json:"photo_url"`
	PhotoAttributionsJSON string         `json:"photo_attributions_json" gorm:"type:text"`
	GoogleMapsURI         string         `json:"google_maps_uri"`
	EditorialSummary      string         `json:"editorial_summary"`
	DetailStatus          string         `json:"detail_status" gorm:"size:32;not null;index"`
	LastFetchedAt         *time.Time     `json:"last_fetched_at"`
	ExpiresAt             *time.Time     `json:"expires_at" gorm:"index"`
	RetryCount            int            `json:"retry_count" gorm:"not null;default:0"`
	NextRetryAt           *time.Time     `json:"next_retry_at" gorm:"index"`
	FetchErrorCode        string         `json:"fetch_error_code" gorm:"size:128"`
	FetchErrorMessage     string         `json:"fetch_error_message"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type GooglePlaceDetailFetchInput struct {
	SourceType string
	Name       string
	Latitude   float64
	Longitude  float64
}

type GooglePhotoAttribution struct {
	DisplayName string `json:"display_name,omitempty"`
	URI         string `json:"uri,omitempty"`
	PhotoURI    string `json:"photo_uri,omitempty"`
}

type GooglePlaceDetailFetchResult struct {
	GooglePlaceID     string
	Detail            PlaceDetail
	PhotoAttributions []GooglePhotoAttribution
}

type GooglePlaceDetailRepository interface {
	FindBySourceKeys(ctx context.Context, sourceKeys []string) (map[string]*GooglePlaceDetail, error)
	SavePending(ctx context.Context, key GooglePlaceDetailSourceKey, nextRetryAt *time.Time) error
	SaveFetched(ctx context.Context, key GooglePlaceDetailSourceKey, result GooglePlaceDetailFetchResult, expiresAt time.Time) error
	MarkUnavailable(ctx context.Context, key GooglePlaceDetailSourceKey, code, message string) error
	MarkRetry(ctx context.Context, key GooglePlaceDetailSourceKey, code, message string, retryCount int, nextRetryAt time.Time) error
}

type GooglePlacesClient interface {
	FetchPlaceDetailByText(ctx context.Context, input GooglePlaceDetailFetchInput) (*GooglePlaceDetailFetchResult, error)
}

type GooglePlacesError struct {
	Code      string
	Message   string
	Transient bool
}

func (e *GooglePlacesError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type GooglePlaceDetailService interface {
	EnrichScheduleItems(ctx context.Context, items []TripSchedule) (map[string]PlaceDetailAttachment, error)
}

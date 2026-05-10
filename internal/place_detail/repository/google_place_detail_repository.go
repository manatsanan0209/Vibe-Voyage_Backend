package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type googlePlaceDetailRepository struct {
	db *gorm.DB
}

func NewGooglePlaceDetailRepository(db *gorm.DB) domain.GooglePlaceDetailRepository {
	return &googlePlaceDetailRepository{db: db}
}

func (r *googlePlaceDetailRepository) FindBySourceKeys(ctx context.Context, sourceKeys []string) (map[string]*domain.GooglePlaceDetail, error) {
	result := make(map[string]*domain.GooglePlaceDetail, len(sourceKeys))
	if len(sourceKeys) == 0 {
		return result, nil
	}

	rows := make([]domain.GooglePlaceDetail, 0, len(sourceKeys))
	if err := r.db.WithContext(ctx).
		Where("source_key IN ?", sourceKeys).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for i := range rows {
		row := rows[i]
		result[row.SourceKey] = &row
	}
	return result, nil
}

func (r *googlePlaceDetailRepository) SavePending(ctx context.Context, key domain.GooglePlaceDetailSourceKey, nextRetryAt *time.Time) error {
	row := detailFromKey(key)
	row.DetailStatus = domain.PlaceDetailStatusPending
	row.NextRetryAt = nextRetryAt

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "source_type"}, {Name: "source_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"source_place_id":   row.SourcePlaceID,
			"source_name":       row.SourceName,
			"normalized_name":   row.NormalizedName,
			"source_latitude":   row.SourceLatitude,
			"source_longitude":  row.SourceLongitude,
			"rounded_latitude":  row.RoundedLatitude,
			"rounded_longitude": row.RoundedLongitude,
			"detail_status":     domain.PlaceDetailStatusPending,
			"next_retry_at":     nextRetryAt,
			"updated_at":        time.Now(),
		}),
	}).Create(row).Error
}

func (r *googlePlaceDetailRepository) SaveFetched(ctx context.Context, key domain.GooglePlaceDetailSourceKey, result domain.GooglePlaceDetailFetchResult, expiresAt time.Time) error {
	row := detailFromKey(key)
	now := time.Now()
	row.GooglePlaceID = result.GooglePlaceID
	row.Rating = result.Detail.Rating
	row.UserRatingCount = result.Detail.UserRatingCount
	row.OpeningHoursJSON = marshalJSON(result.Detail.OpeningHours)
	row.PhotoURL = result.Detail.PhotoURL
	row.PhotoAttributionsJSON = marshalJSON(result.PhotoAttributions)
	row.GoogleMapsURI = result.Detail.GoogleMapsURI
	row.EditorialSummary = result.Detail.EditorialSummary
	row.DetailStatus = domain.PlaceDetailStatusCached
	row.LastFetchedAt = &now
	row.ExpiresAt = &expiresAt
	row.RetryCount = 0

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "source_type"}, {Name: "source_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"source_place_id":         row.SourcePlaceID,
			"source_name":             row.SourceName,
			"normalized_name":         row.NormalizedName,
			"source_latitude":         row.SourceLatitude,
			"source_longitude":        row.SourceLongitude,
			"rounded_latitude":        row.RoundedLatitude,
			"rounded_longitude":       row.RoundedLongitude,
			"google_place_id":         row.GooglePlaceID,
			"rating":                  row.Rating,
			"user_rating_count":       row.UserRatingCount,
			"opening_hours_json":      row.OpeningHoursJSON,
			"photo_url":               row.PhotoURL,
			"photo_attributions_json": row.PhotoAttributionsJSON,
			"google_maps_uri":         row.GoogleMapsURI,
			"editorial_summary":       row.EditorialSummary,
			"detail_status":           domain.PlaceDetailStatusCached,
			"last_fetched_at":         row.LastFetchedAt,
			"expires_at":              row.ExpiresAt,
			"retry_count":             0,
			"next_retry_at":           nil,
			"fetch_error_code":        "",
			"fetch_error_message":     "",
			"updated_at":              now,
		}),
	}).Create(row).Error
}

func (r *googlePlaceDetailRepository) MarkUnavailable(ctx context.Context, key domain.GooglePlaceDetailSourceKey, code, message string) error {
	row := detailFromKey(key)
	row.DetailStatus = domain.PlaceDetailStatusUnavailable
	row.FetchErrorCode = code
	row.FetchErrorMessage = message

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "source_type"}, {Name: "source_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"detail_status":       domain.PlaceDetailStatusUnavailable,
			"fetch_error_code":    code,
			"fetch_error_message": message,
			"next_retry_at":       nil,
			"updated_at":          time.Now(),
		}),
	}).Create(row).Error
}

func (r *googlePlaceDetailRepository) MarkRetry(ctx context.Context, key domain.GooglePlaceDetailSourceKey, code, message string, retryCount int, nextRetryAt time.Time) error {
	row := detailFromKey(key)
	row.DetailStatus = domain.PlaceDetailStatusPending
	row.FetchErrorCode = code
	row.FetchErrorMessage = message
	row.RetryCount = retryCount
	row.NextRetryAt = &nextRetryAt

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "source_type"}, {Name: "source_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"detail_status":       domain.PlaceDetailStatusPending,
			"fetch_error_code":    code,
			"fetch_error_message": message,
			"retry_count":         retryCount,
			"next_retry_at":       nextRetryAt,
			"updated_at":          time.Now(),
		}),
	}).Create(row).Error
}

func detailFromKey(key domain.GooglePlaceDetailSourceKey) *domain.GooglePlaceDetail {
	return &domain.GooglePlaceDetail{
		SourceType:       key.SourceType,
		SourceKey:        key.CacheKey(),
		SourcePlaceID:    key.SourcePlaceID,
		SourceName:       key.SourceName,
		NormalizedName:   key.NormalizedName(),
		SourceLatitude:   key.SourceLatitude,
		SourceLongitude:  key.SourceLongitude,
		RoundedLatitude:  key.RoundedLatitude(),
		RoundedLongitude: key.RoundedLongitude(),
	}
}

func marshalJSON(value interface{}) string {
	if value == nil {
		return "null"
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(payload)
}

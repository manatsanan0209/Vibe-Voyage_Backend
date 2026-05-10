package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

const (
	defaultDetailTTLHours = 720
	maxFetchRetries       = 3
	pendingDebounce       = 10 * time.Minute
	fetchTimeout          = 25 * time.Second
)

type googlePlaceDetailService struct {
	repo      domain.GooglePlaceDetailRepository
	client    domain.GooglePlacesClient
	ttl       time.Duration
	semaphore chan struct{}
	inFlight  sync.Map
}

func NewGooglePlaceDetailService(repo domain.GooglePlaceDetailRepository, client domain.GooglePlacesClient) domain.GooglePlaceDetailService {
	return &googlePlaceDetailService{
		repo:      repo,
		client:    client,
		ttl:       placeDetailsTTL(),
		semaphore: make(chan struct{}, 5),
	}
}

func (s *googlePlaceDetailService) EnrichScheduleItems(ctx context.Context, items []domain.TripSchedule) (map[string]domain.PlaceDetailAttachment, error) {
	attachments := make(map[string]domain.PlaceDetailAttachment)
	keys := uniqueSupportedKeys(items)
	if len(keys) == 0 {
		for _, item := range items {
			key := domain.NewGooglePlaceDetailSourceKey(item)
			attachments[key.CacheKey()] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusUnavailable}
		}
		return attachments, nil
	}

	sourceKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		sourceKeys = append(sourceKeys, key.CacheKey())
	}

	cached, err := s.repo.FindBySourceKeys(ctx, sourceKeys)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, key := range keys {
		cacheKey := key.CacheKey()
		row := cached[cacheKey]
		if row == nil {
			attachments[cacheKey] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusPending}
			s.markPendingAndEnqueue(ctx, key, nil, 0)
			continue
		}

		switch row.DetailStatus {
		case domain.PlaceDetailStatusCached:
			if row.ExpiresAt == nil || row.ExpiresAt.After(now) {
				attachments[cacheKey] = domain.PlaceDetailAttachment{
					Status: domain.PlaceDetailStatusCached,
					Detail: detailFromRecord(row),
				}
				continue
			}
			attachments[cacheKey] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusPending}
			s.markPendingAndEnqueue(ctx, key, nil, 0)
		case domain.PlaceDetailStatusUnavailable:
			if row.FetchErrorCode == domain.PlaceDetailErrorMissingAPIKey {
				attachments[cacheKey] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusPending}
				s.markPendingAndEnqueue(ctx, key, nil, 0)
				continue
			}
			attachments[cacheKey] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusUnavailable}
		case domain.PlaceDetailStatusPending:
			attachments[cacheKey] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusPending}
			if row.NextRetryAt == nil || !row.NextRetryAt.After(now) {
				s.markPendingAndEnqueue(ctx, key, row.NextRetryAt, row.RetryCount)
			}
		default:
			attachments[cacheKey] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusPending}
			s.markPendingAndEnqueue(ctx, key, nil, row.RetryCount)
		}
	}

	for _, item := range items {
		key := domain.NewGooglePlaceDetailSourceKey(item)
		if key.SupportsGooglePlaceDetail() {
			continue
		}
		attachments[key.CacheKey()] = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusUnavailable}
	}

	return attachments, nil
}

func (s *googlePlaceDetailService) markPendingAndEnqueue(_ context.Context, key domain.GooglePlaceDetailSourceKey, nextRetryAt *time.Time, retryCount int) {
	if nextRetryAt == nil {
		next := time.Now().Add(pendingDebounce)
		nextRetryAt = &next
	}
	s.enqueue(key, retryCount, nextRetryAt)
}

func (s *googlePlaceDetailService) enqueue(key domain.GooglePlaceDetailSourceKey, retryCount int, nextRetryAt *time.Time) {
	cacheKey := key.CacheKey()
	if _, loaded := s.inFlight.LoadOrStore(cacheKey, struct{}{}); loaded {
		return
	}

	go func() {
		s.semaphore <- struct{}{}
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[GooglePlaceDetail] async enrichment panic (source_key=%s): %v", cacheKey, r)
			}
			s.inFlight.Delete(cacheKey)
			<-s.semaphore
		}()

		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()

		if err := s.repo.SavePending(ctx, key, nextRetryAt); err != nil {
			log.Printf("[GooglePlaceDetail] save pending failed (source_key=%s): %v", cacheKey, err)
			return
		}

		result, err := s.client.FetchPlaceDetailByText(ctx, domain.GooglePlaceDetailFetchInput{
			SourceType: key.SourceType,
			Name:       key.SourceName,
			Latitude:   key.SourceLatitude,
			Longitude:  key.SourceLongitude,
		})
		if err != nil {
			s.handleFetchError(ctx, key, retryCount, err)
			return
		}

		expiresAt := time.Now().Add(s.ttl)
		if err := s.repo.SaveFetched(ctx, key, *result, expiresAt); err != nil {
			log.Printf("[GooglePlaceDetail] save fetched failed (source_key=%s): %v", cacheKey, err)
		}
	}()
}

func (s *googlePlaceDetailService) handleFetchError(ctx context.Context, key domain.GooglePlaceDetailSourceKey, retryCount int, err error) {
	var googleErr *domain.GooglePlacesError
	if !errors.As(err, &googleErr) {
		googleErr = &domain.GooglePlacesError{Code: "unknown_error", Message: err.Error(), Transient: true}
	}

	nextRetryCount := retryCount + 1
	if !googleErr.Transient || nextRetryCount >= maxFetchRetries {
		if markErr := s.repo.MarkUnavailable(ctx, key, googleErr.Code, googleErr.Message); markErr != nil {
			log.Printf("[GooglePlaceDetail] mark unavailable failed (source_key=%s): %v", key.CacheKey(), markErr)
		}
		return
	}

	nextRetryAt := time.Now().Add(retryBackoff(nextRetryCount))
	if markErr := s.repo.MarkRetry(ctx, key, googleErr.Code, googleErr.Message, nextRetryCount, nextRetryAt); markErr != nil {
		log.Printf("[GooglePlaceDetail] mark retry failed (source_key=%s): %v", key.CacheKey(), markErr)
	}
}

func uniqueSupportedKeys(items []domain.TripSchedule) []domain.GooglePlaceDetailSourceKey {
	seen := map[string]bool{}
	keys := make([]domain.GooglePlaceDetailSourceKey, 0, len(items))
	for _, item := range items {
		key := domain.NewGooglePlaceDetailSourceKey(item)
		if !key.SupportsGooglePlaceDetail() || key.SourceName == "" {
			continue
		}
		cacheKey := key.CacheKey()
		if seen[cacheKey] {
			continue
		}
		seen[cacheKey] = true
		keys = append(keys, key)
	}
	return keys
}

func detailFromRecord(row *domain.GooglePlaceDetail) *domain.PlaceDetail {
	detail := &domain.PlaceDetail{
		Rating:           row.Rating,
		UserRatingCount:  row.UserRatingCount,
		PhotoURL:         row.PhotoURL,
		GoogleMapsURI:    row.GoogleMapsURI,
		EditorialSummary: row.EditorialSummary,
	}
	if row.OpeningHoursJSON != "" && row.OpeningHoursJSON != "null" {
		var openingHours domain.PlaceDetailOpeningHours
		if err := json.Unmarshal([]byte(row.OpeningHoursJSON), &openingHours); err == nil {
			detail.OpeningHours = &openingHours
		}
	}
	return detail
}

func placeDetailsTTL() time.Duration {
	hours := defaultDetailTTLHours
	if raw := os.Getenv("GOOGLE_PLACE_DETAILS_TTL_HOURS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			hours = parsed
		}
	}
	return time.Duration(hours) * time.Hour
}

func retryBackoff(retryCount int) time.Duration {
	if retryCount <= 0 {
		retryCount = 1
	}
	if retryCount > maxFetchRetries {
		retryCount = maxFetchRetries
	}
	return time.Duration(retryCount*retryCount) * time.Minute
}

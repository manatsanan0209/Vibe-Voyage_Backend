package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func TestEnrichScheduleItemsAttachesCachedPendingAndUnavailable(t *testing.T) {
	now := time.Now()
	rating := 4.5
	count := 120
	cachedAttraction := domain.TripSchedule{
		PlaceName: "Cached Attraction",
		PlaceID:   "attr-1",
		Type:      domain.PlaceDetailSourceAttraction,
		Latitude:  13.7563,
		Longitude: 100.5018,
	}
	cachedRestaurant := domain.TripSchedule{
		PlaceName: "Cached Restaurant",
		PlaceID:   "rest-1",
		Type:      domain.PlaceDetailSourceRestaurant,
		Latitude:  13.7563,
		Longitude: 100.5018,
	}
	unavailable := domain.TripSchedule{
		PlaceName: "Unavailable Restaurant",
		PlaceID:   "rest-2",
		Type:      domain.PlaceDetailSourceRestaurant,
		Latitude:  13.7563,
		Longitude: 100.5018,
	}
	missing := domain.TripSchedule{
		PlaceName: "Missing Attraction",
		PlaceID:   "attr-2",
		Type:      domain.PlaceDetailSourceAttraction,
		Latitude:  13.7563,
		Longitude: 100.5018,
	}

	repo := &fakeGooglePlaceDetailRepo{records: map[string]*domain.GooglePlaceDetail{
		domain.NewGooglePlaceDetailSourceKey(cachedAttraction).CacheKey(): {
			SourceKey:       domain.NewGooglePlaceDetailSourceKey(cachedAttraction).CacheKey(),
			DetailStatus:    domain.PlaceDetailStatusCached,
			Rating:          &rating,
			UserRatingCount: &count,
			ExpiresAt:       ptrTime(now.Add(time.Hour)),
		},
		domain.NewGooglePlaceDetailSourceKey(cachedRestaurant).CacheKey(): {
			SourceKey:     domain.NewGooglePlaceDetailSourceKey(cachedRestaurant).CacheKey(),
			DetailStatus:  domain.PlaceDetailStatusCached,
			GoogleMapsURI: "https://maps.google.com/restaurant",
			ExpiresAt:     ptrTime(now.Add(time.Hour)),
		},
		domain.NewGooglePlaceDetailSourceKey(unavailable).CacheKey(): {
			SourceKey:    domain.NewGooglePlaceDetailSourceKey(unavailable).CacheKey(),
			DetailStatus: domain.PlaceDetailStatusUnavailable,
		},
	}}
	svc := NewGooglePlaceDetailService(repo, fakeGooglePlacesClient{})

	result, err := svc.EnrichScheduleItems(context.Background(), []domain.TripSchedule{
		cachedAttraction,
		cachedRestaurant,
		unavailable,
		missing,
		missing,
	})
	if err != nil {
		t.Fatalf("EnrichScheduleItems returned error: %v", err)
	}

	if got := result[domain.NewGooglePlaceDetailSourceKey(cachedAttraction).CacheKey()]; got.Status != domain.PlaceDetailStatusCached || got.Detail == nil {
		t.Fatalf("expected cached attraction detail, got %#v", got)
	}
	if got := result[domain.NewGooglePlaceDetailSourceKey(cachedRestaurant).CacheKey()]; got.Status != domain.PlaceDetailStatusCached || got.Detail == nil {
		t.Fatalf("expected cached restaurant detail, got %#v", got)
	}
	if got := result[domain.NewGooglePlaceDetailSourceKey(unavailable).CacheKey()]; got.Status != domain.PlaceDetailStatusUnavailable || got.Detail != nil {
		t.Fatalf("expected unavailable without detail, got %#v", got)
	}
	if got := result[domain.NewGooglePlaceDetailSourceKey(missing).CacheKey()]; got.Status != domain.PlaceDetailStatusPending || got.Detail != nil {
		t.Fatalf("expected pending missing detail, got %#v", got)
	}
	waitForPendingSaves(t, repo, 1)
}

func TestEnrichScheduleItemsDoesNotRetryUnavailable(t *testing.T) {
	item := domain.TripSchedule{
		PlaceName: "Unavailable Restaurant",
		PlaceID:   "rest-2",
		Type:      domain.PlaceDetailSourceRestaurant,
		Latitude:  13.7563,
		Longitude: 100.5018,
	}
	key := domain.NewGooglePlaceDetailSourceKey(item).CacheKey()
	repo := &fakeGooglePlaceDetailRepo{records: map[string]*domain.GooglePlaceDetail{
		key: {SourceKey: key, DetailStatus: domain.PlaceDetailStatusUnavailable},
	}}
	svc := NewGooglePlaceDetailService(repo, fakeGooglePlacesClient{})

	result, err := svc.EnrichScheduleItems(context.Background(), []domain.TripSchedule{item})
	if err != nil {
		t.Fatalf("EnrichScheduleItems returned error: %v", err)
	}
	if result[key].Status != domain.PlaceDetailStatusUnavailable {
		t.Fatalf("expected unavailable, got %#v", result[key])
	}
	if repo.savePendingCount != 0 {
		t.Fatalf("unavailable place should not be queued again, got %d pending saves", repo.savePendingCount)
	}
}

type fakeGooglePlaceDetailRepo struct {
	mu               sync.Mutex
	records          map[string]*domain.GooglePlaceDetail
	savePendingCount int
}

func (r *fakeGooglePlaceDetailRepo) FindBySourceKeys(ctx context.Context, sourceKeys []string) (map[string]*domain.GooglePlaceDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make(map[string]*domain.GooglePlaceDetail)
	for _, key := range sourceKeys {
		if row := r.records[key]; row != nil {
			copy := *row
			result[key] = &copy
		}
	}
	return result, nil
}

func (r *fakeGooglePlaceDetailRepo) SavePending(ctx context.Context, key domain.GooglePlaceDetailSourceKey, nextRetryAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.savePendingCount++
	return nil
}

func (r *fakeGooglePlaceDetailRepo) SaveFetched(ctx context.Context, key domain.GooglePlaceDetailSourceKey, result domain.GooglePlaceDetailFetchResult, expiresAt time.Time) error {
	return nil
}

func (r *fakeGooglePlaceDetailRepo) MarkUnavailable(ctx context.Context, key domain.GooglePlaceDetailSourceKey, code, message string) error {
	return nil
}

func (r *fakeGooglePlaceDetailRepo) MarkRetry(ctx context.Context, key domain.GooglePlaceDetailSourceKey, code, message string, retryCount int, nextRetryAt time.Time) error {
	return nil
}

type fakeGooglePlacesClient struct{}

func (fakeGooglePlacesClient) FetchPlaceDetailByText(ctx context.Context, input domain.GooglePlaceDetailFetchInput) (*domain.GooglePlaceDetailFetchResult, error) {
	return &domain.GooglePlaceDetailFetchResult{
		GooglePlaceID: "google-place",
		Detail: domain.PlaceDetail{
			GoogleMapsURI: "https://maps.google.com/example",
		},
	}, nil
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func waitForPendingSaves(t *testing.T, repo *fakeGooglePlaceDetailRepo, want int) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		repo.mu.Lock()
		got := repo.savePendingCount
		repo.mu.Unlock()

		if got == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	repo.mu.Lock()
	got := repo.savePendingCount
	repo.mu.Unlock()
	t.Fatalf("expected pending saves to become %d, got %d", want, got)
}

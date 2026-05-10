package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func TestFetchPlaceDetailByTextUsesTextSearchAndDetailsFieldMasks(t *testing.T) {
	var sawSearch bool
	var sawDetails bool
	var searchBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/places:searchText":
			sawSearch = true
			if got := r.Header.Get("X-Goog-FieldMask"); got != textSearchFieldMask {
				t.Fatalf("unexpected search field mask: %s", got)
			}
			if strings.Contains(r.Header.Get("X-Goog-FieldMask"), "*") {
				t.Fatal("search field mask must not use wildcard")
			}
			if err := json.NewDecoder(r.Body).Decode(&searchBody); err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write([]byte(`{
				"places": [
					{"id":"far-place","location":{"latitude":14.0000,"longitude":101.0000}},
					{"id":"google-place-1","location":{"latitude":13.7563,"longitude":100.5018}}
				]
			}`))
		case "/places/google-place-1":
			sawDetails = true
			if got := r.Header.Get("X-Goog-FieldMask"); got != placeDetailsFieldMask {
				t.Fatalf("unexpected details field mask: %s", got)
			}
			if strings.Contains(r.Header.Get("X-Goog-FieldMask"), "*") {
				t.Fatal("details field mask must not use wildcard")
			}
			_, _ = w.Write([]byte(`{
				"rating": 4.5,
				"userRatingCount": 1240,
				"regularOpeningHours": {
					"openNow": true,
					"weekdayDescriptions": ["Monday: 9:00 AM - 6:00 PM"]
				},
				"googleMapsUri": "https://maps.google.com/example",
				"editorialSummary": {"text": "A short summary."}
			}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewGooglePlacesClientForTest(server.Client(), "test-key", server.URL)
	result, err := client.FetchPlaceDetailByText(context.Background(), domain.GooglePlaceDetailFetchInput{
		SourceType: domain.PlaceDetailSourceRestaurant,
		Name:       "Example Restaurant",
		Latitude:   13.7563,
		Longitude:  100.5018,
	})
	if err != nil {
		t.Fatalf("FetchPlaceDetailByText returned error: %v", err)
	}
	if !sawSearch || !sawDetails {
		t.Fatalf("expected both search and details calls, sawSearch=%v sawDetails=%v", sawSearch, sawDetails)
	}
	if searchBody["textQuery"] != "Example Restaurant" {
		t.Fatalf("unexpected textQuery: %v", searchBody["textQuery"])
	}
	if searchBody["includedType"] != "restaurant" {
		t.Fatalf("restaurant search should include restaurant type, got %v", searchBody["includedType"])
	}
	if _, ok := searchBody["locationBias"]; !ok {
		t.Fatal("expected locationBias for search")
	}
	if result.GooglePlaceID != "google-place-1" {
		t.Fatalf("expected nearest place id, got %s", result.GooglePlaceID)
	}
	if result.Detail.Rating == nil || *result.Detail.Rating != 4.5 {
		t.Fatalf("unexpected rating: %#v", result.Detail.Rating)
	}
	if result.Detail.UserRatingCount == nil || *result.Detail.UserRatingCount != 1240 {
		t.Fatalf("unexpected user rating count: %#v", result.Detail.UserRatingCount)
	}
	if result.Detail.OpeningHours == nil || !result.Detail.OpeningHours.OpenNow || len(result.Detail.OpeningHours.WeekdayText) != 1 {
		t.Fatalf("unexpected opening hours: %#v", result.Detail.OpeningHours)
	}
	if result.Detail.GoogleMapsURI != "https://maps.google.com/example" {
		t.Fatalf("unexpected google maps uri: %s", result.Detail.GoogleMapsURI)
	}
	if result.Detail.EditorialSummary != "A short summary." {
		t.Fatalf("unexpected editorial summary: %s", result.Detail.EditorialSummary)
	}
}

func TestFetchPlaceDetailByTextDoesNotApplyRestaurantTypeToAttractions(t *testing.T) {
	var searchBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/places:searchText":
			if err := json.NewDecoder(r.Body).Decode(&searchBody); err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write([]byte(`{"places":[{"id":"google-place-1","location":{"latitude":13.7563,"longitude":100.5018}}]}`))
		case "/places/google-place-1":
			_, _ = w.Write([]byte(`{"googleMapsUri":"https://maps.google.com/example"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewGooglePlacesClientForTest(server.Client(), "test-key", server.URL)
	_, err := client.FetchPlaceDetailByText(context.Background(), domain.GooglePlaceDetailFetchInput{
		SourceType: domain.PlaceDetailSourceAttraction,
		Name:       "Example Temple",
		Latitude:   13.7563,
		Longitude:  100.5018,
	})
	if err != nil {
		t.Fatalf("FetchPlaceDetailByText returned error: %v", err)
	}
	if _, ok := searchBody["includedType"]; ok {
		t.Fatalf("attraction search should not set includedType, got %v", searchBody["includedType"])
	}
}

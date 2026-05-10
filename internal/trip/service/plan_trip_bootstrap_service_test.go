package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func TestGetPlanTripBootstrapForbiddenWhenUserIsNotRoomMember(t *testing.T) {
	svc := &tripService{
		repo: &fakeTripBootstrapRepo{},
	}

	_, err := svc.GetPlanTripBootstrap(context.Background(), 10, 52)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestGetPlanTripBootstrapOwnerIncludesBundleAndPublishStatus(t *testing.T) {
	publishedAt := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	publishedID := uint(91)
	lifestyleID := uint(33)
	repo := &fakeTripBootstrapRepo{
		membership: &domain.TripRoomMembership{
			Trip: &domain.Trips{
				TripID:          52,
				RoomID:          52,
				DestinationName: "Chiang Mai",
				StartDate:       time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				EndDate:         time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
			},
			RoomMemberID: 77,
			Role:         domain.RoleOwner,
		},
		schedules: []domain.TripSchedule{
			{TripScheduleID: 1, TripID: 52, PlaceName: "Suggestion", DayNumber: 0, SequenceOrder: 0},
			{TripScheduleID: 2, TripID: 52, PlaceName: "Scheduled", DayNumber: 1, SequenceOrder: 1},
		},
	}
	roomRepo := &fakeRoomBootstrapRepo{
		statuses: []domain.MemberLifestyleSubmissionStatus{
			{
				RoomMemberID:          77,
				RoomID:                52,
				UserID:                10,
				Username:              "owner",
				Role:                  domain.RoleOwner,
				HasSubmittedLifestyle: true,
				HasAnalyzedLifestyle:  true,
			},
			{
				RoomMemberID:          78,
				RoomID:                52,
				UserID:                11,
				Username:              "member",
				Role:                  domain.RoleMember,
				HasSubmittedLifestyle: true,
				HasAnalyzedLifestyle:  false,
				SubmittedLifestyleID:  &lifestyleID,
			},
		},
	}
	suggestionSvc := &fakeTripBootstrapSuggestionSvc{
		published: &domain.PublishedTrip{
			PublishedTripID: publishedID,
			TripID:          52,
			Title:           "Published trip",
			CreatedAt:       publishedAt,
		},
	}
	svc := &tripService{
		repo:          repo,
		roomRepo:      roomRepo,
		suggestionSvc: suggestionSvc,
	}

	result, err := svc.GetPlanTripBootstrap(context.Background(), 10, 52)
	if err != nil {
		t.Fatalf("GetPlanTripBootstrap returned error: %v", err)
	}
	if result.Trip.RoomID != 52 || result.CurrentMember.Role != domain.RoleOwner {
		t.Fatalf("current trip/member = %#v/%#v, want room 52 owner", result.Trip, result.CurrentMember)
	}
	if len(result.Suggestions) != 1 || len(result.Days) != 1 {
		t.Fatalf("schedule groups = suggestions:%d days:%d, want 1/1", len(result.Suggestions), len(result.Days))
	}
	if len(result.Members) != 2 {
		t.Fatalf("members = %d, want 2", len(result.Members))
	}
	if result.PublishStatus == nil || !result.PublishStatus.IsPublished || result.PublishStatus.PublishedTripID == nil || *result.PublishStatus.PublishedTripID != publishedID {
		t.Fatalf("publish status = %#v, want published id %d", result.PublishStatus, publishedID)
	}
	if result.SchedulePollAfterMS != 5000 || result.ReadinessPollAfterMS != 3000 {
		t.Fatalf("polling = %d/%d, want 5000/3000", result.SchedulePollAfterMS, result.ReadinessPollAfterMS)
	}
}

type fakeTripBootstrapRepo struct {
	membership *domain.TripRoomMembership
	schedules  []domain.TripSchedule
}

func (r *fakeTripBootstrapRepo) GetByID(context.Context, uint) (*domain.Trips, error) {
	return nil, nil
}

func (r *fakeTripBootstrapRepo) GetByRoomID(context.Context, uint) (*domain.Trips, error) {
	return nil, nil
}

func (r *fakeTripBootstrapRepo) IsUserInTripRoom(context.Context, uint, uint) (bool, error) {
	return false, nil
}

func (r *fakeTripBootstrapRepo) GetUserRoleInTripRoom(context.Context, uint, uint) (int, bool, error) {
	return 0, false, nil
}

func (r *fakeTripBootstrapRepo) GetTripRoomMembership(context.Context, uint, uint) (*domain.TripRoomMembership, bool, error) {
	if r.membership == nil {
		return nil, false, nil
	}
	return r.membership, true, nil
}

func (r *fakeTripBootstrapRepo) GetSchedulesByTripID(context.Context, uint) ([]domain.TripSchedule, error) {
	return r.schedules, nil
}

func (r *fakeTripBootstrapRepo) UpdateGroupStructuredLifestyle(context.Context, uint, string) error {
	return nil
}

func (r *fakeTripBootstrapRepo) GetAttractionsByNames(context.Context, []string) (map[string][]domain.Attraction, error) {
	return nil, nil
}

func (r *fakeTripBootstrapRepo) CreateTripBundle(context.Context, uint, domain.CreateTripInput, string, string, string, string) (*domain.CreateTripResult, error) {
	return nil, nil
}

func (r *fakeTripBootstrapRepo) CreateSchedules(context.Context, []domain.TripSchedule) error {
	return nil
}

func (r *fakeTripBootstrapRepo) ReplaceSchedulesByTripID(context.Context, uint, []domain.TripSchedule) error {
	return nil
}

func (r *fakeTripBootstrapRepo) ReplaceScheduleAndSnapshot(context.Context, uint, []domain.TripSchedule, string) error {
	return nil
}

func (r *fakeTripBootstrapRepo) GetTripsByUserID(context.Context, uint) ([]domain.Trips, error) {
	return nil, nil
}

type fakeRoomBootstrapRepo struct {
	statuses []domain.MemberLifestyleSubmissionStatus
}

func (r *fakeRoomBootstrapRepo) GetByRoomID(context.Context, uint) ([]domain.RoomMember, error) {
	return nil, nil
}

func (r *fakeRoomBootstrapRepo) GetRoomsByUserID(context.Context, uint) ([]domain.UserRoomSummary, error) {
	return nil, nil
}

func (r *fakeRoomBootstrapRepo) GetMemberLifestyleStatusesByRoomID(context.Context, uint) ([]domain.MemberLifestyleSubmissionStatus, error) {
	return r.statuses, nil
}

func (r *fakeRoomBootstrapRepo) GetByID(context.Context, uint) (*domain.RoomMember, error) {
	return nil, nil
}

func (r *fakeRoomBootstrapRepo) AddMember(context.Context, *domain.RoomMember) (*domain.RoomMember, error) {
	return nil, nil
}

func (r *fakeRoomBootstrapRepo) DeleteMember(context.Context, uint) error {
	return nil
}

func (r *fakeRoomBootstrapRepo) ExistsByRoomAndUser(context.Context, uint, uint) (bool, error) {
	return false, nil
}

func (r *fakeRoomBootstrapRepo) GetRoomInfoByID(context.Context, uint) (*domain.Room, error) {
	return nil, nil
}

func (r *fakeRoomBootstrapRepo) UpdateRoom(context.Context, uint, domain.UpdateRoomInput) (*domain.Room, error) {
	return nil, nil
}

func (r *fakeRoomBootstrapRepo) UpdateMemberRole(context.Context, uint, int) (*domain.RoomMember, error) {
	return nil, nil
}

func (r *fakeRoomBootstrapRepo) TransferOwnership(context.Context, uint, uint, uint) error {
	return nil
}

type fakeTripBootstrapSuggestionSvc struct {
	published *domain.PublishedTrip
}

func (s *fakeTripBootstrapSuggestionSvc) GetFeed(context.Context, int, int, uint) ([]domain.PublishedTripWithMeta, int64, error) {
	return nil, 0, nil
}

func (s *fakeTripBootstrapSuggestionSvc) GetDetail(context.Context, uint, uint) (*domain.PublishedTripWithMeta, error) {
	return nil, nil
}

func (s *fakeTripBootstrapSuggestionSvc) GetPublishedTripByTripID(context.Context, uint) (*domain.PublishedTrip, error) {
	if s.published == nil {
		return nil, errors.New("not found")
	}
	return s.published, nil
}

func (s *fakeTripBootstrapSuggestionSvc) PublishTrip(context.Context, uint, uint, string, string) (*domain.PublishedTrip, error) {
	return nil, nil
}

func (s *fakeTripBootstrapSuggestionSvc) UnpublishTrip(context.Context, uint, uint) error {
	return nil
}

func (s *fakeTripBootstrapSuggestionSvc) ToggleLike(context.Context, uint, uint) (bool, error) {
	return false, nil
}

func (s *fakeTripBootstrapSuggestionSvc) ToggleBookmark(context.Context, uint, uint) (bool, error) {
	return false, nil
}

func (s *fakeTripBootstrapSuggestionSvc) GetBookmarks(context.Context, uint) ([]domain.PublishedTripWithMeta, error) {
	return nil, nil
}

func (s *fakeTripBootstrapSuggestionSvc) UseAsTemplate(context.Context, uint, uint, domain.UseAsTemplateInput) (*domain.CreateTripResult, error) {
	return nil, nil
}

func (s *fakeTripBootstrapSuggestionSvc) GetMyPosts(context.Context, uint, int, int) ([]domain.PublishedTripWithMeta, int64, error) {
	return nil, 0, nil
}

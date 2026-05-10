package handler

import (
	"testing"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func TestToPlanTripBootstrapResponseDTOOwnerWaitingForAnalysis(t *testing.T) {
	publishedID := uint(91)
	publishedAt := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	lifestyleID := uint(33)

	result := baseBootstrapResult(domain.RoleOwner)
	result.PublishStatus = &domain.PlanTripPublishStatus{
		IsPublished:     true,
		PublishedTripID: &publishedID,
		Title:           "Chiang Mai",
		PublishedAt:     &publishedAt,
	}
	result.Members = append(result.Members, domain.MemberLifestyleSubmissionStatus{
		RoomMemberID:          78,
		RoomID:                52,
		UserID:                11,
		Username:              "member",
		Role:                  domain.RoleMember,
		HasSubmittedLifestyle: true,
		HasAnalyzedLifestyle:  false,
		SubmittedLifestyleID:  &lifestyleID,
	})

	resp := toPlanTripBootstrapResponseDTO(result, 10)

	if resp.CurrentUser.RoleName != "owner" {
		t.Fatalf("role_name = %q, want owner", resp.CurrentUser.RoleName)
	}
	if !resp.CurrentUser.CanEdit || !resp.CurrentUser.CanManageRoom {
		t.Fatalf("owner permissions = edit:%v manage:%v, want both true", resp.CurrentUser.CanEdit, resp.CurrentUser.CanManageRoom)
	}
	if resp.PublishStatus == nil || !resp.PublishStatus.IsPublished || resp.PublishStatus.PublishedTripID == nil || *resp.PublishStatus.PublishedTripID != publishedID {
		t.Fatalf("publish_status = %#v, want published id %d", resp.PublishStatus, publishedID)
	}
	if resp.RescheduleReadiness.Status != "waiting_for_member_analysis" {
		t.Fatalf("readiness = %q, want waiting_for_member_analysis", resp.RescheduleReadiness.Status)
	}
	if len(resp.RescheduleReadiness.WaitingMembers) != 1 {
		t.Fatalf("waiting members = %d, want 1", len(resp.RescheduleReadiness.WaitingMembers))
	}
}

func TestToPlanTripBootstrapResponseDTONonOwnerHidesPublishStatus(t *testing.T) {
	result := baseBootstrapResult(domain.RoleMember)

	resp := toPlanTripBootstrapResponseDTO(result, 10)

	if resp.CurrentUser.RoleName != "member" {
		t.Fatalf("role_name = %q, want member", resp.CurrentUser.RoleName)
	}
	if !resp.CurrentUser.CanEdit || resp.CurrentUser.CanManageRoom {
		t.Fatalf("member permissions = edit:%v manage:%v, want edit true manage false", resp.CurrentUser.CanEdit, resp.CurrentUser.CanManageRoom)
	}
	if resp.PublishStatus != nil {
		t.Fatalf("publish_status = %#v, want nil", resp.PublishStatus)
	}
	if resp.RescheduleReadiness.Status != "not_owner" {
		t.Fatalf("readiness = %q, want not_owner", resp.RescheduleReadiness.Status)
	}
}

func TestToPlanTripBootstrapResponseDTOSpectatorIsViewOnly(t *testing.T) {
	result := baseBootstrapResult(domain.RoleSpectator)

	resp := toPlanTripBootstrapResponseDTO(result, 10)

	if resp.CurrentUser.RoleName != "spectator" {
		t.Fatalf("role_name = %q, want spectator", resp.CurrentUser.RoleName)
	}
	if resp.CurrentUser.CanEdit || resp.CurrentUser.CanManageRoom {
		t.Fatalf("spectator permissions = edit:%v manage:%v, want both false", resp.CurrentUser.CanEdit, resp.CurrentUser.CanManageRoom)
	}
}

func baseBootstrapResult(role int) *domain.PlanTripBootstrapResult {
	return &domain.PlanTripBootstrapResult{
		Trip: &domain.Trips{
			TripID:          52,
			RoomID:          52,
			DestinationName: "Chiang Mai",
			StartDate:       time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		},
		CurrentMember: domain.TripRoomMembership{
			RoomMemberID: 77,
			Role:         role,
		},
		Members: []domain.MemberLifestyleSubmissionStatus{
			{
				RoomMemberID:          77,
				RoomID:                52,
				UserID:                10,
				Username:              "owner",
				Role:                  role,
				HasSubmittedLifestyle: true,
				HasAnalyzedLifestyle:  true,
			},
		},
		SchedulePollAfterMS:  5000,
		ReadinessPollAfterMS: 3000,
	}
}

package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type tripHandler struct {
	svc           domain.TripService
	suggestionSvc domain.TripSuggestionService
}

func NewTripHandler(svc domain.TripService, suggestionSvc domain.TripSuggestionService) *tripHandler {
	return &tripHandler{svc: svc, suggestionSvc: suggestionSvc}
}

func (h *tripHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/trip")
	api.Use(authMiddleware.Authorize())
	api.Post("/", h.CreateTrip)
	api.Post("/join-by-invite-code", h.JoinTripByInviteCode)
	api.Get("/:tripID/plan-trip-bootstrap", h.GetPlanTripBootstrap)
	api.Get("/:tripID/schedule", h.GetTripSchedule)
	api.Post("/:tripID/schedule", h.CreateTripSchedule)
	api.Put("/:tripID/schedule", h.ReplaceTripSchedule)
	api.Post("/:tripID/reschedule", h.RescheduleTrip)
	api.Get("/fairness-report/aggregate", h.GetAggregatedFairnessReport)
	api.Get("/:tripID/fairness-report", h.GetFairnessReport)
	api.Get("/:tripID/plan-trace", h.GetPlanTrace)
	api.Get("/:tripID/reschedule-trace", h.GetReschedulePlanTrace)
	api.Get("/:tripID/publish", h.GetPublishStatus)
	api.Post("/:tripID/publish", h.PublishTrip)
	api.Delete("/:tripID/publish", h.UnpublishTrip)
}

func toScheduleItemDTO(item domain.TripSchedule) dto.TripScheduleItemDTO {
	return dto.TripScheduleItemDTO{
		TripScheduleID: item.TripScheduleID,
		DayNumber:      item.DayNumber,
		SequenceOrder:  item.SequenceOrder,
		PlaceName:      item.PlaceName,
		PlaceID:        item.PlaceID,
		Latitude:       item.Latitude,
		Longitude:      item.Longitude,
		StartTime:      item.StartTime.Format("15:04"),
		EndTime:        item.EndTime.Format("15:04"),
		Type:           item.Type,
	}
}

func toScheduleItemDTOWithPlaceDetail(item domain.TripSchedule, details map[string]domain.PlaceDetailAttachment) dto.TripScheduleItemDTO {
	itemDTO := toScheduleItemDTO(item)
	key := domain.NewGooglePlaceDetailSourceKey(item).CacheKey()
	attachment, ok := details[key]
	if !ok {
		attachment = domain.PlaceDetailAttachment{Status: domain.PlaceDetailStatusUnavailable}
	}
	itemDTO.PlaceDetailStatus = attachment.Status
	itemDTO.PlaceDetail = toPlaceDetailDTO(attachment.Detail)
	return itemDTO
}

func toPlaceDetailDTO(detail *domain.PlaceDetail) *dto.PlaceDetailDTO {
	if detail == nil {
		return nil
	}
	result := &dto.PlaceDetailDTO{
		Rating:           detail.Rating,
		UserRatingCount:  detail.UserRatingCount,
		PhotoURL:         detail.PhotoURL,
		GoogleMapsURI:    detail.GoogleMapsURI,
		EditorialSummary: detail.EditorialSummary,
	}
	if detail.OpeningHours != nil {
		result.OpeningHours = &dto.PlaceDetailOpeningHoursDTO{
			WeekdayText: detail.OpeningHours.WeekdayText,
			OpenNow:     detail.OpeningHours.OpenNow,
		}
	}
	return result
}

func planTripRoleName(role int) string {
	switch role {
	case domain.RoleOwner:
		return "owner"
	case domain.RoleMember:
		return "member"
	case domain.RoleSpectator:
		return "spectator"
	default:
		return "unknown"
	}
}

func optionalProfileImage(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func toScheduleResponseDTO(result *domain.PlanTripBootstrapResult, publishStatus *dto.PublishStatusResponseDTO) dto.GetTripScheduleResponseDTO {
	suggestions := make([]dto.TripScheduleItemDTO, 0, len(result.Suggestions))
	for _, item := range result.Suggestions {
		suggestions = append(suggestions, toScheduleItemDTOWithPlaceDetail(item, result.PlaceDetails))
	}

	days := make([]dto.DayScheduleDTO, 0, len(result.Days))
	for _, day := range result.Days {
		items := make([]dto.TripScheduleItemDTO, 0, len(day.Items))
		for _, item := range day.Items {
			items = append(items, toScheduleItemDTOWithPlaceDetail(item, result.PlaceDetails))
		}
		days = append(days, dto.DayScheduleDTO{
			DayNumber: day.DayNumber,
			Items:     items,
		})
	}

	resp := dto.GetTripScheduleResponseDTO{
		TripID:          result.Trip.TripID,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
		Suggestions:     suggestions,
		Days:            days,
	}
	if publishStatus != nil && publishStatus.IsPublished {
		resp.IsPublished = true
		resp.PublishedTripID = publishStatus.PublishedTripID
	}

	return resp
}

func toPlanTripBootstrapResponseDTO(result *domain.PlanTripBootstrapResult, userID uint) dto.PlanTripBootstrapResponseDTO {
	var publishStatus *dto.PublishStatusResponseDTO
	if result.PublishStatus != nil {
		publishStatus = &dto.PublishStatusResponseDTO{
			IsPublished:     result.PublishStatus.IsPublished,
			PublishedTripID: result.PublishStatus.PublishedTripID,
			Title:           result.PublishStatus.Title,
			Description:     result.PublishStatus.Description,
			ViewCount:       result.PublishStatus.ViewCount,
			LikeCount:       result.PublishStatus.LikeCount,
		}
		if result.PublishStatus.PublishedAt != nil {
			publishStatus.PublishedAt = result.PublishStatus.PublishedAt.Format(time.RFC3339)
		}
	}

	role := result.CurrentMember.Role
	members := make([]dto.PlanTripMemberDTO, 0, len(result.Members))
	waitingMembers := make([]dto.PlanTripWaitingMemberDTO, 0)
	for _, member := range result.Members {
		profileImage := optionalProfileImage(member.ProfileImage)
		members = append(members, dto.PlanTripMemberDTO{
			RoomMemberID:          member.RoomMemberID,
			RoomID:                member.RoomID,
			UserID:                member.UserID,
			Username:              member.Username,
			ProfileImage:          profileImage,
			Role:                  member.Role,
			RoleName:              planTripRoleName(member.Role),
			HasSubmittedLifestyle: member.HasSubmittedLifestyle,
			HasAnalyzedLifestyle:  member.HasAnalyzedLifestyle,
		})
		if member.HasSubmittedLifestyle && !member.HasAnalyzedLifestyle {
			waitingMembers = append(waitingMembers, dto.PlanTripWaitingMemberDTO{
				RoomMemberID: member.RoomMemberID,
				UserID:       member.UserID,
				Username:     member.Username,
				ProfileImage: profileImage,
				LifestyleID:  member.SubmittedLifestyleID,
			})
		}
	}

	readinessStatus := "not_owner"
	if role == domain.RoleOwner {
		if len(waitingMembers) > 0 {
			readinessStatus = "waiting_for_member_analysis"
		} else {
			readinessStatus = "ready_to_reschedule"
		}
	} else {
		waitingMembers = []dto.PlanTripWaitingMemberDTO{}
	}

	return dto.PlanTripBootstrapResponseDTO{
		TripID: result.Trip.TripID,
		RoomID: result.Trip.RoomID,
		CurrentUser: dto.PlanTripCurrentUserDTO{
			UserID:        userID,
			RoomMemberID:  result.CurrentMember.RoomMemberID,
			Role:          role,
			RoleName:      planTripRoleName(role),
			CanEdit:       role == domain.RoleOwner || role == domain.RoleMember,
			CanManageRoom: role == domain.RoleOwner,
		},
		Schedule: toScheduleResponseDTO(result, publishStatus),
		Members:  members,
		RescheduleReadiness: dto.PlanTripRescheduleReadinessDTO{
			Status:         readinessStatus,
			WaitingMembers: waitingMembers,
		},
		PublishStatus: publishStatus,
		Polling: dto.PlanTripPollingDTO{
			SchedulePollAfterMS:          result.SchedulePollAfterMS,
			ScheduleReadinessPollAfterMS: result.ReadinessPollAfterMS,
		},
	}
}

func (h *tripHandler) GetPlanTripBootstrap(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	result, err := h.svc.GetPlanTripBootstrap(c.Context(), userID, uint(tripID))
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get plan trip bootstrap",
			Error:   err.Error(),
		})
	}

	resp := toPlanTripBootstrapResponseDTO(result, userID)

	return c.Status(200).JSON(dto.APIResponse[dto.PlanTripBootstrapResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) GetTripSchedule(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	result, err := h.svc.GetTripSchedule(c.Context(), userID, uint(tripID))
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get trip schedule",
			Error:   err.Error(),
		})
	}

	suggestions := make([]dto.TripScheduleItemDTO, 0, len(result.Suggestions))
	for _, item := range result.Suggestions {
		suggestions = append(suggestions, toScheduleItemDTOWithPlaceDetail(item, result.PlaceDetails))
	}

	days := make([]dto.DayScheduleDTO, 0, len(result.Days))
	for _, day := range result.Days {
		items := make([]dto.TripScheduleItemDTO, 0, len(day.Items))
		for _, item := range day.Items {
			items = append(items, toScheduleItemDTOWithPlaceDetail(item, result.PlaceDetails))
		}
		days = append(days, dto.DayScheduleDTO{
			DayNumber: day.DayNumber,
			Items:     items,
		})
	}

	resp := dto.GetTripScheduleResponseDTO{
		TripID:          result.Trip.TripID,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
		Suggestions:     suggestions,
		Days:            days,
	}

	if pt, err := h.suggestionSvc.GetPublishedTripByTripID(c.Context(), result.Trip.TripID); err == nil {
		resp.IsPublished = true
		id := pt.PublishedTripID
		resp.PublishedTripID = &id
	}

	return c.Status(200).JSON(dto.APIResponse[dto.GetTripScheduleResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) CreateTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	req := new(dto.CreateTripRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "start_date must be in YYYY-MM-DD format",
		})
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "end_date must be in YYYY-MM-DD format",
		})
	}

	preferredDests := make([]domain.PreferredDestination, len(req.PreferredDestinations))
	for i, d := range req.PreferredDestinations {
		preferredDests[i] = domain.PreferredDestination{
			DestinationName: d.DestinationName,
			DestinationID:   d.DestinationID,
			Latitude:        d.Latitude,
			Longitude:       d.Longitude,
		}
	}

	input := domain.CreateTripInput{
		RoomName:              req.RoomName,
		RoomImage:             req.RoomImage,
		DestinationName:       req.DestinationName,
		DestinationID:         req.DestinationID,
		StartDate:             startDate,
		EndDate:               endDate,
		PreferredDestinations: preferredDests,
		TravelVibes:           req.TravelVibes,
		VoyagePriorities:      req.VoyagePriorities,
		FoodVibes:             req.FoodVibes,
		AdditionalNotes:       req.AdditionalNotes,
	}

	result, err := h.svc.CreateTrip(c.Context(), userID, input)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "create trip failed",
			Error:   err.Error(),
		})
	}

	resp := dto.CreateTripResponseDTO{
		RoomID:          result.Room.RoomID,
		TripID:          result.Trip.TripID,
		LifestyleID:     result.Lifestyle.LifestyleID,
		RoomName:        result.Room.RoomName,
		RoomImage:       result.Room.RoomImage,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
	}

	suggestions := make([]dto.TripScheduleItemDTO, 0, len(result.Suggestions))
	for _, item := range result.Suggestions {
		suggestions = append(suggestions, toScheduleItemDTO(item))
	}
	resp.Suggestions = suggestions

	return c.Status(201).JSON(dto.APIResponse[dto.CreateTripResponseDTO]{
		Status:  201,
		Message: "trip created successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) JoinTripByInviteCode(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	req := new(dto.JoinTripByInviteCodeRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	result, err := h.svc.JoinTripByInviteCode(c.Context(), userID, req.InviteCode)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to join trip",
			Error:   err.Error(),
		})
	}

	resp := dto.JoinTripByInviteCodeResponseDTO{
		TripID:          result.Trip.TripID,
		RoomID:          result.Trip.RoomID,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
		RoomMemberID:    result.Member.RoomMemberID,
		UserID:          result.Member.UserID,
		Username:        result.Member.User.Username,
		Role:            result.Member.Role,
		RoleName:        domain.RoomRoleName(result.Member.Role),
		JoinedAt:        result.Member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.JoinTripByInviteCodeResponseDTO]{
		Status:  201,
		Message: "joined trip successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) CreateTripSchedule(c *fiber.Ctx) error {
	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	req := new(dto.CreateTripScheduleRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	inputs := make([]domain.CreateTripScheduleInput, len(req.Items))
	for i, item := range req.Items {
		inputs[i] = domain.CreateTripScheduleInput{
			TripID:        uint(tripID),
			DayNumber:     item.DayNumber,
			SequenceOrder: item.SequenceOrder,
			PlaceName:     item.PlaceName,
			PlaceID:       item.PlaceID,
			Latitude:      item.Latitude,
			Longitude:     item.Longitude,
			StartTime:     item.StartTime,
			EndTime:       item.EndTime,
			Type:          item.Type,
		}
	}

	created, err := h.svc.CreateTripSchedule(c.Context(), inputs)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to create trip schedule",
			Error:   err.Error(),
		})
	}

	result := make([]dto.TripScheduleItemDTO, len(created))
	for i, item := range created {
		result[i] = toScheduleItemDTO(item)
	}

	return c.Status(201).JSON(dto.APIResponse[[]dto.TripScheduleItemDTO]{
		Status:  201,
		Message: "trip schedule created successfully",
		Data:    &result,
	})
}

func (h *tripHandler) ReplaceTripSchedule(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	req := new(dto.CreateTripScheduleRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	inputs := make([]domain.CreateTripScheduleInput, len(req.Items))
	for i, item := range req.Items {
		inputs[i] = domain.CreateTripScheduleInput{
			TripScheduleID: item.TripScheduleID,
			TripID:         uint(tripID),
			DayNumber:      item.DayNumber,
			SequenceOrder:  item.SequenceOrder,
			PlaceName:      item.PlaceName,
			PlaceID:        item.PlaceID,
			Latitude:       item.Latitude,
			Longitude:      item.Longitude,
			StartTime:      item.StartTime,
			EndTime:        item.EndTime,
			Type:           item.Type,
		}
	}

	replaced, err := h.svc.ReplaceTripSchedule(c.Context(), userID, uint(tripID), inputs)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have permission to edit this trip schedule",
			})
		}

		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to replace trip schedule",
			Error:   err.Error(),
		})
	}

	result := make([]dto.TripScheduleItemDTO, len(replaced))
	for i, item := range replaced {
		result[i] = toScheduleItemDTO(item)
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.TripScheduleItemDTO]{
		Status:  200,
		Message: "trip schedule replaced successfully",
		Data:    &result,
	})
}

func (h *tripHandler) RescheduleTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	result, err := h.svc.RescheduleTrip(c.Context(), userID, uint(tripID))
	if err != nil {
		var notReadyErr *domain.RescheduleAnalysisNotReadyError
		if errors.As(err, &notReadyErr) {
			notReady := make([]dto.RescheduleNotReadyMemberDTO, 0, len(notReadyErr.NotReadyMembers))
			for _, item := range notReadyErr.NotReadyMembers {
				notReady = append(notReady, dto.RescheduleNotReadyMemberDTO{
					UserID:      item.UserID,
					Username:    item.Username,
					LifestyleID: item.LifestyleID,
				})
			}

			conflict := dto.RescheduleConflictResponseDTO{
				NotReadyMembers: notReady,
			}
			return c.Status(409).JSON(dto.APIResponse[dto.RescheduleConflictResponseDTO]{
				Status:  409,
				Message: "reschedule blocked: lifestyle analysis is incomplete",
				Data:    &conflict,
				Error:   "analysis_incomplete",
			})
		}

		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "only room owner can reschedule this trip",
			})
		}
		if errors.Is(err, domain.ErrRescheduleConcurrentModification) {
			return c.Status(409).JSON(dto.APIResponse[any]{
				Status:  409,
				Message: "reschedule conflict: another reschedule is currently in progress",
				Error:   err.Error(),
			})
		}

		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to reschedule trip",
			Error:   err.Error(),
		})
	}

	scoreboard := make([]dto.RescheduleTripMemberScoreDTO, 0, len(result.Scoreboard))
	for _, item := range result.Scoreboard {
		scoreboard = append(scoreboard, dto.RescheduleTripMemberScoreDTO{
			UserID:         item.UserID,
			Username:       item.Username,
			Score:          item.Score,
			EffectiveScore: item.EffectiveScore,
			TimesServed:    item.TimesServed,
			DeferredCount:  item.DeferredCount,
		})
	}

	resp := dto.RescheduleTripResponseDTO{
		TripID:           result.TripID,
		ScheduledCount:   result.ScheduledCount,
		SuggestionsCount: result.SuggestionsCount,
		RoundCount:       result.RoundCount,
		SelectedPlaceIDs: result.SelectedPlaceIDs,
		Scoreboard:       scoreboard,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.RescheduleTripResponseDTO]{
		Status:  200,
		Message: "trip rescheduled successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) GetAggregatedFairnessReport(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	report, err := h.svc.GetAggregatedFairnessReport(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to retrieve aggregated fairness report",
			Error:   err.Error(),
		})
	}

	trips := make([]dto.TripFairnessSummaryDTO, len(report.Trips))
	for i, t := range report.Trips {
		trips[i] = dto.TripFairnessSummaryDTO{
			TripID:          t.TripID,
			DestinationName: t.DestinationName,
			GeneratedAt:     t.GeneratedAt,
			RoundCount:      t.RoundCount,
			TotalPlaces:     t.TotalPlaces,
			GiniCoefficient: t.GiniCoefficient,
			FairnessRatio:   t.FairnessRatio,
			ScoreStdDev:     t.ScoreStdDev,
		}
	}

	resp := dto.AggregatedFairnessReportDTO{
		TripCount:        report.TripCount,
		AvgGini:          report.AvgGini,
		AvgFairnessRatio: report.AvgFairnessRatio,
		AvgScoreStdDev:   report.AvgScoreStdDev,
		AvgTotalPlaces:   report.AvgTotalPlaces,
		Trips:            trips,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.AggregatedFairnessReportDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) GetFairnessReport(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	report, err := h.svc.GetFairnessReport(c.Context(), userID, uint(tripID))
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to retrieve fairness report",
			Error:   err.Error(),
		})
	}

	if report == nil {
		return c.Status(404).JSON(dto.APIResponse[any]{
			Status:  404,
			Message: "no fairness report available",
			Error:   "this trip has not been rescheduled yet",
		})
	}

	members := make([]dto.FairnessReportMemberDTO, len(report.Members))
	for i, m := range report.Members {
		members[i] = dto.FairnessReportMemberDTO{
			UserID:         m.UserID,
			Username:       m.Username,
			TimesServed:    m.TimesServed,
			Score:          m.Score,
			EffectiveScore: m.EffectiveScore,
			DeferredCount:  m.DeferredCount,
			ScheduleShare:  m.ScheduleShare,
			DeferredRate:   m.DeferredRate,
		}
	}

	resp := dto.FairnessReportDTO{
		GeneratedAt:      report.GeneratedAt,
		AlgorithmVersion: report.AlgorithmVersion,
		RoundCount:       report.RoundCount,
		TotalPlaces:      report.TotalPlaces,
		GiniCoefficient:  report.GiniCoefficient,
		FairnessRatio:    report.FairnessRatio,
		ScoreStdDev:      report.ScoreStdDev,
		Members:          members,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.FairnessReportDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) GetPlanTrace(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	trace, err := h.svc.GetPlanTrace(c.Context(), userID, uint(tripID))
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get plan trace",
			Error:   err.Error(),
		})
	}

	toPlaceDTO := func(p domain.PlanTracePlace) dto.PlanTracePlaceDTO {
		return dto.PlanTracePlaceDTO{Name: p.Name, PlaceID: p.PlaceID, Latitude: p.Latitude, Longitude: p.Longitude}
	}

	nnSteps := make([]dto.NearestNeighborStepDTO, len(trace.NearestNeighborSteps))
	for i, s := range trace.NearestNeighborSteps {
		nnSteps[i] = dto.NearestNeighborStepDTO{
			Step:       s.Step,
			From:       toPlaceDTO(s.From),
			To:         toPlaceDTO(s.To),
			DistanceKm: s.DistanceKm,
		}
	}

	ordered := make([]dto.PlanTracePlaceDTO, len(trace.OrderedPlaces))
	for i, p := range trace.OrderedPlaces {
		ordered[i] = toPlaceDTO(p)
	}

	aiRecs := make([]dto.PlanTracePlaceDTO, len(trace.AIRecommendations))
	for i, p := range trace.AIRecommendations {
		aiRecs[i] = toPlaceDTO(p)
	}

	scheduled := make([]dto.ScheduledPlaceTraceDTO, len(trace.ScheduledPlaces))
	for i, p := range trace.ScheduledPlaces {
		scheduled[i] = dto.ScheduledPlaceTraceDTO{
			Name: p.Name, PlaceID: p.PlaceID,
			Latitude: p.Latitude, Longitude: p.Longitude,
			DayNumber: p.DayNumber, SequenceOrder: p.SequenceOrder,
		}
	}

	unscheduled := make([]dto.PlanTracePlaceDTO, len(trace.UnscheduledPlaces))
	for i, p := range trace.UnscheduledPlaces {
		unscheduled[i] = toPlaceDTO(p)
	}

	meals := make([]dto.MealSelectionDetailDTO, len(trace.MealSelections))
	for i, m := range trace.MealSelections {
		meals[i] = dto.MealSelectionDetailDTO{
			MealType:      m.MealType,
			DayNumber:     m.DayNumber,
			SequenceOrder: m.SequenceOrder,
			AnchorPlace:   toPlaceDTO(m.AnchorPlace),
			SelectedPlace: toPlaceDTO(m.SelectedPlace),
			DistanceKm:    m.DistanceKm,
		}
	}

	resp := dto.PlanTraceResponseDTO{
		TripID:                    trace.TripID,
		DestinationName:           trace.DestinationName,
		StartDate:                 trace.StartDate,
		EndDate:                   trace.EndDate,
		TotalDays:                 trace.TotalDays,
		PlacesPerDay:              trace.PlacesPerDay,
		Step1AIRecommendations:    aiRecs,
		Step2NearestNeighborSteps: nnSteps,
		Step2OrderedPlaces:        ordered,
		Step3ScheduledPlaces:      scheduled,
		Step3UnscheduledPlaces:    unscheduled,
		Step4MealSelections:       meals,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.PlanTraceResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) GetReschedulePlanTrace(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	trace, err := h.svc.GetReschedulePlanTrace(c.Context(), userID, uint(tripID))
	if err != nil {
		var notReadyErr *domain.RescheduleAnalysisNotReadyError
		if errors.As(err, &notReadyErr) {
			notReady := make([]dto.RescheduleNotReadyMemberDTO, 0, len(notReadyErr.NotReadyMembers))
			for _, item := range notReadyErr.NotReadyMembers {
				notReady = append(notReady, dto.RescheduleNotReadyMemberDTO{
					UserID:      item.UserID,
					Username:    item.Username,
					LifestyleID: item.LifestyleID,
				})
			}
			conflict := dto.RescheduleConflictResponseDTO{NotReadyMembers: notReady}
			return c.Status(409).JSON(dto.APIResponse[dto.RescheduleConflictResponseDTO]{
				Status:  409,
				Message: "trace unavailable: lifestyle analysis is incomplete",
				Data:    &conflict,
				Error:   "analysis_incomplete",
			})
		}
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get reschedule plan trace",
			Error:   err.Error(),
		})
	}

	toPlaceDTO := func(p domain.PlanTracePlace) dto.PlanTracePlaceDTO {
		return dto.PlanTracePlaceDTO{Name: p.Name, PlaceID: p.PlaceID, Latitude: p.Latitude, Longitude: p.Longitude}
	}
	toCandidateDTO := func(c domain.RescheduleCandidateTrace) dto.RescheduleCandidateTraceDTO {
		return dto.RescheduleCandidateTraceDTO{Name: c.Name, PlaceID: c.PlaceID, Category: c.Category, Latitude: c.Latitude, Longitude: c.Longitude}
	}

	// Step 1
	members := make([]dto.RescheduleMemberCandidateTraceDTO, len(trace.Step1Members))
	for i, m := range trace.Step1Members {
		cands := make([]dto.RescheduleCandidateTraceDTO, len(m.Candidates))
		for j, c := range m.Candidates {
			cands[j] = toCandidateDTO(c)
		}
		members[i] = dto.RescheduleMemberCandidateTraceDTO{
			UserID: m.UserID, Username: m.Username,
			Candidates:   cands,
			CategoryRank: m.CategoryRank,
		}
	}
	var centroidDTO *dto.GeoPointTraceDTO
	if trace.Step1Centroid != nil {
		centroidDTO = &dto.GeoPointTraceDTO{Latitude: trace.Step1Centroid.Latitude, Longitude: trace.Step1Centroid.Longitude}
	}

	// Step 2
	rounds := make([]dto.FairnessRoundTraceDTO, len(trace.Step2FairnessRounds))
	for i, r := range trace.Step2FairnessRounds {
		var selPlace *dto.RescheduleCandidateTraceDTO
		if r.SelectedPlace != nil {
			cp := toCandidateDTO(*r.SelectedPlace)
			selPlace = &cp
		}
		updates := make([]dto.RescheduleScoreUpdateTraceDTO, len(r.ScoreUpdates))
		for j, u := range r.ScoreUpdates {
			updates[j] = dto.RescheduleScoreUpdateTraceDTO{
				UserID: u.UserID, Username: u.Username, Gained: u.Gained, Reason: u.Reason,
				OldScore: u.OldScore, NewScore: u.NewScore,
			}
		}
		states := make([]dto.RescheduleMemberStateTraceDTO, len(r.MemberStatesAfter))
		for j, s := range r.MemberStatesAfter {
			states[j] = dto.RescheduleMemberStateTraceDTO{
				UserID: s.UserID, Username: s.Username, Score: s.Score,
				EffectiveScore: s.EffectiveScore, TimesServed: s.TimesServed, DeferredCount: s.DeferredCount,
			}
		}
		rounds[i] = dto.FairnessRoundTraceDTO{
			Round:                  r.Round,
			PickedMemberID:         r.PickedMemberID,
			PickedMemberUsername:   r.PickedMemberUsername,
			EffectiveScoreBefore:   r.EffectiveScoreBefore,
			IsDeferred:             r.IsDeferred,
			DeferReason:            r.DeferReason,
			SelectedPlace:          selPlace,
			DistanceFromPrevKm:     r.DistanceFromPrevKm,
			DistanceFromCentroidKm: r.DistanceFromCentroidKm,
			ScoreUpdates:           updates,
			MemberStatesAfter:      states,
		}
	}
	fairnessOrdered := make([]dto.RescheduleCandidateTraceDTO, len(trace.Step2FairnessOrderedPlaces))
	for i, c := range trace.Step2FairnessOrderedPlaces {
		fairnessOrdered[i] = toCandidateDTO(c)
	}

	// Step 3
	nnSteps := make([]dto.NearestNeighborStepDTO, len(trace.Step3NearestNeighborSteps))
	for i, s := range trace.Step3NearestNeighborSteps {
		nnSteps[i] = dto.NearestNeighborStepDTO{
			Step:       s.Step,
			From:       dto.PlanTracePlaceDTO{Name: s.From.Name, PlaceID: s.From.PlaceID, Latitude: s.From.Latitude, Longitude: s.From.Longitude},
			To:         dto.PlanTracePlaceDTO{Name: s.To.Name, PlaceID: s.To.PlaceID, Latitude: s.To.Latitude, Longitude: s.To.Longitude},
			DistanceKm: s.DistanceKm,
		}
	}
	nnOrdered := make([]dto.PlanTracePlaceDTO, len(trace.Step3OrderedPlaces))
	for i, p := range trace.Step3OrderedPlaces {
		nnOrdered[i] = toPlaceDTO(p)
	}

	// Step 4
	scheduled := make([]dto.ScheduledPlaceTraceDTO, len(trace.Step4ScheduledPlaces))
	for i, p := range trace.Step4ScheduledPlaces {
		scheduled[i] = dto.ScheduledPlaceTraceDTO{
			Name: p.Name, PlaceID: p.PlaceID, Latitude: p.Latitude, Longitude: p.Longitude,
			DayNumber: p.DayNumber, SequenceOrder: p.SequenceOrder,
		}
	}
	unscheduled := make([]dto.PlanTracePlaceDTO, len(trace.Step4UnscheduledPlaces))
	for i, p := range trace.Step4UnscheduledPlaces {
		unscheduled[i] = toPlaceDTO(p)
	}

	// Step 5
	meals := make([]dto.MealSelectionDetailDTO, len(trace.Step5MealSelections))
	for i, m := range trace.Step5MealSelections {
		meals[i] = dto.MealSelectionDetailDTO{
			MealType: m.MealType, DayNumber: m.DayNumber, SequenceOrder: m.SequenceOrder,
			AnchorPlace:   toPlaceDTO(m.AnchorPlace),
			SelectedPlace: toPlaceDTO(m.SelectedPlace),
			DistanceKm:    m.DistanceKm,
		}
	}

	resp := dto.ReschedulePlanTraceResponseDTO{
		TripID: trace.TripID, DestinationName: trace.DestinationName,
		StartDate: trace.StartDate, EndDate: trace.EndDate,
		TotalDays: trace.TotalDays, PlacesPerDay: trace.PlacesPerDay,
		Step1MembersAndCandidates:  members,
		Step1Centroid:              centroidDTO,
		Step2FairnessRounds:        rounds,
		Step2FairnessOrderedPlaces: fairnessOrdered,
		Step2TotalRounds:           trace.Step2TotalRounds,
		Step3NearestNeighborSteps:  nnSteps,
		Step3OrderedPlaces:         nnOrdered,
		Step4ScheduledPlaces:       scheduled,
		Step4UnscheduledPlaces:     unscheduled,
		Step5MealSelections:        meals,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.ReschedulePlanTraceResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) GetPublishStatus(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	inRoom, err := h.svc.GetTripSchedule(c.Context(), userID, uint(tripID))
	if err != nil {
		if err.Error() == "forbidden" {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get trip",
			Error:   err.Error(),
		})
	}
	_ = inRoom

	pt, err := h.suggestionSvc.GetPublishedTripByTripID(c.Context(), uint(tripID))
	if err != nil {
		return c.Status(200).JSON(dto.APIResponse[dto.PublishStatusResponseDTO]{
			Status:  200,
			Message: "success",
			Data: &dto.PublishStatusResponseDTO{
				IsPublished: false,
			},
		})
	}

	resp := dto.PublishStatusResponseDTO{
		IsPublished:     true,
		PublishedTripID: &pt.PublishedTripID,
		Title:           pt.Title,
		Description:     pt.Description,
		ViewCount:       pt.ViewCount,
		LikeCount:       pt.LikeCount,
		PublishedAt:     pt.CreatedAt.Format(time.RFC3339),
	}

	return c.Status(200).JSON(dto.APIResponse[dto.PublishStatusResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) PublishTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	req := new(dto.PublishTripRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	pt, err := h.suggestionSvc.PublishTrip(c.Context(), uint(tripID), userID, req.Title, req.Description)
	if err != nil {
		if err.Error() == "forbidden" {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "only the trip owner can publish",
			})
		}
		if err.Error() == "trip already published" {
			return c.Status(409).JSON(dto.APIResponse[any]{
				Status:  409,
				Message: "conflict",
				Error:   "trip already published",
			})
		}
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to publish trip",
			Error:   err.Error(),
		})
	}

	resp := dto.PublishTripResponseDTO{
		PublishedTripID: pt.PublishedTripID,
		TripID:          pt.TripID,
		Title:           pt.Title,
		Description:     pt.Description,
		PublishedAt:     pt.CreatedAt.Format(time.RFC3339),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.PublishTripResponseDTO]{
		Status:  201,
		Message: "trip published successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) UnpublishTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	if err := h.suggestionSvc.UnpublishTrip(c.Context(), uint(tripID), userID); err != nil {
		if err.Error() == "forbidden" {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "only the publisher can unpublish",
			})
		}
		if err.Error() == "trip is not published" {
			return c.Status(404).JSON(dto.APIResponse[any]{
				Status:  404,
				Message: "not found",
				Error:   "trip is not published",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to unpublish trip",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[any]{
		Status:  200,
		Message: "trip unpublished successfully",
	})
}

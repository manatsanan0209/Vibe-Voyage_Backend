package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type tripSuggestionRepository struct {
	db *gorm.DB
}

func NewTripSuggestionRepository(db *gorm.DB) domain.TripSuggestionRepository {
	return &tripSuggestionRepository{db: db}
}

func (r *tripSuggestionRepository) GetPublishedTrips(ctx context.Context, opts domain.GetPublishedTripsOptions) ([]domain.PublishedTripWithMeta, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&domain.PublishedTrip{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var published []domain.PublishedTrip
	offset := (opts.Page - 1) * opts.Limit
	if err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("User").
		Order("view_count DESC, created_at DESC").
		Offset(offset).
		Limit(opts.Limit).
		Find(&published).Error; err != nil {
		return nil, 0, err
	}

	if len(published) == 0 {
		return []domain.PublishedTripWithMeta{}, total, nil
	}

	ids := make([]uint, len(published))
	for i, p := range published {
		ids[i] = p.PublishedTripID
	}

	likedSet := map[uint]bool{}
	bookmarkedSet := map[uint]bool{}
	if opts.UserID > 0 {
		var likedIDs []uint
		r.db.WithContext(ctx).Model(&domain.TripLike{}).
			Where("published_trip_id IN ? AND user_id = ?", ids, opts.UserID).
			Pluck("published_trip_id", &likedIDs)
		for _, id := range likedIDs {
			likedSet[id] = true
		}

		var bookmarkedIDs []uint
		r.db.WithContext(ctx).Model(&domain.TripBookmark{}).
			Where("published_trip_id IN ? AND user_id = ?", ids, opts.UserID).
			Pluck("published_trip_id", &bookmarkedIDs)
		for _, id := range bookmarkedIDs {
			bookmarkedSet[id] = true
		}
	}

	result := make([]domain.PublishedTripWithMeta, 0, len(published))
	for i := range published {
		result = append(result, domain.PublishedTripWithMeta{
			PublishedTrip:  &published[i],
			Trip:           &published[i].Trip,
			PublisherName:  published[i].User.Username,
			PublisherImage: published[i].User.ProfileImage,
			IsLiked:        likedSet[published[i].PublishedTripID],
			IsBookmarked:   bookmarkedSet[published[i].PublishedTripID],
		})
	}

	return result, total, nil
}

func (r *tripSuggestionRepository) GetPublishedTripByID(ctx context.Context, publishedTripID, userID uint) (*domain.PublishedTripWithMeta, error) {
	var pt domain.PublishedTrip
	if err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("User").
		Where("published_trip_id = ?", publishedTripID).
		First(&pt).Error; err != nil {
		return nil, err
	}

	var schedules []domain.TripSchedule
	r.db.WithContext(ctx).
		Where("trip_id = ? AND day_number >= 1", pt.TripID).
		Order("day_number ASC, sequence_order ASC").
		Find(&schedules)

	dayMap := make(map[int][]domain.TripSchedule)
	for _, s := range schedules {
		dayMap[s.DayNumber] = append(dayMap[s.DayNumber], s)
	}
	days := make([]domain.DaySchedule, 0, len(dayMap))
	for dayNum, items := range dayMap {
		days = append(days, domain.DaySchedule{DayNumber: dayNum, Items: items})
	}
	sort.Slice(days, func(i, j int) bool { return days[i].DayNumber < days[j].DayNumber })

	meta := &domain.PublishedTripWithMeta{
		PublishedTrip:  &pt,
		Trip:           &pt.Trip,
		PublisherName:  pt.User.Username,
		PublisherImage: pt.User.ProfileImage,
		ScheduleDays:   days,
	}

	if userID > 0 {
		var likeCount int64
		r.db.WithContext(ctx).Model(&domain.TripLike{}).
			Where("published_trip_id = ? AND user_id = ?", publishedTripID, userID).
			Count(&likeCount)
		meta.IsLiked = likeCount > 0

		var bmCount int64
		r.db.WithContext(ctx).Model(&domain.TripBookmark{}).
			Where("published_trip_id = ? AND user_id = ?", publishedTripID, userID).
			Count(&bmCount)
		meta.IsBookmarked = bmCount > 0
	}

	return meta, nil
}

func (r *tripSuggestionRepository) GetPublishedTripByTripID(ctx context.Context, tripID uint) (*domain.PublishedTrip, error) {
	var pt domain.PublishedTrip
	if err := r.db.WithContext(ctx).Where("trip_id = ?", tripID).First(&pt).Error; err != nil {
		return nil, err
	}
	return &pt, nil
}

func (r *tripSuggestionRepository) PublishTrip(ctx context.Context, tripID, userID uint, title, description string) (*domain.PublishedTrip, error) {
	var trip domain.Trips
	if err := r.db.WithContext(ctx).First(&trip, tripID).Error; err != nil {
		return nil, errors.New("trip not found")
	}

	var room domain.Room
	if err := r.db.WithContext(ctx).First(&room, trip.RoomID).Error; err != nil {
		return nil, errors.New("room not found")
	}

	if room.OwnerID != userID {
		return nil, errors.New("forbidden")
	}

	var existing domain.PublishedTrip
	err := r.db.WithContext(ctx).Where("trip_id = ?", tripID).First(&existing).Error
	if err == nil {
		return nil, errors.New("trip already published")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if title == "" {
		title = trip.DestinationName
	}

	pt := &domain.PublishedTrip{
		TripID:      tripID,
		UserID:      userID,
		Title:       title,
		Description: description,
	}
	if err := r.db.WithContext(ctx).Create(pt).Error; err != nil {
		return nil, err
	}

	return pt, nil
}

func (r *tripSuggestionRepository) UnpublishTrip(ctx context.Context, tripID, userID uint) error {
	var pt domain.PublishedTrip
	if err := r.db.WithContext(ctx).Where("trip_id = ?", tripID).First(&pt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("trip is not published")
		}
		return err
	}

	if pt.UserID != userID {
		return errors.New("forbidden")
	}

	return r.db.WithContext(ctx).Unscoped().Delete(&pt).Error
}

func (r *tripSuggestionRepository) IncrementViewCount(ctx context.Context, publishedTripID uint) error {
	return r.db.WithContext(ctx).Model(&domain.PublishedTrip{}).
		Where("published_trip_id = ?", publishedTripID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *tripSuggestionRepository) ToggleLike(ctx context.Context, publishedTripID, userID uint) (liked bool, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing domain.TripLike
		txErr := tx.Where("published_trip_id = ? AND user_id = ?", publishedTripID, userID).First(&existing).Error

		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			newLike := domain.TripLike{PublishedTripID: publishedTripID, UserID: userID}
			if createErr := tx.Create(&newLike).Error; createErr != nil {
				return createErr
			}
			tx.Model(&domain.PublishedTrip{}).Where("published_trip_id = ?", publishedTripID).
				UpdateColumn("like_count", gorm.Expr("like_count + 1"))
			liked = true
			return nil
		} else if txErr != nil {
			return txErr
		}

		if deleteErr := tx.Unscoped().Delete(&existing).Error; deleteErr != nil {
			return deleteErr
		}
		tx.Model(&domain.PublishedTrip{}).Where("published_trip_id = ?", publishedTripID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
		liked = false
		return nil
	})
	return liked, err
}

func (r *tripSuggestionRepository) ToggleBookmark(ctx context.Context, publishedTripID, userID uint) (bookmarked bool, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing domain.TripBookmark
		txErr := tx.Where("published_trip_id = ? AND user_id = ?", publishedTripID, userID).First(&existing).Error

		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			newBm := domain.TripBookmark{PublishedTripID: publishedTripID, UserID: userID}
			if createErr := tx.Create(&newBm).Error; createErr != nil {
				return createErr
			}
			bookmarked = true
			return nil
		} else if txErr != nil {
			return txErr
		}

		if deleteErr := tx.Unscoped().Delete(&existing).Error; deleteErr != nil {
			return deleteErr
		}
		bookmarked = false
		return nil
	})
	return bookmarked, err
}

func (r *tripSuggestionRepository) GetBookmarkedTrips(ctx context.Context, userID uint) ([]domain.PublishedTripWithMeta, error) {
	var bookmarks []domain.TripBookmark
	r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&bookmarks)

	if len(bookmarks) == 0 {
		return []domain.PublishedTripWithMeta{}, nil
	}

	publishedTripIDs := make([]uint, len(bookmarks))
	for i, b := range bookmarks {
		publishedTripIDs[i] = b.PublishedTripID
	}

	var published []domain.PublishedTrip
	r.db.WithContext(ctx).
		Preload("Trip").
		Preload("User").
		Where("published_trip_id IN ?", publishedTripIDs).
		Find(&published)

	var likedIDs []uint
	r.db.WithContext(ctx).Model(&domain.TripLike{}).
		Where("published_trip_id IN ? AND user_id = ?", publishedTripIDs, userID).
		Pluck("published_trip_id", &likedIDs)
	likedSet := make(map[uint]bool)
	for _, id := range likedIDs {
		likedSet[id] = true
	}

	result := make([]domain.PublishedTripWithMeta, 0, len(published))
	for i := range published {
		result = append(result, domain.PublishedTripWithMeta{
			PublishedTrip:  &published[i],
			Trip:           &published[i].Trip,
			PublisherName:  published[i].User.Username,
			PublisherImage: published[i].User.ProfileImage,
			IsLiked:        likedSet[published[i].PublishedTripID],
			IsBookmarked:   true,
		})
	}

	return result, nil
}

func (r *tripSuggestionRepository) UseAsTemplate(ctx context.Context, publishedTripID, userID uint, input domain.UseAsTemplateInput) (*domain.CreateTripResult, error) {
	var pt domain.PublishedTrip
	if err := r.db.WithContext(ctx).
		Preload("Trip").
		Where("published_trip_id = ?", publishedTripID).
		First(&pt).Error; err != nil {
		return nil, errors.New("published trip not found")
	}

	var memberCount int64
	r.db.WithContext(ctx).
		Table("room_members rm").
		Joins("JOIN trips t ON t.room_id = rm.room_id AND t.deleted_at IS NULL").
		Where("rm.user_id = ? AND rm.deleted_at IS NULL AND t.trip_id = ?", userID, pt.TripID).
		Count(&memberCount)
	if memberCount > 0 {
		return nil, errors.New("cannot use own trip as template")
	}

	templateDays := int(pt.Trip.EndDate.Sub(pt.Trip.StartDate).Hours() / 24)
	newDays := int(input.EndDate.Sub(input.StartDate).Hours() / 24)
	if newDays < templateDays {
		return nil, fmt.Errorf("trip duration must be at least %d days to match the template", templateDays)
	}

	var schedules []domain.TripSchedule
	r.db.WithContext(ctx).
		Where("trip_id = ? AND day_number >= 1", pt.TripID).
		Order("day_number ASC, sequence_order ASC").
		Find(&schedules)

	var result domain.CreateTripResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		room := &domain.Room{
			OwnerID:   userID,
			RoomName:  input.RoomName,
			RoomImage: input.RoomImage,
		}
		if err := tx.Create(room).Error; err != nil {
			return err
		}

		trip := &domain.Trips{
			RoomID:          room.RoomID,
			DestinationName: pt.Trip.DestinationName,
			DestinationID:   pt.Trip.DestinationID,
			StartDate:       input.StartDate,
			EndDate:         input.EndDate,
		}
		if err := tx.Create(trip).Error; err != nil {
			return err
		}

		member := &domain.RoomMember{
			RoomID: room.RoomID,
			UserID: userID,
			Role:   domain.RoleOwner,
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		if len(schedules) > 0 {
			newSchedules := make([]domain.TripSchedule, len(schedules))
			for i, s := range schedules {
				newSchedules[i] = domain.TripSchedule{
					TripID:        trip.TripID,
					DayNumber:     s.DayNumber,
					SequenceOrder: s.SequenceOrder,
					PlaceName:     s.PlaceName,
					PlaceID:       s.PlaceID,
					Latitude:      s.Latitude,
					Longitude:     s.Longitude,
					StartTime:     s.StartTime,
					EndTime:       s.EndTime,
					Type:          s.Type,
				}
			}
			if err := tx.Create(&newSchedules).Error; err != nil {
				return err
			}
		}

		result = domain.CreateTripResult{
			Room:   room,
			Trip:   trip,
			Member: member,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

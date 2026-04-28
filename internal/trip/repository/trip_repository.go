package repository

import (
	"context"
	"strings"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type tripRepository struct {
	db *gorm.DB
}

func NewTripRepository(db *gorm.DB) domain.TripRepository {
	return &tripRepository{db: db}
}

func (r *tripRepository) GetByID(ctx context.Context, tripID uint) (*domain.Trips, error) {
	var trip domain.Trips
	if err := r.db.WithContext(ctx).First(&trip, tripID).Error; err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepository) GetByRoomID(ctx context.Context, roomID uint) (*domain.Trips, error) {
	var trip domain.Trips
	if err := r.db.WithContext(ctx).Where("room_id = ?", roomID).First(&trip).Error; err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepository) IsUserInTripRoom(ctx context.Context, userID, tripID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("room_members rm").
		Joins("JOIN trips t ON t.room_id = rm.room_id AND t.deleted_at IS NULL").
		Where("rm.user_id = ? AND rm.deleted_at IS NULL AND t.trip_id = ?", userID, tripID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *tripRepository) GetUserRoleInTripRoom(ctx context.Context, userID, tripID uint) (int, bool, error) {
	var row struct {
		Role int
	}

	err := r.db.WithContext(ctx).
		Table("room_members rm").
		Select("rm.role").
		Joins("JOIN trips t ON t.room_id = rm.room_id AND t.deleted_at IS NULL").
		Where("rm.user_id = ? AND rm.deleted_at IS NULL AND t.trip_id = ?", userID, tripID).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return 0, false, err
	}

	if row.Role == 0 {
		return 0, false, nil
	}

	return row.Role, true, nil
}

func (r *tripRepository) GetSchedulesByTripID(ctx context.Context, tripID uint) ([]domain.TripSchedule, error) {
	var schedules []domain.TripSchedule
	if err := r.db.WithContext(ctx).
		Where("trip_id = ?", tripID).
		Order("day_number ASC, sequence_order ASC").
		Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *tripRepository) UpdateGroupStructuredLifestyle(ctx context.Context, tripID uint, snapshot string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Trips{}).
		Where("trip_id = ?", tripID).
		Update("group_structured_lifestyle", snapshot).Error
}

func (r *tripRepository) GetAttractionsByNames(ctx context.Context, names []string) (map[string][]domain.Attraction, error) {
	normalizedNames := make([]string, 0, len(names))
	for _, name := range names {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized == "" {
			continue
		}
		exists := false
		for _, existing := range normalizedNames {
			if existing == normalized {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		normalizedNames = append(normalizedNames, normalized)
	}

	result := make(map[string][]domain.Attraction, len(normalizedNames))
	if len(normalizedNames) == 0 {
		return result, nil
	}

	query := r.db.WithContext(ctx).Model(&domain.Attraction{})
	for idx, normalizedName := range normalizedNames {
		like := "%" + normalizedName + "%"
		if idx == 0 {
			query = query.Where("LOWER(name_th) LIKE ? OR LOWER(name_en) LIKE ?", like, like)
			continue
		}
		query = query.Or("LOWER(name_th) LIKE ? OR LOWER(name_en) LIKE ?", like, like)
	}

	attractions := make([]domain.Attraction, 0)
	if err := query.Find(&attractions).Error; err != nil {
		return nil, err
	}

	for _, attraction := range attractions {
		nameTH := strings.ToLower(strings.TrimSpace(attraction.NameTH))
		nameEN := strings.ToLower(strings.TrimSpace(attraction.NameEN))
		for _, normalizedName := range normalizedNames {
			if strings.Contains(nameTH, normalizedName) || strings.Contains(nameEN, normalizedName) {
				result[normalizedName] = append(result[normalizedName], attraction)
			}
		}
	}

	return result, nil
}

func (r *tripRepository) CreateTripBundle(
	ctx context.Context,
	userID uint,
	input domain.CreateTripInput,
	preferredDestinationsJSON string,
	travelVibesJSON string,
	voyagePrioritiesJSON string,
	foodVibesJSON string,
) (*domain.CreateTripResult, error) {
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
			DestinationName: input.DestinationName,
			DestinationID:   input.DestinationID,
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

		lifestyle := &domain.UserLifestyle{
			UserID:                userID,
			RoomID:                room.RoomID,
			PreferredDestinations: preferredDestinationsJSON,
			TravelVibes:           travelVibesJSON,
			VoyagePriorities:      voyagePrioritiesJSON,
			FoodVibes:             foodVibesJSON,
			AdditionalNotes:       input.AdditionalNotes,
		}
		if err := tx.Create(lifestyle).Error; err != nil {
			return err
		}

		var preferredSchedules []domain.TripSchedule
		for _, dest := range input.PreferredDestinations {
			preferredSchedules = append(preferredSchedules, domain.TripSchedule{
				TripID:        trip.TripID,
				DayNumber:     0,
				SequenceOrder: 0,
				PlaceName:     dest.DestinationName,
				PlaceID:       dest.DestinationID,
				Latitude:      dest.Latitude,
				Longitude:     dest.Longitude,
				Type:          "preferred_destination",
			})
		}
		if len(preferredSchedules) > 0 {
			if err := tx.Create(&preferredSchedules).Error; err != nil {
				return err
			}
		}

		result = domain.CreateTripResult{
			Room:        room,
			Trip:        trip,
			Member:      member,
			Lifestyle:   lifestyle,
			Suggestions: preferredSchedules,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *tripRepository) CreateSchedules(ctx context.Context, schedules []domain.TripSchedule) error {
	if len(schedules) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&schedules).Error
}

func (r *tripRepository) ReplaceSchedulesByTripID(ctx context.Context, tripID uint, schedules []domain.TripSchedule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingIDs []uint
		if err := tx.Model(&domain.TripSchedule{}).
			Where("trip_id = ?", tripID).
			Pluck("trip_schedule_id", &existingIDs).Error; err != nil {
			return err
		}

		incomingIDs := map[uint]bool{}
		for _, s := range schedules {
			if s.TripScheduleID > 0 {
				incomingIDs[s.TripScheduleID] = true
			}
		}

		toDelete := []uint{}
		for _, id := range existingIDs {
			if !incomingIDs[id] {
				toDelete = append(toDelete, id)
			}
		}
		if len(toDelete) > 0 {
			if err := tx.Unscoped().Delete(&domain.TripSchedule{}, toDelete).Error; err != nil {
				return err
			}
		}

		for i := range schedules {
			if schedules[i].TripScheduleID > 0 {
				if err := tx.Model(&schedules[i]).Updates(map[string]interface{}{
					"day_number":     schedules[i].DayNumber,
					"sequence_order": schedules[i].SequenceOrder,
					"place_name":     schedules[i].PlaceName,
					"place_id":       schedules[i].PlaceID,
					"latitude":       schedules[i].Latitude,
					"longitude":      schedules[i].Longitude,
					"start_time":     schedules[i].StartTime,
					"end_time":       schedules[i].EndTime,
					"type":           schedules[i].Type,
				}).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Create(&schedules[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *tripRepository) ReplaceScheduleAndSnapshot(ctx context.Context, tripID uint, schedules []domain.TripSchedule, snapshot string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("trip_id = ?", tripID).Delete(&domain.TripSchedule{}).Error; err != nil {
			return err
		}
		if len(schedules) > 0 {
			if err := tx.Create(&schedules).Error; err != nil {
				return err
			}
		}
		return tx.Model(&domain.Trips{}).
			Where("trip_id = ?", tripID).
			Updates(map[string]interface{}{
				"group_structured_lifestyle": snapshot,
				"version":                    gorm.Expr("version + 1"),
			}).Error
	})
}

package repository

import (
	"context"

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

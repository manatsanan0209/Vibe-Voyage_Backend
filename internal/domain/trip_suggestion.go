package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type PublishedTrip struct {
	PublishedTripID uint           `json:"published_trip_id" gorm:"primaryKey;autoIncrement"`
	TripID          uint           `json:"trip_id" gorm:"not null;uniqueIndex"`
	Trip            Trips          `json:"-" gorm:"foreignKey:TripID;references:TripID"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`
	User            User           `json:"-" gorm:"foreignKey:UserID;references:UserID"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	ViewCount       int64          `json:"view_count" gorm:"default:0"`
	LikeCount       int64          `json:"like_count" gorm:"default:0"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type TripLike struct {
	TripLikeID      uint      `json:"trip_like_id" gorm:"primaryKey;autoIncrement"`
	PublishedTripID uint      `json:"published_trip_id" gorm:"not null;uniqueIndex:idx_trip_like_user"`
	UserID          uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_trip_like_user"`
	CreatedAt       time.Time `json:"created_at"`
}

type TripBookmark struct {
	TripBookmarkID  uint      `json:"trip_bookmark_id" gorm:"primaryKey;autoIncrement"`
	PublishedTripID uint      `json:"published_trip_id" gorm:"not null;uniqueIndex:idx_trip_bookmark_user"`
	UserID          uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_trip_bookmark_user"`
	CreatedAt       time.Time `json:"created_at"`
}

type GetPublishedTripsOptions struct {
	Page   int
	Limit  int
	UserID uint
}

type PublishedTripWithMeta struct {
	PublishedTrip  *PublishedTrip
	Trip           *Trips
	PublisherName  string
	PublisherImage string
	ScheduleDays   []DaySchedule
	IsLiked        bool
	IsBookmarked   bool
}

type UseAsTemplateInput struct {
	RoomName  string
	RoomImage string
	StartDate time.Time
	EndDate   time.Time
}

type TripSuggestionRepository interface {
	GetPublishedTrips(ctx context.Context, opts GetPublishedTripsOptions) ([]PublishedTripWithMeta, int64, error)
	GetPublishedTripByID(ctx context.Context, publishedTripID, userID uint) (*PublishedTripWithMeta, error)
	GetPublishedTripByTripID(ctx context.Context, tripID uint) (*PublishedTrip, error)
	PublishTrip(ctx context.Context, tripID, userID uint, title, description string) (*PublishedTrip, error)
	UnpublishTrip(ctx context.Context, tripID, userID uint) error
	IncrementViewCount(ctx context.Context, publishedTripID uint) error
	ToggleLike(ctx context.Context, publishedTripID, userID uint) (bool, error)
	ToggleBookmark(ctx context.Context, publishedTripID, userID uint) (bool, error)
	GetBookmarkedTrips(ctx context.Context, userID uint) ([]PublishedTripWithMeta, error)
	UseAsTemplate(ctx context.Context, publishedTripID, userID uint, input UseAsTemplateInput) (*CreateTripResult, error)
}

type TripSuggestionService interface {
	GetFeed(ctx context.Context, page, limit int, userID uint) ([]PublishedTripWithMeta, int64, error)
	GetDetail(ctx context.Context, publishedTripID, userID uint) (*PublishedTripWithMeta, error)
	GetPublishedTripByTripID(ctx context.Context, tripID uint) (*PublishedTrip, error)
	PublishTrip(ctx context.Context, tripID, userID uint, title, description string) (*PublishedTrip, error)
	UnpublishTrip(ctx context.Context, tripID, userID uint) error
	ToggleLike(ctx context.Context, publishedTripID, userID uint) (bool, error)
	ToggleBookmark(ctx context.Context, publishedTripID, userID uint) (bool, error)
	GetBookmarks(ctx context.Context, userID uint) ([]PublishedTripWithMeta, error)
	UseAsTemplate(ctx context.Context, publishedTripID, userID uint, input UseAsTemplateInput) (*CreateTripResult, error)
}

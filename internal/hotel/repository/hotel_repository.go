package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgHotelRepository struct {
	db *gorm.DB
}

func NewHotelRepository(db *gorm.DB) domain.HotelRepository {
	return &pgHotelRepository{db: db}
}

func (r *pgHotelRepository) GetByID(ctx context.Context, id string) (*domain.Hotel, error) {
	var hotel domain.Hotel
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Preload("Province").
		Preload("District").
		Preload("Subdistrict").
		Preload("AccommodationType").
		Preload("PriceRange").
		First(&hotel, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &hotel, nil
}

func (r *pgHotelRepository) List(ctx context.Context, filter domain.HotelFilter) ([]*domain.Hotel, int64, error) {
	var hotels []*domain.Hotel
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Hotel{})

	if filter.ProvinceID != "" {
		query = query.Where("province_id = ?", filter.ProvinceID)
	}
	if filter.DistrictID != "" {
		query = query.Where("district_id = ?", filter.DistrictID)
	}
	if filter.AccomTypeID != "" {
		query = query.Where("accom_type_id = ?", filter.AccomTypeID)
	}
	if filter.PriceRangeID != "" {
		query = query.Where("price_range_id = ?", filter.PriceRangeID)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("name_th ILIKE ? OR name_en ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.
		Preload("Region").
		Preload("Province").
		Preload("District").
		Preload("Subdistrict").
		Preload("AccommodationType").
		Preload("PriceRange").
		Find(&hotels).Error; err != nil {
		return nil, 0, err
	}

	return hotels, total, nil
}

func (r *pgHotelRepository) GetAccommodationTypes(ctx context.Context) ([]*domain.AccommodationType, error) {
	var types []*domain.AccommodationType
	if err := r.db.WithContext(ctx).Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

func (r *pgHotelRepository) GetPriceRanges(ctx context.Context) ([]*domain.PriceRange, error) {
	var ranges []*domain.PriceRange
	if err := r.db.WithContext(ctx).Find(&ranges).Error; err != nil {
		return nil, err
	}
	return ranges, nil
}

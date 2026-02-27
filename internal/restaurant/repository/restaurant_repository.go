package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgRestaurantRepository struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) domain.RestaurantRepository {
	return &pgRestaurantRepository{db: db}
}

func (r *pgRestaurantRepository) GetByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	var restaurant domain.Restaurant
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Preload("Province").
		Preload("District").
		Preload("Subdistrict").
		First(&restaurant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &restaurant, nil
}

func (r *pgRestaurantRepository) List(ctx context.Context, filter domain.RestaurantFilter) ([]*domain.Restaurant, int64, error) {
	var restaurants []*domain.Restaurant
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Restaurant{})

	if filter.ProvinceID != "" {
		query = query.Where("province_id = ?", filter.ProvinceID)
	}
	if filter.DistrictID != "" {
		query = query.Where("district_id = ?", filter.DistrictID)
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
		Find(&restaurants).Error; err != nil {
		return nil, 0, err
	}

	return restaurants, total, nil
}

func (r *pgRestaurantRepository) GetFoodTypes(ctx context.Context) ([]*domain.FoodType, error) {
	var foodTypes []*domain.FoodType
	if err := r.db.WithContext(ctx).Find(&foodTypes).Error; err != nil {
		return nil, err
	}
	return foodTypes, nil
}

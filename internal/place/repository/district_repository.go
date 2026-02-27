package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgDistrictRepository struct {
	db *gorm.DB
}

func NewDistrictRepository(db *gorm.DB) domain.DistrictRepository {
	return &pgDistrictRepository{db: db}
}

func (r *pgDistrictRepository) GetAll(ctx context.Context) ([]*domain.District, error) {
	var districts []*domain.District
	if err := r.db.WithContext(ctx).
		Preload("Province").
		Preload("Province.Region").
		Where("district_name_th != ?", "").
		Find(&districts).Error; err != nil {
		return nil, err
	}
	return districts, nil
}

func (r *pgDistrictRepository) GetByID(ctx context.Context, id string) (*domain.District, error) {
	var district domain.District
	if err := r.db.WithContext(ctx).
		Preload("Province").
		Preload("Province.Region").
		Where("district_name_th != ?", "").
		First(&district, "district_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &district, nil
}

func (r *pgDistrictRepository) GetByProvinceID(ctx context.Context, provinceID string) ([]*domain.District, error) {
	var districts []*domain.District
	if err := r.db.WithContext(ctx).
		Preload("Province").
		Preload("Province.Region").
		Where("province_id = ? AND district_name_th != ?", provinceID, "").
		Find(&districts).Error; err != nil {
		return nil, err
	}
	return districts, nil
}

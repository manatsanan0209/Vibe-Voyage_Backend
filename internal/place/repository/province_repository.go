package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgProvinceRepository struct {
	db *gorm.DB
}

func NewProvinceRepository(db *gorm.DB) domain.ProvinceRepository {
	return &pgProvinceRepository{db: db}
}

func (r *pgProvinceRepository) GetAll(ctx context.Context) ([]*domain.Province, error) {
	var provinces []*domain.Province
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Where("province_name_th != ?", "").
		Find(&provinces).Error; err != nil {
		return nil, err
	}
	return provinces, nil
}

func (r *pgProvinceRepository) GetByID(ctx context.Context, id string) (*domain.Province, error) {
	var province domain.Province
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Where("province_name_th != ?", "").
		First(&province, "province_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &province, nil
}

func (r *pgProvinceRepository) GetByRegionID(ctx context.Context, regionID string) ([]*domain.Province, error) {
	var provinces []*domain.Province
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Where("region_id = ? AND province_name_th != ?", regionID, "").
		Find(&provinces).Error; err != nil {
		return nil, err
	}
	return provinces, nil
}

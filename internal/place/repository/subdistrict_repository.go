package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgSubdistrictRepository struct {
	db *gorm.DB
}

func NewSubdistrictRepository(db *gorm.DB) domain.SubdistrictRepository {
	return &pgSubdistrictRepository{db: db}
}

func (r *pgSubdistrictRepository) GetAll(ctx context.Context) ([]*domain.Subdistrict, error) {
	var subdistricts []*domain.Subdistrict
	if err := r.db.WithContext(ctx).
		Preload("District").
		Preload("District.Province").
		Preload("District.Province.Region").
		Where("subdistrict_name_th != ?", "").
		Find(&subdistricts).Error; err != nil {
		return nil, err
	}
	return subdistricts, nil
}

func (r *pgSubdistrictRepository) GetByID(ctx context.Context, id string) (*domain.Subdistrict, error) {
	var subdistrict domain.Subdistrict
	if err := r.db.WithContext(ctx).
		Preload("District").
		Preload("District.Province").
		Preload("District.Province.Region").
		Where("subdistrict_name_th != ?", "").
		First(&subdistrict, "subdistrict_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subdistrict, nil
}

func (r *pgSubdistrictRepository) GetByDistrictID(ctx context.Context, districtID string) ([]*domain.Subdistrict, error) {
	var subdistricts []*domain.Subdistrict
	if err := r.db.WithContext(ctx).
		Preload("District").
		Preload("District.Province").
		Preload("District.Province.Region").
		Where("district_id = ? AND subdistrict_name_th != ?", districtID, "").
		Find(&subdistricts).Error; err != nil {
		return nil, err
	}
	return subdistricts, nil
}

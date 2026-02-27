package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgRegionRepository struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) domain.RegionRepository {
	return &pgRegionRepository{db: db}
}

func (r *pgRegionRepository) GetAll(ctx context.Context) ([]*domain.Region, error) {
	var regions []*domain.Region
	if err := r.db.WithContext(ctx).
		Where("region_name_th != ?", "").
		Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

func (r *pgRegionRepository) GetByID(ctx context.Context, id string) (*domain.Region, error) {
	var region domain.Region
	if err := r.db.WithContext(ctx).
		Where("region_name_th != ?", "").
		First(&region, "region_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &region, nil
}

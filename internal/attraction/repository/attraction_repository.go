package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgAttractionRepository struct {
	db *gorm.DB
}

func NewAttractionRepository(db *gorm.DB) domain.AttractionRepository {
	return &pgAttractionRepository{db: db}
}

func (r *pgAttractionRepository) GetByID(ctx context.Context, id string) (*domain.Attraction, error) {
	var attraction domain.Attraction
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Preload("Province").
		Preload("District").
		Preload("Subdistrict").
		Preload("Category").
		Preload("Type").
		First(&attraction, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attraction, nil
}

func (r *pgAttractionRepository) GetByName(ctx context.Context, name string) ([]*domain.Attraction, error) {
	var attractions []*domain.Attraction
	like := "%" + name + "%"
	if err := r.db.WithContext(ctx).
		Preload("Region").
		Preload("Province").
		Preload("District").
		Preload("Subdistrict").
		Preload("Category").
		Preload("Type").
		Where("name_th ILIKE ? OR name_en ILIKE ?", like, like).
		Limit(10).
		Find(&attractions).Error; err != nil {
		return nil, err
	}
	return attractions, nil
}

func (r *pgAttractionRepository) List(ctx context.Context, filter domain.AttractionFilter) ([]*domain.Attraction, int64, error) {
	var attractions []*domain.Attraction
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Attraction{})

	if filter.ProvinceID != "" {
		query = query.Where("province_id = ?", filter.ProvinceID)
	}
	if filter.DistrictID != "" {
		query = query.Where("district_id = ?", filter.DistrictID)
	}
	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.TypeID != "" {
		query = query.Where("type_id = ?", filter.TypeID)
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
		Preload("Category").
		Preload("Type").
		Find(&attractions).Error; err != nil {
		return nil, 0, err
	}

	return attractions, total, nil
}

func (r *pgAttractionRepository) GetCategories(ctx context.Context) ([]*domain.AttractionCategory, error) {
	var categories []*domain.AttractionCategory
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *pgAttractionRepository) GetTypes(ctx context.Context) ([]*domain.AttractionType, error) {
	var types []*domain.AttractionType
	if err := r.db.WithContext(ctx).Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

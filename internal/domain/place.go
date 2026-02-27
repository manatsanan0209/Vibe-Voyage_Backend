package domain

import "context"

// ========================
// Entities
// ========================

type Region struct {
	RegionID     string `json:"region_id" gorm:"primaryKey;column:region_id"`
	RegionNameTH string `json:"region_name_th" gorm:"column:region_name_th"`
}

type Province struct {
	ProvinceID     string  `json:"province_id" gorm:"primaryKey;column:province_id"`
	ProvinceNameTH string  `json:"province_name_th" gorm:"column:province_name_th"`
	RegionID       string  `json:"region_id" gorm:"column:region_id"`
	Region         *Region `json:"region,omitempty" gorm:"foreignKey:RegionID;references:RegionID"`
}

type District struct {
	DistrictID     string    `json:"district_id" gorm:"primaryKey;column:district_id"`
	DistrictNameTH string    `json:"district_name_th" gorm:"column:district_name_th"`
	ProvinceID     string    `json:"province_id" gorm:"column:province_id"`
	Province       *Province `json:"province,omitempty" gorm:"foreignKey:ProvinceID;references:ProvinceID"`
}

type Subdistrict struct {
	SubdistrictID     string    `json:"subdistrict_id" gorm:"primaryKey;column:subdistrict_id"`
	SubdistrictNameTH string    `json:"subdistrict_name_th" gorm:"column:subdistrict_name_th"`
	DistrictID        string    `json:"district_id" gorm:"column:district_id"`
	District          *District `json:"district,omitempty" gorm:"foreignKey:DistrictID;references:DistrictID"`
}

// ========================
// Repositories
// ========================

type RegionRepository interface {
	GetAll(ctx context.Context) ([]*Region, error)
	GetByID(ctx context.Context, id string) (*Region, error)
}

type ProvinceRepository interface {
	GetAll(ctx context.Context) ([]*Province, error)
	GetByID(ctx context.Context, id string) (*Province, error)
	GetByRegionID(ctx context.Context, regionID string) ([]*Province, error)
}

type DistrictRepository interface {
	GetAll(ctx context.Context) ([]*District, error)
	GetByID(ctx context.Context, id string) (*District, error)
	GetByProvinceID(ctx context.Context, provinceID string) ([]*District, error)
}

type SubdistrictRepository interface {
	GetAll(ctx context.Context) ([]*Subdistrict, error)
	GetByID(ctx context.Context, id string) (*Subdistrict, error)
	GetByDistrictID(ctx context.Context, districtID string) ([]*Subdistrict, error)
}

// ========================
// Service
// ========================

type PlaceService interface {
	GetAllRegions(ctx context.Context) ([]*Region, error)
	GetRegionByID(ctx context.Context, id string) (*Region, error)

	GetAllProvinces(ctx context.Context) ([]*Province, error)
	GetProvinceByID(ctx context.Context, id string) (*Province, error)
	GetProvincesByRegion(ctx context.Context, regionID string) ([]*Province, error)

	GetAllDistricts(ctx context.Context) ([]*District, error)
	GetDistrictByID(ctx context.Context, id string) (*District, error)
	GetDistrictsByProvince(ctx context.Context, provinceID string) ([]*District, error)

	GetAllSubdistricts(ctx context.Context) ([]*Subdistrict, error)
	GetSubdistrictByID(ctx context.Context, id string) (*Subdistrict, error)
	GetSubdistrictsByDistrict(ctx context.Context, districtID string) ([]*Subdistrict, error)
}

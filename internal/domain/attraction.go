package domain

import "context"

// ========================
// Entities
// ========================

type AttractionCategory struct {
	CategoryID   string `json:"category_id" gorm:"primaryKey;column:category_id"`
	CategoryName string `json:"category_name" gorm:"column:category_name"`
}

type AttractionType struct {
	TypeID   string `json:"type_id" gorm:"primaryKey;column:type_id"`
	TypeName string `json:"type_name" gorm:"column:type_name"`
}

type Attraction struct {
	ID            string              `json:"id" gorm:"primaryKey;column:id"`
	NameTH        string              `json:"name_th" gorm:"column:name_th"`
	NameEN        string              `json:"name_en" gorm:"column:name_en"`
	Latitude      float64             `json:"latitude" gorm:"column:latitude"`
	Longitude     float64             `json:"longitude" gorm:"column:longitude"`
	Address       string              `json:"address" gorm:"column:address"`
	Postcode      string              `json:"postcode" gorm:"column:postcode"`
	Phone         string              `json:"phone" gorm:"column:phone"`
	Email         string              `json:"email" gorm:"column:email"`
	Website       string              `json:"website" gorm:"column:website"`
	Facebook      string              `json:"facebook" gorm:"column:facebook"`
	Instagram     string              `json:"instagram" gorm:"column:instagram"`
	Line          string              `json:"line" gorm:"column:line"`
	TikTok        string              `json:"tiktok" gorm:"column:tiktok"`
	YouTube       string              `json:"youtube" gorm:"column:youtube"`
	RegionID      string              `json:"region_id" gorm:"column:region_id"`
	Region        *Region             `json:"region,omitempty" gorm:"foreignKey:RegionID;references:RegionID"`
	ProvinceID    string              `json:"province_id" gorm:"column:province_id"`
	Province      *Province           `json:"province,omitempty" gorm:"foreignKey:ProvinceID;references:ProvinceID"`
	DistrictID    string              `json:"district_id" gorm:"column:district_id"`
	District      *District           `json:"district,omitempty" gorm:"foreignKey:DistrictID;references:DistrictID"`
	SubdistrictID string              `json:"subdistrict_id" gorm:"column:subdistrict_id"`
	Subdistrict   *Subdistrict        `json:"subdistrict,omitempty" gorm:"foreignKey:SubdistrictID;references:SubdistrictID"`
	CategoryID    string              `json:"category_id" gorm:"column:category_id"`
	Category      *AttractionCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:CategoryID"`
	TypeID        string              `json:"type_id" gorm:"column:type_id"`
	Type          *AttractionType     `json:"type,omitempty" gorm:"foreignKey:TypeID;references:TypeID"`
	DetailTH      string              `json:"detail_th" gorm:"column:detail_th"`
	DetailEN      string              `json:"detail_en" gorm:"column:detail_en"`
	Highlight     string              `json:"highlight" gorm:"column:highlight"`
	Remark        string              `json:"remark" gorm:"column:remark"`
	Reward        string              `json:"reward" gorm:"column:reward"`
	Status        string              `json:"status" gorm:"column:status"`
	StatusData    string              `json:"status_data" gorm:"column:status_data"`
	CreatedDate   string              `json:"created_date" gorm:"column:created_date"`
	UpdatedDate   string              `json:"updated_date" gorm:"column:updated_date"`
}

// ========================
// Filter
// ========================

type AttractionFilter struct {
	ProvinceID string
	DistrictID string
	CategoryID string
	TypeID     string
	Search     string
	Limit      int
	Offset     int
}

// ========================
// Repository
// ========================

type AttractionRepository interface {
	GetByID(ctx context.Context, id string) (*Attraction, error)
	GetByName(ctx context.Context, name string) ([]*Attraction, error)
	List(ctx context.Context, filter AttractionFilter) ([]*Attraction, int64, error)
	GetCategories(ctx context.Context) ([]*AttractionCategory, error)
	GetTypes(ctx context.Context) ([]*AttractionType, error)
}

// ========================
// Service
// ========================

type AttractionService interface {
	GetAttractionByID(ctx context.Context, id string) (*Attraction, error)
	GetAttractionByName(ctx context.Context, name string) ([]*Attraction, error)
	ListAttractions(ctx context.Context, filter AttractionFilter) ([]*Attraction, int64, error)
	GetAttractionCategories(ctx context.Context) ([]*AttractionCategory, error)
	GetAttractionTypes(ctx context.Context) ([]*AttractionType, error)
}

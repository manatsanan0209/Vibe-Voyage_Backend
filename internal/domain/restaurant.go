package domain

import "context"

// ========================
// Entities
// ========================

type FoodType struct {
	FoodTypeID   int    `json:"food_type_id" gorm:"primaryKey;column:food_type_id"`
	FoodTypeName string `json:"food_type_name" gorm:"column:food_type_name"`
}

type Restaurant struct {
	ID                  string       `json:"id" gorm:"primaryKey;column:id"`
	NameTH              string       `json:"name_th" gorm:"column:name_th"`
	NameEN              string       `json:"name_en" gorm:"column:name_en"`
	Latitude            float64      `json:"latitude" gorm:"column:latitude"`
	Longitude           float64      `json:"longitude" gorm:"column:longitude"`
	Address             string       `json:"address" gorm:"column:address"`
	Alley               string       `json:"alley" gorm:"column:alley"`
	Postcode            string       `json:"postcode" gorm:"column:postcode"`
	Phone               string       `json:"phone" gorm:"column:phone"`
	Email               string       `json:"email" gorm:"column:email"`
	Website             string       `json:"website" gorm:"column:website"`
	Facebook            string       `json:"facebook" gorm:"column:facebook"`
	Instagram           string       `json:"instagram" gorm:"column:instagram"`
	Line                string       `json:"line" gorm:"column:line"`
	TikTok              string       `json:"tiktok" gorm:"column:tiktok"`
	YouTube             string       `json:"youtube" gorm:"column:youtube"`
	RegionID            string       `json:"region_id" gorm:"column:region_id"`
	Region              *Region      `json:"region,omitempty" gorm:"foreignKey:RegionID;references:RegionID"`
	ProvinceID          string       `json:"province_id" gorm:"column:province_id"`
	Province            *Province    `json:"province,omitempty" gorm:"foreignKey:ProvinceID;references:ProvinceID"`
	DistrictID          string       `json:"district_id" gorm:"column:district_id"`
	District            *District    `json:"district,omitempty" gorm:"foreignKey:DistrictID;references:DistrictID"`
	SubdistrictID       string       `json:"subdistrict_id" gorm:"column:subdistrict_id"`
	Subdistrict         *Subdistrict `json:"subdistrict,omitempty" gorm:"foreignKey:SubdistrictID;references:SubdistrictID"`
	NearbyLocation      string       `json:"nearby_location" gorm:"column:nearby_location"`
	Contact             string       `json:"contact" gorm:"column:contact"`
	PlaceType           string       `json:"place_type" gorm:"column:place_type"`
	Detail              string       `json:"detail" gorm:"column:detail"`
	Highlight           string       `json:"highlight" gorm:"column:highlight"`
	RecommendFood       string       `json:"recommend_food" gorm:"column:recommend_food"`
	AccessibilityDetail string       `json:"accessibility_detail" gorm:"column:accessibility_detail"`
	Remark              string       `json:"remark" gorm:"column:remark"`
	AcceptCredit        string       `json:"accept_credit" gorm:"column:accept_credit"`
	AcceptCash          string       `json:"accept_cash" gorm:"column:accept_cash"`
	AcceptPayment       string       `json:"accept_payment" gorm:"column:accept_payment"`
	FoodTypes           string       `json:"food_types" gorm:"column:food_types"`
}

// ========================
// Filter
// ========================

type RestaurantFilter struct {
	ProvinceID string
	DistrictID string
	Search     string
	Limit      int
	Offset     int
}

// ========================
// Repository
// ========================

type RestaurantRepository interface {
	GetByID(ctx context.Context, id string) (*Restaurant, error)
	GetByName(ctx context.Context, name string) ([]*Restaurant, error)
	List(ctx context.Context, filter RestaurantFilter) ([]*Restaurant, int64, error)
	GetFoodTypes(ctx context.Context) ([]*FoodType, error)
}

// ========================
// Service
// ========================

type RestaurantService interface {
	GetRestaurantByID(ctx context.Context, id string) (*Restaurant, error)
	GetRestaurantByName(ctx context.Context, name string) ([]*Restaurant, error)
	ListRestaurants(ctx context.Context, filter RestaurantFilter) ([]*Restaurant, int64, error)
	GetFoodTypes(ctx context.Context) ([]*FoodType, error)
}

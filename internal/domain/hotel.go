package domain

import "context"

// ========================
// Entities
// ========================

type AccommodationType struct {
	AccomTypeID   string `json:"accom_type_id" gorm:"primaryKey;column:accom_type_id"`
	AccomTypeName string `json:"accom_type_name" gorm:"column:accom_type_name"`
}

type PriceRange struct {
	PriceRangeID   string `json:"price_range_id" gorm:"primaryKey;column:price_range_id"`
	PriceRangeName string `json:"price_range_name" gorm:"column:price_range_name"`
}

type Hotel struct {
	ID                  string             `json:"id" gorm:"primaryKey;column:id"`
	NameTH              string             `json:"name_th" gorm:"column:name_th"`
	NameEN              string             `json:"name_en" gorm:"column:name_en"`
	Latitude            float64            `json:"latitude" gorm:"column:latitude"`
	Longitude           float64            `json:"longitude" gorm:"column:longitude"`
	Address             string             `json:"address" gorm:"column:address"`
	Alley               string             `json:"alley" gorm:"column:alley"`
	Road                string             `json:"road" gorm:"column:road"`
	Postcode            string             `json:"postcode" gorm:"column:postcode"`
	Phone               string             `json:"phone" gorm:"column:phone"`
	Email               string             `json:"email" gorm:"column:email"`
	Website             string             `json:"website" gorm:"column:website"`
	Facebook            string             `json:"facebook" gorm:"column:facebook"`
	Instagram           string             `json:"instagram" gorm:"column:instagram"`
	Line                string             `json:"line" gorm:"column:line"`
	TikTok              string             `json:"tiktok" gorm:"column:tiktok"`
	YouTube             string             `json:"youtube" gorm:"column:youtube"`
	RegionID            string             `json:"region_id" gorm:"column:region_id"`
	Region              *Region            `json:"region,omitempty" gorm:"foreignKey:RegionID;references:RegionID"`
	ProvinceID          string             `json:"province_id" gorm:"column:province_id"`
	Province            *Province          `json:"province,omitempty" gorm:"foreignKey:ProvinceID;references:ProvinceID"`
	DistrictID          string             `json:"district_id" gorm:"column:district_id"`
	District            *District          `json:"district,omitempty" gorm:"foreignKey:DistrictID;references:DistrictID"`
	SubdistrictID       string             `json:"subdistrict_id" gorm:"column:subdistrict_id"`
	Subdistrict         *Subdistrict       `json:"subdistrict,omitempty" gorm:"foreignKey:SubdistrictID;references:SubdistrictID"`
	AccomTypeID         string             `json:"accom_type_id" gorm:"column:accom_type_id"`
	AccommodationType   *AccommodationType `json:"accommodation_type,omitempty" gorm:"foreignKey:AccomTypeID;references:AccomTypeID"`
	PriceRangeID        string             `json:"price_range_id" gorm:"column:price_range_id"`
	PriceRange          *PriceRange        `json:"price_range,omitempty" gorm:"foreignKey:PriceRangeID;references:PriceRangeID"`
	RoomCount           float64            `json:"room_count" gorm:"column:room_count"`
	HighRate            float64            `json:"high_rate" gorm:"column:high_rate"`
	LowRate             float64            `json:"low_rate" gorm:"column:low_rate"`
	MeetingRoom         float64            `json:"meeting_room" gorm:"column:meeting_room"`
	MeetingValue        string             `json:"meeting_value" gorm:"column:meeting_value"`
	StarStandard        float64            `json:"star_standard" gorm:"column:star_standard"`
	GreenStandard       string             `json:"green_standard" gorm:"column:green_standard"`
	OpeningYear         float64            `json:"opening_year" gorm:"column:opening_year"`
	CheckInTime         string             `json:"check_in_time" gorm:"column:check_in_time"`
	CheckOutTime        string             `json:"check_out_time" gorm:"column:check_out_time"`
	Highlight           string             `json:"highlight" gorm:"column:highlight"`
	Reward              string             `json:"reward" gorm:"column:reward"`
	AccessibilityDetail string             `json:"accessibility_detail" gorm:"column:accessibility_detail"`
	Remark              string             `json:"remark" gorm:"column:remark"`
	Status              string             `json:"status" gorm:"column:status"`
	TravelStatus        string             `json:"travel_status" gorm:"column:travel_status"`
	RegisStatus         string             `json:"regis_status" gorm:"column:regis_status"`
	MemberStatus        string             `json:"member_status" gorm:"column:member_status"`
	StatusData          string             `json:"status_data" gorm:"column:status_data"`
	CreatedDate         string             `json:"created_date" gorm:"column:created_date"`
	UpdatedDate         string             `json:"updated_date" gorm:"column:updated_date"`
}

// ========================
// Filter
// ========================

type HotelFilter struct {
	ProvinceID   string
	DistrictID   string
	AccomTypeID  string
	PriceRangeID string
	Search       string
	Limit        int
	Offset       int
}

// ========================
// Repository
// ========================

type HotelRepository interface {
	GetByID(ctx context.Context, id string) (*Hotel, error)
	GetByName(ctx context.Context, name string) ([]*Hotel, error)
	List(ctx context.Context, filter HotelFilter) ([]*Hotel, int64, error)
	GetAccommodationTypes(ctx context.Context) ([]*AccommodationType, error)
	GetPriceRanges(ctx context.Context) ([]*PriceRange, error)
}

// ========================
// Service
// ========================

type HotelService interface {
	GetHotelByID(ctx context.Context, id string) (*Hotel, error)
	GetHotelByName(ctx context.Context, name string) ([]*Hotel, error)
	ListHotels(ctx context.Context, filter HotelFilter) ([]*Hotel, int64, error)
	GetAccommodationTypes(ctx context.Context) ([]*AccommodationType, error)
	GetPriceRanges(ctx context.Context) ([]*PriceRange, error)
}

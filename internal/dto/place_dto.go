package dto

import "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"

type RegionResponseDTO struct {
	RegionID     string `json:"region_id"`
	RegionNameTH string `json:"region_name_th"`
}

type ProvinceResponseDTO struct {
	ProvinceID     string             `json:"province_id"`
	ProvinceNameTH string             `json:"province_name_th"`
	RegionID       string             `json:"region_id"`
	Region         *RegionResponseDTO `json:"region,omitempty"`
}

type DistrictResponseDTO struct {
	DistrictID     string               `json:"district_id"`
	DistrictNameTH string               `json:"district_name_th"`
	ProvinceID     string               `json:"province_id"`
	Province       *ProvinceResponseDTO `json:"province,omitempty"`
}

type SubdistrictResponseDTO struct {
	SubdistrictID     string               `json:"subdistrict_id"`
	SubdistrictNameTH string               `json:"subdistrict_name_th"`
	DistrictID        string               `json:"district_id"`
	District          *DistrictResponseDTO `json:"district,omitempty"`
}

func NewRegionResponseDTO(region *domain.Region) *RegionResponseDTO {
	if region == nil {
		return nil
	}

	return &RegionResponseDTO{
		RegionID:     region.RegionID,
		RegionNameTH: region.RegionNameTH,
	}
}

func NewProvinceResponseDTO(province *domain.Province) *ProvinceResponseDTO {
	if province == nil {
		return nil
	}

	return &ProvinceResponseDTO{
		ProvinceID:     province.ProvinceID,
		ProvinceNameTH: province.ProvinceNameTH,
		RegionID:       province.RegionID,
		Region:         NewRegionResponseDTO(province.Region),
	}
}

func NewDistrictResponseDTO(district *domain.District) *DistrictResponseDTO {
	if district == nil {
		return nil
	}

	return &DistrictResponseDTO{
		DistrictID:     district.DistrictID,
		DistrictNameTH: district.DistrictNameTH,
		ProvinceID:     district.ProvinceID,
		Province:       NewProvinceResponseDTO(district.Province),
	}
}

func NewSubdistrictResponseDTO(subdistrict *domain.Subdistrict) *SubdistrictResponseDTO {
	if subdistrict == nil {
		return nil
	}

	return &SubdistrictResponseDTO{
		SubdistrictID:     subdistrict.SubdistrictID,
		SubdistrictNameTH: subdistrict.SubdistrictNameTH,
		DistrictID:        subdistrict.DistrictID,
		District:          NewDistrictResponseDTO(subdistrict.District),
	}
}

func NewRegionResponseDTOList(regions []*domain.Region) []RegionResponseDTO {
	result := make([]RegionResponseDTO, 0, len(regions))
	for _, region := range regions {
		if converted := NewRegionResponseDTO(region); converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func NewProvinceResponseDTOList(provinces []*domain.Province) []ProvinceResponseDTO {
	result := make([]ProvinceResponseDTO, 0, len(provinces))
	for _, province := range provinces {
		if converted := NewProvinceResponseDTO(province); converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func NewDistrictResponseDTOList(districts []*domain.District) []DistrictResponseDTO {
	result := make([]DistrictResponseDTO, 0, len(districts))
	for _, district := range districts {
		if converted := NewDistrictResponseDTO(district); converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func NewSubdistrictResponseDTOList(subdistricts []*domain.Subdistrict) []SubdistrictResponseDTO {
	result := make([]SubdistrictResponseDTO, 0, len(subdistricts))
	for _, subdistrict := range subdistricts {
		if converted := NewSubdistrictResponseDTO(subdistrict); converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

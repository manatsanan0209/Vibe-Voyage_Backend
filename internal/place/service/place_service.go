package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type placeService struct {
	regionRepo      domain.RegionRepository
	provinceRepo    domain.ProvinceRepository
	districtRepo    domain.DistrictRepository
	subdistrictRepo domain.SubdistrictRepository
}

func NewPlaceService(
	regionRepo domain.RegionRepository,
	provinceRepo domain.ProvinceRepository,
	districtRepo domain.DistrictRepository,
	subdistrictRepo domain.SubdistrictRepository,
) domain.PlaceService {
	return &placeService{
		regionRepo:      regionRepo,
		provinceRepo:    provinceRepo,
		districtRepo:    districtRepo,
		subdistrictRepo: subdistrictRepo,
	}
}

func (s *placeService) GetAllRegions(ctx context.Context) ([]*domain.Region, error) {
	return s.regionRepo.GetAll(ctx)
}

func (s *placeService) GetRegionByID(ctx context.Context, id string) (*domain.Region, error) {
	return s.regionRepo.GetByID(ctx, id)
}

func (s *placeService) GetAllProvinces(ctx context.Context) ([]*domain.Province, error) {
	return s.provinceRepo.GetAll(ctx)
}

func (s *placeService) GetProvinceByID(ctx context.Context, id string) (*domain.Province, error) {
	return s.provinceRepo.GetByID(ctx, id)
}

func (s *placeService) GetProvincesByRegion(ctx context.Context, regionID string) ([]*domain.Province, error) {
	return s.provinceRepo.GetByRegionID(ctx, regionID)
}

func (s *placeService) GetAllDistricts(ctx context.Context) ([]*domain.District, error) {
	return s.districtRepo.GetAll(ctx)
}

func (s *placeService) GetDistrictByID(ctx context.Context, id string) (*domain.District, error) {
	return s.districtRepo.GetByID(ctx, id)
}

func (s *placeService) GetDistrictsByProvince(ctx context.Context, provinceID string) ([]*domain.District, error) {
	return s.districtRepo.GetByProvinceID(ctx, provinceID)
}

func (s *placeService) GetAllSubdistricts(ctx context.Context) ([]*domain.Subdistrict, error) {
	return s.subdistrictRepo.GetAll(ctx)
}

func (s *placeService) GetSubdistrictByID(ctx context.Context, id string) (*domain.Subdistrict, error) {
	return s.subdistrictRepo.GetByID(ctx, id)
}

func (s *placeService) GetSubdistrictsByDistrict(ctx context.Context, districtID string) ([]*domain.Subdistrict, error) {
	return s.subdistrictRepo.GetByDistrictID(ctx, districtID)
}

package place

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place/handler"
	placeRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place/repository"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	regionRepo := placeRepo.NewRegionRepository(db)
	provinceRepo := placeRepo.NewProvinceRepository(db)
	districtRepo := placeRepo.NewDistrictRepository(db)
	subdistrictRepo := placeRepo.NewSubdistrictRepository(db)

	svc := service.NewPlaceService(regionRepo, provinceRepo, districtRepo, subdistrictRepo)

	handler.NewRegionHandler(svc).RegisterRoutes(app)
	handler.NewProvinceHandler(svc).RegisterRoutes(app)
	handler.NewDistrictHandler(svc).RegisterRoutes(app)
	handler.NewSubdistrictHandler(svc).RegisterRoutes(app)
}

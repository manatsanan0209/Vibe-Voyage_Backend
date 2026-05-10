package server

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/db"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	tripRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip/repository"
	tripService "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip/service"
	tripSuggestionPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip_suggestion"
	tripSuggestionRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip_suggestion/repository"
	tripSuggestionService "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip_suggestion/service"

	userRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/repository"
	userService "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/service"
	userLifestyleClient "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/client"
	userLifestyleRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/repository"
	userLifestyleService "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/service"

	attractionPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/attraction"
	attractionRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/attraction/repository"
	authPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth"
	healthPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/health"
	hotelPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/hotel"
	notificationPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/notification"
	notificationRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/notification/repository"
	placePkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place"
	placeDetailClient "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place_detail/client"
	placeDetailRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place_detail/repository"
	placeDetailService "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place_detail/service"
	restaurantPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/restaurant"
	restaurantRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/restaurant/repository"
	roomPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room"
	roomRepoPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/repository"
	roomServicePkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/service"
	settingsPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/settings"
	settingsRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/settings/repository"
	tripPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip"
	userPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user"
	userLifestylePkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle"
)

func Run() error {
	_ = godotenv.Load()

	gormDB, err := db.Connect()
	if err != nil {
		return err
	}

	err = gormDB.AutoMigrate(
		&domain.User{},
		&domain.Room{},
		&domain.Trips{},
		&domain.RoomMember{},
		&domain.RoomInviteCode{},
		&domain.UserLifestyle{},
		&domain.TripSchedule{},
		&domain.UserSettings{},
		&domain.Notification{},
		&domain.PublishedTrip{},
		&domain.TripLike{},
		&domain.TripBookmark{},
		&domain.GooglePlaceDetail{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Database Migration Completed!")

	app := fiber.New()
	app.Use(cors.New(corsConfigFromEnv()))
	app.Use(logger.New(logger.Config{
		TimeFormat: "2006-01-02 15:04:05",
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
	}))

	repo := userRepo.NewUserRepository(gormDB)
	svc := userService.NewUserService(repo)
	tripRepository := tripRepo.NewTripRepository(gormDB)
	restaurantRepository := restaurantRepo.NewRestaurantRepository(gormDB)
	attractionRepository := attractionRepo.NewAttractionRepository(gormDB)
	roomRepository := roomRepoPkg.NewRoomRepository(gormDB)
	roomInviteRepository := roomRepoPkg.NewRoomInviteCodeRepository(gormDB)
	lifestyleRepository := userLifestyleRepo.NewUserLifestyleRepository(gormDB)
	recommendationClient := userLifestyleClient.NewHTTPRecommendationClient()
	lifestyleSvc := userLifestyleService.NewUserLifestyleService(lifestyleRepository, recommendationClient)

	settingsRepository := settingsRepo.NewUserSettingsRepository(gormDB)
	notifRepository := notificationRepo.NewNotificationRepository(gormDB)
	notifSvc := notificationPkg.SetupService(notifRepository, settingsRepository)
	googlePlaceDetailRepository := placeDetailRepo.NewGooglePlaceDetailRepository(gormDB)
	googlePlacesClient := placeDetailClient.NewGooglePlacesClientFromEnv()
	googlePlaceDetailSvc := placeDetailService.NewGooglePlaceDetailService(googlePlaceDetailRepository, googlePlacesClient)
	tripSuggestionRepository := tripSuggestionRepo.NewTripSuggestionRepository(gormDB)
	tripSuggestionSvc := tripSuggestionService.NewTripSuggestionService(tripSuggestionRepository)

	roomSvc := roomServicePkg.NewRoomService(roomRepository, roomInviteRepository, lifestyleRepository, lifestyleSvc, notifSvc)
	tripSvc := tripService.NewTripService(tripRepository, restaurantRepository, attractionRepository, lifestyleSvc, roomSvc, roomRepository, tripSuggestionSvc, notifSvc, googlePlaceDetailSvc)

	healthPkg.RegisterRoutes(app)
	userPkg.Setup(app, svc, tripSuggestionSvc)
	authPkg.Setup(app, repo)

	placePkg.Setup(app, gormDB)
	attractionPkg.Setup(app, gormDB)
	hotelPkg.Setup(app, gormDB)
	restaurantPkg.Setup(app, gormDB)
	tripPkg.Setup(app, tripSvc, tripSuggestionSvc)
	userLifestylePkg.Setup(app, lifestyleSvc)
	roomPkg.Setup(app, roomSvc)
	settingsPkg.SetupWithRepo(app, settingsRepository)
	notificationPkg.SetupHandler(app, notifSvc)
	tripSuggestionPkg.Setup(app, tripSuggestionSvc)

	port, err := portFromEnv()
	if err != nil {
		return err
	}

	return app.Listen(":" + port)
}

func portFromEnv() (string, error) {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		return "", fmt.Errorf("PORT environment variable is required")
	}

	return port, nil
}

func corsConfigFromEnv() cors.Config {
	return cors.Config{
		AllowOrigins: envOrDefault("CORS_ALLOW_ORIGINS", "*"),
		AllowMethods: envOrDefault("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
		AllowHeaders: envOrDefault("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

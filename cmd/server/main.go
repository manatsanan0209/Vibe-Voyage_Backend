package server

import (
	"log"

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
	authPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth"
	healthPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/health"
	hotelPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/hotel"
	notificationPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/notification"
	notificationRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/notification/repository"
	placePkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/place"
	settingsPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/settings"
	settingsRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/settings/repository"
	attractionRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/attraction/repository"
	restaurantPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/restaurant"
	restaurantRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/restaurant/repository"
	roomPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room"
	roomRepoPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/repository"
	roomServicePkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/service"
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
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Database Migration Completed!")

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
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

	roomSvc := roomServicePkg.NewRoomService(roomRepository, roomInviteRepository, lifestyleRepository, lifestyleSvc, notifSvc)
	tripSvc := tripService.NewTripService(tripRepository, restaurantRepository, attractionRepository, lifestyleSvc, roomSvc, notifSvc)
	tripSuggestionRepository := tripSuggestionRepo.NewTripSuggestionRepository(gormDB)
	tripSuggestionSvc := tripSuggestionService.NewTripSuggestionService(tripSuggestionRepository)

	healthPkg.RegisterRoutes(app)
	userPkg.Setup(app, svc)
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

	return app.Listen(":8080")
}

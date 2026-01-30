package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chivta/spotiarch/internal/config"
	"github.com/chivta/spotiarch/internal/handlers"
	"github.com/chivta/spotiarch/internal/logger"
	"github.com/chivta/spotiarch/internal/middlewares"
	"github.com/chivta/spotiarch/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(
		cfg.Log.Debug,
		cfg.Log.EnableInfo,
		cfg.Log.ErrorOutput,
		cfg.Log.InfoOutput,
		cfg.Log.DebugOutput,
	)
	appLogger.Infof("Starting spotiarch application")

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		appLogger.Errorf("Failed to connect database: %v", err)
		log.Fatal("Failed to connect database")
	}
	appLogger.Infof("Database connection established")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// health check endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	app.Use(fiberLogger.New())
	app.Use(recover.New())

	authService := services.NewAuthService(db, appLogger, cfg.Token.SecretKey)
	authHandler := handlers.NewAuthHandler(authService)
	authMiddleware := middlewares.NewAuthMiddleware(authService, appLogger)

	userService := services.NewUserService(db, appLogger)
	userHandler := handlers.NewUserHandler(userService)


	api := app.Group("/api")
	{
		auth := api.Group("/auth")
		auth.Post("/signup", authHandler.SignUp)
		auth.Post("/login", authHandler.Login)
		auth.Post("/logout", authHandler.Logout)
		
		protected := api.Group("/")
		protected.Use(authMiddleware.New())

		user := protected.Group("/user")
		{
			user.Get("/me", userHandler.Me)
		}
	}

	appLogger.Infof("Server starting on port 3000")
	app.Listen(":3000")
}

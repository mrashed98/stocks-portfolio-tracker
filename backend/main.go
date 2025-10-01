package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"portfolio-app/config"
	"portfolio-app/internal/database"
	"portfolio-app/internal/handlers"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/routes"
	"portfolio-app/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Market data caching will be disabled")
	} else {
		log.Println("Redis connection established successfully")
	}

	// Run database migrations
	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed development data if in development mode
	if cfg.Server.Env == "development" {
		seeder := database.NewSeeder(db)
		if err := seeder.SeedDevelopmentData(); err != nil {
			log.Printf("Warning: Failed to seed development data: %v", err)
		}
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173", // Include Vite dev server
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowCredentials: true,
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database health
		if err := db.Health(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
				"message": "Database connection failed",
				"error": err.Error(),
			})
		}

		// Check Redis health
		redisStatus := "connected"
		if err := redisClient.Ping(c.Context()).Err(); err != nil {
			redisStatus = "disconnected"
		}

		return c.JSON(fiber.Map{
			"status": "ok",
			"message": "Portfolio API is running",
			"database": "connected",
			"redis": redisStatus,
		})
	})

	// Initialize repositories
	strategyRepo := repositories.NewStrategyRepository(db.DB)
	stockRepo := repositories.NewStockRepository(db.DB)
	signalRepo := repositories.NewSignalRepository(db.DB)
	portfolioRepo := repositories.NewPortfolioRepository(db.DB)
	userRepo := repositories.NewUserRepository(db.DB)

	// Initialize services
	authService := services.NewAuthService(userRepo, redisClient, cfg.JWT.Secret)
	strategyService := services.NewStrategyService(strategyRepo, db.DB)
	stockService := services.NewStockService(stockRepo, signalRepo, strategyRepo, db.DB)
	
	// Initialize market data service
	marketDataServiceFactory := services.NewMarketDataServiceFactory(redisClient)
	marketDataProvider := os.Getenv("MARKET_DATA_PROVIDER")
	if marketDataProvider == "" {
		marketDataProvider = "mock" // Default to mock for development
	}
	marketDataService := marketDataServiceFactory.CreateService(marketDataProvider, cfg.Market.APIKey)

	// Initialize allocation engine
	allocationEngine := services.NewAllocationEngine(strategyRepo, stockRepo, signalRepo, marketDataService)
	
	// Initialize portfolio service
	portfolioService := services.NewPortfolioService(allocationEngine, strategyRepo, portfolioRepo, marketDataService)
	
	// Initialize NAV scheduler
	navScheduler := services.NewNAVScheduler(portfolioService, portfolioRepo, nil) // Use default config
	
	// Start NAV scheduler in development mode
	if cfg.Server.Env == "development" {
		if err := navScheduler.Start(); err != nil {
			log.Printf("Warning: Failed to start NAV scheduler: %v", err)
		} else {
			log.Println("NAV scheduler started successfully")
		}
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	strategyHandler := handlers.NewStrategyHandler(strategyService)
	stockHandler := handlers.NewStockHandler(stockService)
	marketDataHandler := handlers.NewMarketDataHandler(marketDataService)
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService)
	navSchedulerHandler := handlers.NewNAVSchedulerHandler(navScheduler)

	// API routes
	api := app.Group("/api/v1")
	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Portfolio API v1",
		})
	})

	// Setup routes
	routes.SetupAuthRoutes(api, authHandler, authService, userRepo)
	routes.SetupStrategyRoutes(api, strategyHandler, authService, userRepo)
	routes.SetupStockRoutes(api, stockHandler, authService, userRepo)
	routes.SetupMarketDataRoutes(api, marketDataHandler, authService, userRepo)
	routes.SetupPortfolioRoutes(api, portfolioHandler, authService, userRepo)
	routes.SetupNAVSchedulerRoutes(api, navSchedulerHandler, authService, userRepo)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
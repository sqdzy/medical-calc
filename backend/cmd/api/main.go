package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/medical-app/backend/config"
	"github.com/medical-app/backend/internal/app"
	"github.com/medical-app/backend/internal/handler/middleware"
	v1 "github.com/medical-app/backend/internal/handler/v1"
	"github.com/medical-app/backend/internal/infrastructure/postgres"
	"github.com/medical-app/backend/internal/repository"
	"github.com/medical-app/backend/internal/service"
	applogger "github.com/medical-app/backend/pkg/logger"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	zapLogger, err := applogger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Initialize database connection
	db, err := postgres.NewConnection(context.Background(), cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Run migrations
	if err := app.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		zapLogger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	services := service.NewServices(service.Deps{
		Repos:            repos,
		JWTSecret:        cfg.JWTSecret,
		JWTAccessExpiry:  cfg.JWTAccessExpiry,
		JWTRefreshExpiry: cfg.JWTRefreshExpiry,
		EncryptionKey:    cfg.EncryptionKey,
		NCBIApiKey:       cfg.NCBIApiKey,
		YandexGPTApiKey:  cfg.YandexGPTApiKey,
		YandexIAMToken:   cfg.YandexIAMToken,
		YandexFolderID:   cfg.YandexFolderID,
		YandexGPTModel:   cfg.YandexGPTModel,
	})

	// Initialize Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:      "GIBP Medical API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: v1.ErrorHandler,
	})

	// Middleware
	fiberApp.Use(recover.New())
	fiberApp.Use(fiberlogger.New())
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	auditMiddleware := middleware.NewAuditMiddleware(repos.AuditLog)

	// Setup routes
	v1.SetupRoutes(fiberApp, v1.RouterDeps{
		Services:        services,
		AuthMiddleware:  authMiddleware,
		AuditMiddleware: auditMiddleware,
	})

	// Health check
	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	// Graceful shutdown
	go func() {
		if err := fiberApp.Listen(":" + cfg.Port); err != nil {
			zapLogger.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := fiberApp.ShutdownWithContext(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", "error", err)
	}

	zapLogger.Info("Server exited properly")
}

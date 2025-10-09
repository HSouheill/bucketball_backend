package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/HSouheil/bucketball_backend/config"
	"github.com/HSouheil/bucketball_backend/repositories"
	"github.com/HSouheil/bucketball_backend/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting %s v%s in %s mode", cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Initialize database connections
	mongoClient := config.InitMongoDB(cfg)
	redisClient := config.InitRedis(cfg)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(mongoClient, cfg)
	authRepo := repositories.NewAuthRepository(redisClient)

	// Initialize routes
	routes.SetupRoutes(e, userRepo, authRepo)

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		config.CloseDatabases()
		os.Exit(0)
	}()

	// Start server
	port := cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

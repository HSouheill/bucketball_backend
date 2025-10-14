package routes

import (
	"time"

	"github.com/HSouheil/bucketball_backend/config"
	"github.com/HSouheil/bucketball_backend/controllers"
	"github.com/HSouheil/bucketball_backend/middleware"
	"github.com/HSouheil/bucketball_backend/repositories"
	"github.com/HSouheil/bucketball_backend/services"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoutes configures all routes
func SetupRoutes(e *echo.Echo, userRepo *repositories.UserRepository, authRepo *repositories.AuthRepository, db *mongo.Database) {
	// Get config
	cfg := config.GetConfig()

	// Initialize repositories
	otpRepo := repositories.NewOTPRepository(db)

	// Initialize services
	emailService := services.NewEmailService(&cfg.Email)
	otpService := services.NewOTPService(otpRepo, emailService)
	authService := services.NewAuthService(userRepo, authRepo, otpService)
	userService := services.NewUserService(userRepo)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	adminController := controllers.NewAdminController(authService)

	// API v1 group
	v1 := e.Group("/api")

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok", "message": "BucketBall API is running"})
	})

	// Static file serving for uploads
	e.Static("/uploads", "uploads")

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.POST("/register", authController.Register, middleware.RateLimitMiddleware(authRepo, 5, time.Minute))
	auth.POST("/verify-email", authController.VerifyEmail, middleware.RateLimitMiddleware(authRepo, 10, time.Minute))
	auth.POST("/resend-otp", authController.ResendOTP, middleware.RateLimitMiddleware(authRepo, 5, time.Minute))
	auth.POST("/login", authController.Login, middleware.RateLimitMiddleware(authRepo, 10, time.Minute))
	auth.POST("/logout", authController.Logout, middleware.AuthMiddleware(authRepo))

	// User routes (protected)
	users := v1.Group("/users")
	users.Use(middleware.AuthMiddleware(authRepo))
	users.Use(middleware.AuthRateLimitMiddleware(authRepo, 100, time.Hour))

	users.GET("/profile", authController.GetProfile)
	users.PUT("/profile", authController.UpdateProfile)

	// Admin routes
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(authRepo))
	admin.Use(middleware.AdminMiddleware())
	admin.Use(middleware.AuthRateLimitMiddleware(authRepo, 200, time.Hour))

	admin.GET("/users", userController.GetUsers)
	admin.GET("/users/:id", userController.GetUser)
	admin.PUT("/users/:id", userController.UpdateUser)
	admin.DELETE("/users/:id", userController.DeleteUser)
	admin.PATCH("/users/:id/toggle-status", userController.ToggleUserStatus)

	// Rate limit management endpoints
	admin.GET("/rate-limit/info", adminController.GetRateLimitInfo)
	admin.POST("/rate-limit/reset", adminController.ResetRateLimit)
}

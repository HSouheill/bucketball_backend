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

	// Initialize repositories
	gameRepo := repositories.NewGameRepository(db)

	// Initialize services
	emailService := services.NewEmailService(&cfg.Email)
	otpService := services.NewOTPService(otpRepo, emailService)
	authService := services.NewAuthService(userRepo, authRepo, otpService)
	userService := services.NewUserService(userRepo)
	referralService := services.NewReferralService(userRepo)
	paymentService := services.NewPaymentService(userRepo, referralService)
	gameService := services.NewGameService(gameRepo, userRepo)

	// Initialize controllers
	authController := controllers.NewAuthController(authService, paymentService)
	userController := controllers.NewUserController(userService)
	adminController := controllers.NewAdminController(authService)
	gameController := controllers.NewGameController(gameService)

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
	users.GET("/referral-stats", authController.GetReferralStats)
	users.POST("/payment", authController.ProcessPayment)

	// Game routes (protected)
	games := v1.Group("/games")
	games.Use(middleware.AuthMiddleware(authRepo))
	games.Use(middleware.AuthRateLimitMiddleware(authRepo, 50, time.Hour))

	games.GET("/state", gameController.GetGameState)
	games.POST("/bet", gameController.PlaceBet)
	games.POST("/:id/play", gameController.PlayGame)
	games.GET("/history", gameController.GetGameHistory)
	games.GET("/stats", gameController.GetUserGameStats)
	games.GET("/balls", gameController.GetAvailableBalls)
	games.GET("/baskets", gameController.GetAvailableBaskets)

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

	// Admin game management endpoints
	admin.GET("/games/stats", gameController.GetGameStats)
	admin.GET("/games/house-wallet", gameController.GetHouseWallet)
	admin.POST("/games/:id/simulate", gameController.SimulateOtherPlayers)
}

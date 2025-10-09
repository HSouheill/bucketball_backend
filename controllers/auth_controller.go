package controllers

import (
	"net/http"
	"strings"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/services"
	"github.com/HSouheil/bucketball_backend/utils"
	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register handles user registration
func (ac *AuthController) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request data", err)
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	ctx := c.Request().Context()
	user, token, err := ac.authService.Register(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "already taken") {
			return utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to register user", err)
	}

	// Return response
	userResponse := models.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		ProfilePic:  user.ProfilePic,
		DOB:         user.DOB,
		PhoneNumber: user.PhoneNumber,
		Location:    user.Location,
		Balance:     user.Balance,
		Withdraw:    user.Withdraw,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	authResponse := models.AuthResponse{
		Token: token,
		User:  userResponse,
	}

	return utils.SuccessResponse(c, "User registered successfully", authResponse)
}

// Login handles user login
func (ac *AuthController) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request data", err)
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	ctx := c.Request().Context()
	clientIP := c.RealIP()
	user, token, err := ac.authService.Login(ctx, &req, clientIP)
	if err != nil {
		// Check if it's a rate limit error
		if strings.Contains(err.Error(), "too many login attempts") {
			return utils.ErrorResponse(c, http.StatusTooManyRequests, err.Error(), nil)
		}
		return utils.UnauthorizedResponse(c, err.Error())
	}

	// Return response
	userResponse := models.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		ProfilePic:  user.ProfilePic,
		DOB:         user.DOB,
		PhoneNumber: user.PhoneNumber,
		Location:    user.Location,
		Balance:     user.Balance,
		Withdraw:    user.Withdraw,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	authResponse := models.AuthResponse{
		Token: token,
		User:  userResponse,
	}

	return utils.SuccessResponse(c, "Login successful", authResponse)
}

// Logout handles user logout
func (ac *AuthController) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return utils.UnauthorizedResponse(c, "Authorization header required")
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix
	ctx := c.Request().Context()

	if err := ac.authService.Logout(ctx, tokenString); err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to logout", err)
	}

	return utils.SuccessResponse(c, "Logout successful", nil)
}

// GetProfile gets the current user's profile
func (ac *AuthController) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)
	ctx := c.Request().Context()

	user, err := ac.authService.GetUserByID(ctx, userID)
	if err != nil {
		return utils.NotFoundResponse(c, "User not found")
	}

	userResponse := models.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		ProfilePic:  user.ProfilePic,
		DOB:         user.DOB,
		PhoneNumber: user.PhoneNumber,
		Location:    user.Location,
		Balance:     user.Balance,
		Withdraw:    user.Withdraw,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return utils.SuccessResponse(c, "Profile retrieved successfully", userResponse)
}

// UpdateProfile updates the current user's profile
func (ac *AuthController) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request data", err)
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	ctx := c.Request().Context()
	if err := ac.authService.UpdateUser(ctx, userID, &req); err != nil {
		if strings.Contains(err.Error(), "already taken") {
			return utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "no fields") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to update profile", err)
	}

	return utils.SuccessResponse(c, "Profile updated successfully", nil)
}

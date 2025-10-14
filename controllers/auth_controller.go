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

	// Parse form data
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil { // 10MB max
		// Try binding as JSON if not multipart
		if err := c.Bind(&req); err != nil {
			return utils.ValidationErrorResponse(c, "Invalid request data", err)
		}
	} else {
		// Bind form fields
		req.Email = c.FormValue("email")
		req.Username = c.FormValue("username")
		req.Password = c.FormValue("password")
		req.FirstName = c.FormValue("first_name")
		req.LastName = c.FormValue("last_name")
		req.DOB = c.FormValue("dob")
		req.PhoneNumber = c.FormValue("phone_number")

		// Handle profile picture upload
		file, err := c.FormFile("profile_pic")
		if err == nil && file != nil {
			// Upload the file
			filePath, uploadErr := utils.UploadFile(file, "uploads/users")
			if uploadErr != nil {
				return utils.ValidationErrorResponse(c, "Failed to upload profile picture", uploadErr)
			}
			req.ProfilePic = filePath
		}
	}

	if err := utils.ValidateStruct(&req); err != nil {
		// Clean up uploaded file if validation fails
		if req.ProfilePic != "" {
			utils.DeleteFile(req.ProfilePic)
		}
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	ctx := c.Request().Context()
	user, _, err := ac.authService.Register(ctx, &req)
	if err != nil {
		// Clean up uploaded file if registration fails
		if req.ProfilePic != "" {
			utils.DeleteFile(req.ProfilePic)
		}
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "already taken") {
			return utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to register user", err)
	}

	// Return response without token - user needs to verify email first
	response := map[string]interface{}{
		"email":       user.Email,
		"profile_pic": user.ProfilePic,
		"message":     "Registration successful. Please check your email for the OTP to verify your account.",
	}

	return utils.SuccessResponse(c, "Registration successful. OTP sent to your email.", response)
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
		ID:              user.ID,
		Email:           user.Email,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		ProfilePic:      user.ProfilePic,
		DOB:             user.DOB,
		PhoneNumber:     user.PhoneNumber,
		Location:        user.Location,
		Balance:         user.Balance,
		Withdraw:        user.Withdraw,
		Role:            user.Role,
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
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
		ID:              user.ID,
		Email:           user.Email,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		ProfilePic:      user.ProfilePic,
		DOB:             user.DOB,
		PhoneNumber:     user.PhoneNumber,
		Location:        user.Location,
		Balance:         user.Balance,
		Withdraw:        user.Withdraw,
		Role:            user.Role,
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
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

// VerifyEmail handles email verification with OTP
func (ac *AuthController) VerifyEmail(c echo.Context) error {
	var req models.VerifyOTPRequest
	if err := c.Bind(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request data", err)
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	ctx := c.Request().Context()
	token, err := ac.authService.VerifyEmailAndGenerateToken(ctx, req.Email, req.OTP)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "used") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to verify email", err)
	}

	// Get updated user
	user, err := ac.authService.GetUserByID(ctx, "")
	if err == nil {
		userResponse := models.UserResponse{
			ID:              user.ID,
			Email:           user.Email,
			Username:        user.Username,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			ProfilePic:      user.ProfilePic,
			DOB:             user.DOB,
			PhoneNumber:     user.PhoneNumber,
			Location:        user.Location,
			Balance:         user.Balance,
			Withdraw:        user.Withdraw,
			Role:            user.Role,
			IsActive:        user.IsActive,
			IsEmailVerified: user.IsEmailVerified,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
		}

		authResponse := models.AuthResponse{
			Token: token,
			User:  userResponse,
		}

		return utils.SuccessResponse(c, "Email verified successfully", authResponse)
	}

	// Fallback if user retrieval fails
	response := map[string]interface{}{
		"token": token,
	}

	return utils.SuccessResponse(c, "Email verified successfully", response)
}

// ResendOTP handles resending OTP
func (ac *AuthController) ResendOTP(c echo.Context) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.Bind(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request data", err)
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	ctx := c.Request().Context()
	if err := ac.authService.ResendOTP(ctx, req.Email); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.NotFoundResponse(c, "User not found")
		}
		if strings.Contains(err.Error(), "already verified") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to resend OTP", err)
	}

	return utils.SuccessResponse(c, "OTP resent successfully", nil)
}

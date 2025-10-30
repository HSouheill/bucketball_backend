package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/services"
	"github.com/HSouheil/bucketball_backend/utils"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new user controller
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetUsers gets a list of users with pagination
func (uc *UserController) GetUsers(c echo.Context) error {
	// Parse pagination parameters
	page, _ := strconv.ParseInt(c.QueryParam("page"), 10, 64)
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)

	ctx := c.Request().Context()
	users, total, err := uc.userService.GetUsers(ctx, page, limit)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get users", err)
	}

	// Convert to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, models.UserResponse{
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
			ReferralCode:    user.ReferralCode,
			ReferredBy:      user.ReferredBy,
			ReferralEarnings: user.ReferralEarnings,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
		})
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	response := map[string]interface{}{
		"users": userResponses,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	}

	return utils.SuccessResponse(c, "Users retrieved successfully", response)
}

// GetUser gets a user by ID
func (uc *UserController) GetUser(c echo.Context) error {
	userID := c.Param("id")
	ctx := c.Request().Context()

	user, err := uc.userService.GetUserByID(ctx, userID)
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
		ReferralCode:    user.ReferralCode,
		ReferredBy:      user.ReferredBy,
		ReferralEarnings: user.ReferralEarnings,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}

	return utils.SuccessResponse(c, "User retrieved successfully", userResponse)
}

// UpdateUser updates a user (admin only)
func (uc *UserController) UpdateUser(c echo.Context) error {
	userID := c.Param("id")

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request data", err)
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	ctx := c.Request().Context()
	if err := uc.userService.UpdateUser(ctx, userID, &req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.NotFoundResponse(c, err.Error())
		}
		if strings.Contains(err.Error(), "already taken") {
			return utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "no fields") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to update user", err)
	}

	return utils.SuccessResponse(c, "User updated successfully", nil)
}

// DeleteUser deletes a user (admin only)
func (uc *UserController) DeleteUser(c echo.Context) error {
	userID := c.Param("id")
	ctx := c.Request().Context()

	if err := uc.userService.DeleteUser(ctx, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.InternalServerErrorResponse(c, "Failed to delete user", err)
	}

	return utils.SuccessResponse(c, "User deleted successfully", nil)
}

// ToggleUserStatus toggles user active status (admin only)
func (uc *UserController) ToggleUserStatus(c echo.Context) error {
	userID := c.Param("id")
	ctx := c.Request().Context()

	newStatus, err := uc.userService.ToggleUserStatus(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.InternalServerErrorResponse(c, "Failed to update user status", err)
	}

	status := "activated"
	if !newStatus {
		status = "deactivated"
	}

	return utils.SuccessResponse(c, "User "+status+" successfully", nil)
}

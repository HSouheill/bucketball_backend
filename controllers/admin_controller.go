package controllers

import (
	"net/http"

	"github.com/HSouheil/bucketball_backend/services"
	"github.com/HSouheil/bucketball_backend/utils"
	"github.com/labstack/echo/v4"
)

type AdminController struct {
	authService *services.AuthService
}

// NewAdminController creates a new admin controller
func NewAdminController(authService *services.AuthService) *AdminController {
	return &AdminController{
		authService: authService,
	}
}

// GetRateLimitInfo gets rate limit information for debugging
func (ac *AdminController) GetRateLimitInfo(c echo.Context) error {
	email := c.QueryParam("email")
	ip := c.QueryParam("ip")

	if email == "" && ip == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Either email or IP must be provided", nil)
	}

	ctx := c.Request().Context()
	info, err := ac.authService.GetRateLimitInfo(ctx, email, ip)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get rate limit info", err)
	}

	return utils.SuccessResponse(c, "Rate limit info retrieved", info)
}

// ResetRateLimit resets rate limits for an email/IP
func (ac *AdminController) ResetRateLimit(c echo.Context) error {
	email := c.QueryParam("email")
	ip := c.QueryParam("ip")

	if email == "" && ip == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Either email or IP must be provided", nil)
	}

	ctx := c.Request().Context()
	if err := ac.authService.ResetRateLimit(ctx, email, ip); err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to reset rate limit", err)
	}

	message := "Rate limit reset successfully"
	if email != "" && ip != "" {
		message = "Rate limits reset for both email and IP"
	} else if email != "" {
		message = "Rate limit reset for email"
	} else {
		message = "Rate limit reset for IP"
	}

	return utils.SuccessResponse(c, message, nil)
}

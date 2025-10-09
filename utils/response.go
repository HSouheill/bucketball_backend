package utils

import (
	"net/http"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/labstack/echo/v4"
)

// SuccessResponse sends a success response
func SuccessResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c echo.Context, statusCode int, message string, err error) error {
	response := models.APIResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return c.JSON(statusCode, response)
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c echo.Context, message string, err error) error {
	return ErrorResponse(c, http.StatusBadRequest, message, err)
}

// UnauthorizedResponse sends an unauthorized response
func UnauthorizedResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// ForbiddenResponse sends a forbidden response
func ForbiddenResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusForbidden, message, nil)
}

// NotFoundResponse sends a not found response
func NotFoundResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusNotFound, message, nil)
}

// InternalServerErrorResponse sends an internal server error response
func InternalServerErrorResponse(c echo.Context, message string, err error) error {
	return ErrorResponse(c, http.StatusInternalServerError, message, err)
}

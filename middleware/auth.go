package middleware

import (
	"strings"

	"github.com/HSouheil/bucketball_backend/repositories"
	"github.com/HSouheil/bucketball_backend/security"
	"github.com/HSouheil/bucketball_backend/utils"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(authRepo *repositories.AuthRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return utils.UnauthorizedResponse(c, "Authorization header required")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return utils.UnauthorizedResponse(c, "Bearer token required")
			}

			// Check if token is blacklisted
			ctx := c.Request().Context()
			isBlacklisted, err := authRepo.IsTokenBlacklisted(ctx, tokenString)
			if err != nil {
				return utils.InternalServerErrorResponse(c, "Failed to check token status", err)
			}
			if isBlacklisted {
				return utils.UnauthorizedResponse(c, "Token has been revoked")
			}

			// Validate token
			claims, err := security.ValidateToken(tokenString)
			if err != nil {
				return utils.UnauthorizedResponse(c, "Invalid token")
			}

			// Set user info in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_username", claims.Username)
			c.Set("user_role", claims.Role)

			return next(c)
		}
	}
}

// AdminMiddleware checks if user is admin
func AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("user_role").(string)
			if role != "admin" {
				return utils.ForbiddenResponse(c, "Admin access required")
			}
			return next(c)
		}
	}
}

// OptionalAuthMiddleware validates JWT token if present
func OptionalAuthMiddleware(authRepo *repositories.AuthRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				if tokenString != authHeader {
					// Check if token is blacklisted
					ctx := c.Request().Context()
					isBlacklisted, err := authRepo.IsTokenBlacklisted(ctx, tokenString)
					if err == nil && !isBlacklisted {
						// Validate token
						if claims, err := security.ValidateToken(tokenString); err == nil {
							// Set user info in context
							c.Set("user_id", claims.UserID)
							c.Set("user_email", claims.Email)
							c.Set("user_username", claims.Username)
							c.Set("user_role", claims.Role)
						}
					}
				}
			}
			return next(c)
		}
	}
}

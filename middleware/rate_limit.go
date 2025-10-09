package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/HSouheil/bucketball_backend/repositories"
	"github.com/labstack/echo/v4"
)

// RateLimitMiddleware implements rate limiting using Redis
func RateLimitMiddleware(authRepo *repositories.AuthRepository, requests int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := c.RealIP()
			key := fmt.Sprintf("rate_limit:%s", clientIP)

			ctx := c.Request().Context()

			// Get current count
			count, err := authRepo.GetRedis().Get(ctx, key).Int()
			if err != nil && err.Error() != "redis: nil" {
				// If there's an error (not nil), continue without rate limiting
				return next(c)
			}

			if count >= requests {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"success": false,
					"message": "Rate limit exceeded",
					"error":   fmt.Sprintf("Maximum %d requests per %v", requests, window),
				})
			}

			// Increment counter
			pipe := authRepo.GetRedis().Pipeline()
			pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, window)
			_, err = pipe.Exec(ctx)
			if err != nil {
				// If there's an error, continue without rate limiting
				return next(c)
			}

			return next(c)
		}
	}
}

// AuthRateLimitMiddleware implements rate limiting for authenticated users
func AuthRateLimitMiddleware(authRepo *repositories.AuthRepository, requests int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Get("user_id")
			if userID == nil {
				return next(c)
			}

			key := fmt.Sprintf("auth_rate_limit:%s", userID)
			ctx := c.Request().Context()

			// Get current count
			count, err := authRepo.GetRedis().Get(ctx, key).Int()
			if err != nil && err.Error() != "redis: nil" {
				return next(c)
			}

			if count >= requests {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"success": false,
					"message": "Rate limit exceeded",
					"error":   fmt.Sprintf("Maximum %d requests per %v", requests, window),
				})
			}

			// Increment counter
			pipe := authRepo.GetRedis().Pipeline()
			pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, window)
			_, err = pipe.Exec(ctx)
			if err != nil {
				return next(c)
			}

			return next(c)
		}
	}
}

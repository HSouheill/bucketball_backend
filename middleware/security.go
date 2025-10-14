package middleware

import (
	"github.com/HSouheil/bucketball_backend/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent clickjacking
			c.Response().Header().Set("X-Frame-Options", "DENY")

			// Prevent MIME sniffing
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// Enable XSS protection
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Strict Transport Security (HSTS) - only in production
			cfg := config.GetConfig()
			if cfg.App.Environment == "production" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			// Content Security Policy
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'")

			// Referrer Policy
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			return next(c)
		}
	}
}

// HTTPSRedirectMiddleware redirects HTTP to HTTPS in production
func HTTPSRedirectMiddleware() echo.MiddlewareFunc {
	cfg := config.GetConfig()

	// Only enforce HTTPS in production
	if cfg.App.Environment == "production" {
		return middleware.HTTPSRedirect()
	}

	// In development, just pass through
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}

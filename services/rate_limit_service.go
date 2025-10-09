package services

import (
	"context"
	"fmt"
	"time"

	"github.com/HSouheil/bucketball_backend/repositories"
	"github.com/redis/go-redis/v9"
)

type RateLimitService struct {
	authRepo *repositories.AuthRepository
}

// NewRateLimitService creates a new rate limit service
func NewRateLimitService(authRepo *repositories.AuthRepository) *RateLimitService {
	return &RateLimitService{
		authRepo: authRepo,
	}
}

// LoginAttempt represents a login attempt
type LoginAttempt struct {
	Email     string    `json:"email"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	MaxAttempts     int           // Maximum attempts allowed
	WindowDuration  time.Duration // Time window for rate limiting
	LockoutDuration time.Duration // Duration to lock out after max attempts
}

// Default login rate limit configuration
var DefaultLoginRateLimit = RateLimitConfig{
	MaxAttempts:     5,                // 5 attempts
	WindowDuration:  15 * time.Minute, // within 15 minutes
	LockoutDuration: 30 * time.Minute, // lockout for 30 minutes
}

// CheckLoginRateLimit checks if login is allowed for the given email/IP
func (s *RateLimitService) CheckLoginRateLimit(ctx context.Context, email, ip string) (bool, time.Duration, error) {
	config := DefaultLoginRateLimit

	// Check email-based rate limit
	emailKey := fmt.Sprintf("login_attempts:email:%s", email)
	emailAllowed, emailTimeLeft, err := s.checkRateLimit(ctx, emailKey, config)
	if err != nil {
		return false, 0, err
	}

	// Check IP-based rate limit
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
	ipAllowed, ipTimeLeft, err := s.checkRateLimit(ctx, ipKey, config)
	if err != nil {
		return false, 0, err
	}

	// Both email and IP must be allowed
	if !emailAllowed || !ipAllowed {
		// Return the longer of the two timeouts
		if emailTimeLeft > ipTimeLeft {
			return false, emailTimeLeft, nil
		}
		return false, ipTimeLeft, nil
	}

	return true, 0, nil
}

// RecordLoginAttempt records a login attempt
func (s *RateLimitService) RecordLoginAttempt(ctx context.Context, email, ip string, success bool) error {
	config := DefaultLoginRateLimit

	// Record email-based attempt
	emailKey := fmt.Sprintf("login_attempts:email:%s", email)
	if err := s.recordAttempt(ctx, emailKey, config, success); err != nil {
		return err
	}

	// Record IP-based attempt
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
	return s.recordAttempt(ctx, ipKey, config, success)
}

// checkRateLimit checks if rate limit is exceeded
func (s *RateLimitService) checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (bool, time.Duration, error) {
	// Check if currently locked out
	lockoutKey := fmt.Sprintf("lockout:%s", key)
	lockoutExists, err := s.authRepo.GetRedis().Exists(ctx, lockoutKey).Result()
	if err != nil {
		return false, 0, err
	}

	if lockoutExists > 0 {
		// Get remaining lockout time
		ttl, err := s.authRepo.GetRedis().TTL(ctx, lockoutKey).Result()
		if err != nil {
			return false, 0, err
		}
		return false, ttl, nil
	}

	// Check attempt count
	count, err := s.authRepo.GetRedis().Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, 0, err
	}

	if count >= config.MaxAttempts {
		// Set lockout
		if err := s.authRepo.GetRedis().Set(ctx, lockoutKey, "1", config.LockoutDuration).Err(); err != nil {
			return false, 0, err
		}
		return false, config.LockoutDuration, nil
	}

	return true, 0, nil
}

// recordAttempt records an attempt and manages rate limiting
func (s *RateLimitService) recordAttempt(ctx context.Context, key string, config RateLimitConfig, success bool) error {
	pipe := s.authRepo.GetRedis().Pipeline()

	if success {
		// On successful login, clear the attempt counter
		pipe.Del(ctx, key)
		pipe.Del(ctx, fmt.Sprintf("lockout:%s", key))
	} else {
		// On failed login, increment counter
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, config.WindowDuration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetLoginAttemptsInfo returns information about login attempts for debugging
func (s *RateLimitService) GetLoginAttemptsInfo(ctx context.Context, email, ip string) (map[string]interface{}, error) {
	emailKey := fmt.Sprintf("login_attempts:email:%s", email)
	ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)

	emailCount, _ := s.authRepo.GetRedis().Get(ctx, emailKey).Int()
	ipCount, _ := s.authRepo.GetRedis().Get(ctx, ipKey).Int()

	emailLockout, _ := s.authRepo.GetRedis().TTL(ctx, fmt.Sprintf("lockout:%s", emailKey)).Result()
	ipLockout, _ := s.authRepo.GetRedis().TTL(ctx, fmt.Sprintf("lockout:%s", ipKey)).Result()

	return map[string]interface{}{
		"email_attempts":          emailCount,
		"ip_attempts":             ipCount,
		"email_locked":            emailLockout > 0,
		"ip_locked":               ipLockout > 0,
		"email_lockout_remaining": emailLockout,
		"ip_lockout_remaining":    ipLockout,
	}, nil
}

// ResetLoginAttempts resets login attempts for an email/IP (admin function)
func (s *RateLimitService) ResetLoginAttempts(ctx context.Context, email, ip string) error {
	pipe := s.authRepo.GetRedis().Pipeline()

	if email != "" {
		emailKey := fmt.Sprintf("login_attempts:email:%s", email)
		pipe.Del(ctx, emailKey)
		pipe.Del(ctx, fmt.Sprintf("lockout:%s", emailKey))
	}

	if ip != "" {
		ipKey := fmt.Sprintf("login_attempts:ip:%s", ip)
		pipe.Del(ctx, ipKey)
		pipe.Del(ctx, fmt.Sprintf("lockout:%s", ipKey))
	}

	_, err := pipe.Exec(ctx)
	return err
}

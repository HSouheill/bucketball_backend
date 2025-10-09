package repositories

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type AuthRepository struct {
	client *redis.Client
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(client *redis.Client) *AuthRepository {
	return &AuthRepository{client: client}
}

// SetToken stores a token in Redis with expiration
func (r *AuthRepository) SetToken(ctx context.Context, token, userID string, expiration time.Duration) error {
	return r.client.Set(ctx, "token:"+token, userID, expiration).Err()
}

// GetToken gets a token from Redis
func (r *AuthRepository) GetToken(ctx context.Context, token string) (string, error) {
	return r.client.Get(ctx, "token:"+token).Result()
}

// DeleteToken removes a token from Redis
func (r *AuthRepository) DeleteToken(ctx context.Context, token string) error {
	return r.client.Del(ctx, "token:"+token).Err()
}

// SetRefreshToken stores a refresh token in Redis
func (r *AuthRepository) SetRefreshToken(ctx context.Context, refreshToken, userID string, expiration time.Duration) error {
	return r.client.Set(ctx, "refresh:"+refreshToken, userID, expiration).Err()
}

// GetRefreshToken gets a refresh token from Redis
func (r *AuthRepository) GetRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return r.client.Get(ctx, "refresh:"+refreshToken).Result()
}

// DeleteRefreshToken removes a refresh token from Redis
func (r *AuthRepository) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	return r.client.Del(ctx, "refresh:"+refreshToken).Err()
}

// SetBlacklistToken adds a token to the blacklist
func (r *AuthRepository) SetBlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	return r.client.Set(ctx, "blacklist:"+token, "1", expiration).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *AuthRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	result := r.client.Get(ctx, "blacklist:"+token)
	if result.Err() == redis.Nil {
		return false, nil
	}
	if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

// GetRedis returns the Redis client
func (r *AuthRepository) GetRedis() *redis.Client {
	return r.client
}

package services

import (
	"context"
	"errors"
	"time"

	"github.com/HSouheil/bucketball_backend/security"
)

// VerifyEmailAndGenerateToken verifies the OTP and generates a token for the user
func (s *AuthService) VerifyEmailAndGenerateToken(ctx context.Context, email, otp string) (string, error) {
	// Verify OTP
	if err := s.otpService.VerifyOTP(ctx, email, otp, "registration"); err != nil {
		return "", err
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Update user's email verification status
	updateData := map[string]interface{}{
		"is_email_verified": true,
		"updated_at":        time.Now(),
	}

	if err := s.userRepo.Update(ctx, user.ID, updateData); err != nil {
		return "", errors.New("failed to update user verification status")
	}

	// Generate token
	token, err := security.GenerateToken(user.ID.Hex(), user.Email, user.Username, user.Role)
	if err != nil {
		return "", err
	}

	// Store token in Redis
	if err := s.authRepo.SetToken(ctx, token, user.ID.Hex(), 24*time.Hour); err != nil {
		return "", err
	}

	return token, nil
}

// ResendOTP resends the OTP to the user's email
func (s *AuthService) ResendOTP(ctx context.Context, email string) error {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if email is already verified
	if user.IsEmailVerified {
		return errors.New("email is already verified")
	}

	// Generate and send new OTP
	if err := s.otpService.GenerateAndSendOTP(ctx, user.Email, user.Username, "registration"); err != nil {
		return err
	}

	return nil
}

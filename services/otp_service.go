package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/repositories"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	OTPTypeRegistration  = "registration"
	OTPTypePasswordReset = "password_reset"
	OTPExpiryMinutes     = 10
)

// OTPService handles OTP operations
type OTPService struct {
	otpRepo      *repositories.OTPRepository
	emailService *EmailService
}

// NewOTPService creates a new OTP service
func NewOTPService(otpRepo *repositories.OTPRepository, emailService *EmailService) *OTPService {
	return &OTPService{
		otpRepo:      otpRepo,
		emailService: emailService,
	}
}

// GenerateAndSendOTP generates a 6-digit OTP and sends it via email
func (s *OTPService) GenerateAndSendOTP(ctx context.Context, email, username, otpType string) error {
	// Generate 6-digit OTP
	code, err := s.generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Invalidate any old unused OTPs for this email and type
	if err := s.otpRepo.InvalidateOldOTPs(ctx, email, otpType); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to invalidate old OTPs: %v\n", err)
	}

	// Create OTP record
	otp := &models.OTP{
		Email:     email,
		Code:      code,
		Type:      otpType,
		ExpiresAt: time.Now().Add(OTPExpiryMinutes * time.Minute),
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return fmt.Errorf("failed to save OTP: %w", err)
	}

	// Send OTP via email
	if err := s.emailService.SendOTPEmail(email, username, code); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// VerifyOTP verifies an OTP code
func (s *OTPService) VerifyOTP(ctx context.Context, email, code, otpType string) error {
	// Get latest OTP for this email and type
	otp, err := s.otpRepo.GetLatestByEmailAndType(ctx, email, otpType)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no OTP found for this email")
		}
		return fmt.Errorf("failed to retrieve OTP: %w", err)
	}

	// Check if OTP matches
	if otp.Code != code {
		return errors.New("invalid OTP code")
	}

	// Check if OTP is already used
	if otp.IsUsed {
		return errors.New("OTP has already been used")
	}

	// Check if OTP is expired
	if otp.IsExpired() {
		return errors.New("OTP has expired")
	}

	// Mark OTP as used
	if err := s.otpRepo.MarkAsUsed(ctx, email, code, otpType); err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	return nil
}

// generateOTP generates a secure 6-digit OTP
func (s *OTPService) generateOTP() (string, error) {
	// Generate a random number between 100000 and 999999
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Add 100000 to ensure it's always 6 digits
	otp := n.Int64() + 100000

	return fmt.Sprintf("%06d", otp), nil
}

// CleanupExpiredOTPs removes expired OTPs from the database
func (s *OTPService) CleanupExpiredOTPs(ctx context.Context) error {
	return s.otpRepo.DeleteExpiredOTPs(ctx)
}

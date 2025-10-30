package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/repositories"
	"github.com/HSouheil/bucketball_backend/security"
	"github.com/HSouheil/bucketball_backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userRepo       *repositories.UserRepository
	authRepo       *repositories.AuthRepository
	rateLimitSvc   *RateLimitService
	otpService     *OTPService
	referralService *ReferralService
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repositories.UserRepository, authRepo *repositories.AuthRepository, otpService *OTPService) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		authRepo:        authRepo,
		rateLimitSvc:    NewRateLimitService(authRepo),
		otpService:      otpService,
		referralService: NewReferralService(userRepo),
	}
}

// Register registers a new user and sends OTP for email verification
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, string, error) {
	// Sanitize inputs to prevent XSS
	req.Email = utils.SanitizeEmail(req.Email)
	req.Username = utils.SanitizeUsername(req.Username)
	req.FirstName = utils.SanitizeString(req.FirstName)
	req.LastName = utils.SanitizeString(req.LastName)
	req.PhoneNumber = utils.SanitizeString(req.PhoneNumber)

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, "", errors.New("user with this email already exists")
	}

	existingUser, _ = s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, "", errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, "", err
	}

	// Parse DOB if provided
	var dob *time.Time
	if req.DOB != "" {
		if parsedDOB, err := time.Parse("2006-01-02", req.DOB); err == nil {
			dob = &parsedDOB
		}
	}

	// Validate referral code if provided
	var referredBy *primitive.ObjectID
	if req.ReferralCode != "" {
		referrer, err := s.referralService.ValidateReferralCode(ctx, req.ReferralCode)
		if err != nil {
			return nil, "", fmt.Errorf("invalid referral code: %v", err)
		}
		if referrer != nil {
			referredBy = &referrer.ID
		}
	}

	// Generate unique referral code for the new user
	userReferralCode, err := s.referralService.GenerateReferralCode(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate referral code: %v", err)
	}

	// Set default location if not provided
	location := models.Location{}
	if req.Location != nil {
		location = *req.Location
	}

	// Create user with email not verified
	user := &models.User{
		Email:            req.Email,
		Username:         req.Username,
		Password:         hashedPassword,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		ProfilePic:       req.ProfilePic,
		DOB:              dob,
		PhoneNumber:      req.PhoneNumber,
		Location:         location,
		Balance:          0.0,
		Withdraw:         0.0,
		Role:             "user",
		IsActive:         true,
		IsEmailVerified:  false,
		ReferralCode:     userReferralCode,
		ReferredBy:       referredBy,
		ReferralEarnings: 0.0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate and send OTP for email verification
	if err := s.otpService.GenerateAndSendOTP(ctx, user.Email, user.Username, "registration"); err != nil {
		// Log the error but don't fail the registration
		fmt.Printf("Warning: failed to send OTP email: %v\n", err)
		// Note: In production, you might want to return this error or handle it differently
	}

	// Don't generate token yet - user needs to verify email first
	// Return empty token to indicate verification is pending
	return user, "", nil
}

// Login authenticates a user with rate limiting
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, clientIP string) (*models.User, string, error) {
	// Sanitize email input
	req.Email = utils.SanitizeEmail(req.Email)

	// Check rate limit before attempting login
	allowed, timeLeft, err := s.rateLimitSvc.CheckLoginRateLimit(ctx, req.Email, clientIP)
	if err != nil {
		return nil, "", errors.New("rate limit check failed")
	}

	if !allowed {
		return nil, "", fmt.Errorf("too many login attempts. Please try again in %v", timeLeft.Round(time.Minute))
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Record failed attempt
			s.rateLimitSvc.RecordLoginAttempt(ctx, req.Email, clientIP, false)
			return nil, "", errors.New("invalid email or password")
		}
		return nil, "", err
	}

	// Check if user is active
	if !user.IsActive {
		// Record failed attempt
		s.rateLimitSvc.RecordLoginAttempt(ctx, req.Email, clientIP, false)
		return nil, "", errors.New("account is deactivated")
	}

	// Check if email is verified
	if !user.IsEmailVerified {
		// Record failed attempt
		s.rateLimitSvc.RecordLoginAttempt(ctx, req.Email, clientIP, false)
		return nil, "", errors.New("email not verified. Please check your email for the OTP code")
	}

	// Check password
	if !security.CheckPasswordHash(req.Password, user.Password) {
		// Record failed attempt
		s.rateLimitSvc.RecordLoginAttempt(ctx, req.Email, clientIP, false)
		return nil, "", errors.New("invalid email or password")
	}

	// Record successful attempt (clears rate limit counters)
	s.rateLimitSvc.RecordLoginAttempt(ctx, req.Email, clientIP, true)

	// Generate token
	token, err := security.GenerateToken(user.ID.Hex(), user.Email, user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	// Store token in Redis
	if err := s.authRepo.SetToken(ctx, token, user.ID.Hex(), 24*time.Hour); err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Logout logs out a user
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.authRepo.SetBlacklistToken(ctx, token, 24*time.Hour)
}

// GetUserByID gets a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return s.userRepo.GetByID(ctx, objectID)
}

// UpdateUser updates user profile
func (s *AuthService) UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Check if username is being updated and if it's available
	if req.Username != nil {
		existingUser, _ := s.userRepo.GetByUsername(ctx, *req.Username)
		if existingUser != nil && existingUser.ID != objectID {
			return errors.New("username already taken")
		}
	}

	// Prepare update data with sanitization
	updateData := make(map[string]interface{})
	if req.Username != nil {
		sanitized := utils.SanitizeUsername(*req.Username)
		updateData["username"] = sanitized
	}
	if req.FirstName != nil {
		sanitized := utils.SanitizeString(*req.FirstName)
		updateData["first_name"] = sanitized
	}
	if req.LastName != nil {
		sanitized := utils.SanitizeString(*req.LastName)
		updateData["last_name"] = sanitized
	}
	if req.ProfilePic != nil {
		updateData["profile_pic"] = *req.ProfilePic
	}
	if req.DOB != nil {
		if parsedDOB, err := time.Parse("2006-01-02", *req.DOB); err == nil {
			updateData["dob"] = parsedDOB
		}
	}
	if req.PhoneNumber != nil {
		sanitized := utils.SanitizeString(*req.PhoneNumber)
		updateData["phone_number"] = sanitized
	}
	if req.Location != nil {
		updateData["location"] = *req.Location
	}
	if req.Balance != nil {
		updateData["balance"] = *req.Balance
	}
	if req.Withdraw != nil {
		updateData["withdraw"] = *req.Withdraw
	}

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	updateData["updated_at"] = time.Now()

	return s.userRepo.Update(ctx, objectID, updateData)
}

// GetRateLimitInfo gets rate limit information for debugging
func (s *AuthService) GetRateLimitInfo(ctx context.Context, email, ip string) (map[string]interface{}, error) {
	return s.rateLimitSvc.GetLoginAttemptsInfo(ctx, email, ip)
}

// ResetRateLimit resets rate limits for an email/IP
func (s *AuthService) ResetRateLimit(ctx context.Context, email, ip string) error {
	return s.rateLimitSvc.ResetLoginAttempts(ctx, email, ip)
}

// GetReferralStats gets referral statistics for a user
func (s *AuthService) GetReferralStats(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	return s.referralService.GetReferralStats(ctx, userID)
}

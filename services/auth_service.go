package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/repositories"
	"github.com/HSouheil/bucketball_backend/security"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userRepo     *repositories.UserRepository
	authRepo     *repositories.AuthRepository
	rateLimitSvc *RateLimitService
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repositories.UserRepository, authRepo *repositories.AuthRepository) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		authRepo:     authRepo,
		rateLimitSvc: NewRateLimitService(authRepo),
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, string, error) {
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

	// Set default location if not provided
	location := models.Location{}
	if req.Location != nil {
		location = *req.Location
	}

	// Create user
	user := &models.User{
		Email:       req.Email,
		Username:    req.Username,
		Password:    hashedPassword,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		ProfilePic:  req.ProfilePic,
		DOB:         dob,
		PhoneNumber: req.PhoneNumber,
		Location:    location,
		Balance:     0.0,
		Withdraw:    0.0,
		Role:        "user",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

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

// Login authenticates a user with rate limiting
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, clientIP string) (*models.User, string, error) {
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

	// Prepare update data
	updateData := make(map[string]interface{})
	if req.Username != nil {
		updateData["username"] = *req.Username
	}
	if req.FirstName != nil {
		updateData["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updateData["last_name"] = *req.LastName
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
		updateData["phone_number"] = *req.PhoneNumber
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

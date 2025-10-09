package services

import (
	"context"
	"errors"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUsers gets a list of users with pagination
func (s *UserService) GetUsers(ctx context.Context, page, limit int64) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	skip := (page - 1) * limit

	users, err := s.userRepo.List(ctx, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	return s.userRepo.GetByID(ctx, objectID)
}

// UpdateUser updates a user (admin only)
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Check if user exists
	_, err = s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return errors.New("user not found")
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

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	return s.userRepo.Update(ctx, objectID, updateData)
}

// DeleteUser deletes a user (admin only)
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Check if user exists
	_, err = s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(ctx, objectID)
}

// ToggleUserStatus toggles user active status (admin only)
func (s *UserService) ToggleUserStatus(ctx context.Context, userID string) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, errors.New("invalid user ID")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return false, errors.New("user not found")
	}

	// Toggle status
	newStatus := !user.IsActive
	updateData := map[string]interface{}{
		"is_active": newStatus,
	}

	if err := s.userRepo.Update(ctx, objectID, updateData); err != nil {
		return false, err
	}

	return newStatus, nil
}


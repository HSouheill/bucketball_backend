package services

import (
	"context"
	"fmt"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentService struct {
	userRepo        *repositories.UserRepository
	referralService *ReferralService
}

// NewPaymentService creates a new payment service
func NewPaymentService(userRepo *repositories.UserRepository, referralService *ReferralService) *PaymentService {
	return &PaymentService{
		userRepo:        userRepo,
		referralService: referralService,
	}
}

// ProcessPayment processes a payment and handles referral commissions
func (ps *PaymentService) ProcessPayment(ctx context.Context, userID primitive.ObjectID, amount float64, description string) error {
	// Get the user
	user, err := ps.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// Update user's balance
	updateData := map[string]interface{}{
		"balance":    user.Balance + amount,
		"updated_at": time.Now(),
	}

	if err := ps.userRepo.Update(ctx, user.ID, updateData); err != nil {
		return fmt.Errorf("failed to update user balance: %v", err)
	}

	// Process referral commission if applicable
	if err := ps.referralService.ProcessReferralCommission(ctx, userID, amount); err != nil {
		// Log the error but don't fail the payment
		fmt.Printf("Warning: failed to process referral commission: %v\n", err)
	}

	// Log the payment
	fmt.Printf("Payment processed: User %s received $%.2f. Description: %s\n",
		user.Email, amount, description)

	return nil
}

// ProcessWithdrawal processes a withdrawal request
func (ps *PaymentService) ProcessWithdrawal(ctx context.Context, userID primitive.ObjectID, amount float64, bankAccount string) error {
	// Get the user
	user, err := ps.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// Check if user has sufficient balance
	if user.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	// Update user's balance and withdrawal amount
	updateData := map[string]interface{}{
		"balance":    user.Balance - amount,
		"withdraw":   user.Withdraw + amount,
		"updated_at": time.Now(),
	}

	if err := ps.userRepo.Update(ctx, user.ID, updateData); err != nil {
		return fmt.Errorf("failed to process withdrawal: %v", err)
	}

	// Log the withdrawal
	fmt.Printf("Withdrawal processed: User %s withdrew $%.2f to account %s\n",
		user.Email, amount, bankAccount)

	return nil
}

// GetPaymentHistory gets payment history for a user (placeholder implementation)
func (ps *PaymentService) GetPaymentHistory(ctx context.Context, userID primitive.ObjectID) ([]models.Transaction, error) {
	// This would typically query a transactions collection
	// For now, return empty slice
	return []models.Transaction{}, nil
}

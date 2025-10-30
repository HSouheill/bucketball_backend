package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReferralService struct {
	userRepo *repositories.UserRepository
}

// NewReferralService creates a new referral service
func NewReferralService(userRepo *repositories.UserRepository) *ReferralService {
	return &ReferralService{
		userRepo: userRepo,
	}
}

// GenerateReferralCode generates a unique referral code
func (rs *ReferralService) GenerateReferralCode(ctx context.Context) (string, error) {
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		// Generate a random 8-character code
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}

		code := hex.EncodeToString(bytes)[:8]

		// Check if code already exists
		existingUser, err := rs.userRepo.GetByReferralCode(ctx, code)
		if err != nil && err.Error() != "no documents found" {
			return "", err
		}

		if existingUser == nil {
			return code, nil
		}
	}

	return "", errors.New("failed to generate unique referral code after multiple attempts")
}

// ValidateReferralCode validates if a referral code exists and returns the referrer
func (rs *ReferralService) ValidateReferralCode(ctx context.Context, referralCode string) (*models.User, error) {
	if referralCode == "" {
		return nil, nil // Empty referral code is allowed
	}

	referrer, err := rs.userRepo.GetByReferralCode(ctx, referralCode)
	if err != nil {
		return nil, errors.New("invalid referral code")
	}

	return referrer, nil
}

// CalculateCommission calculates the commission amount based on the payment amount
func (rs *ReferralService) CalculateCommission(paymentAmount float64) float64 {
	// 0.5% commission for every full $100 paid
	// Only pay commission for complete $100 increments
	fullHundreds := math.Floor(paymentAmount / 100.0)
	if fullHundreds < 1 {
		return 0 // No commission for amounts under $100
	}

	commissionRate := 0.005 // 0.5%
	commissionAmount := fullHundreds * 100 * commissionRate
	return math.Round(commissionAmount*100) / 100 // Round to 2 decimal places
}

// ProcessReferralCommission processes a referral commission when a referred user pays
func (rs *ReferralService) ProcessReferralCommission(ctx context.Context, referredUserID primitive.ObjectID, paymentAmount float64) error {
	// Get the referred user to find their referrer
	referredUser, err := rs.userRepo.GetByID(ctx, referredUserID)
	if err != nil {
		return fmt.Errorf("failed to get referred user: %v", err)
	}

	// Check if user was referred by someone
	if referredUser.ReferredBy == nil {
		return nil // No referrer, no commission
	}

	// Calculate commission
	commissionAmount := rs.CalculateCommission(paymentAmount)
	if commissionAmount <= 0 {
		return nil // No commission for small amounts
	}

	// Get the referrer
	referrer, err := rs.userRepo.GetByID(ctx, *referredUser.ReferredBy)
	if err != nil {
		return fmt.Errorf("failed to get referrer: %v", err)
	}

	// Update referrer's balance and referral earnings
	updateData := map[string]interface{}{
		"balance":           referrer.Balance + commissionAmount,
		"referral_earnings": referrer.ReferralEarnings + commissionAmount,
		"updated_at":        time.Now(),
	}

	if err := rs.userRepo.Update(ctx, referrer.ID, updateData); err != nil {
		return fmt.Errorf("failed to update referrer balance: %v", err)
	}

	// Create commission record
	commission := &models.ReferralCommission{
		ReferrerID:       referrer.ID,
		ReferredUserID:   referredUserID,
		OriginalAmount:   paymentAmount,
		CommissionRate:   0.005, // 0.5%
		CommissionAmount: commissionAmount,
		Description:      fmt.Sprintf("Referral commission from %s's payment", referredUser.Email),
		Status:           "completed",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Store commission record (you might want to create a separate repository for this)
	// For now, we'll just log it
	fmt.Printf("Referral Commission: Referrer %s earned $%.2f from referred user %s's payment of $%.2f\n",
		referrer.Email, commissionAmount, referredUser.Email, paymentAmount)

	// Log commission details for debugging
	_ = commission // Suppress unused variable warning

	return nil
}

// GetReferralStats gets referral statistics for a user
func (rs *ReferralService) GetReferralStats(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	user, err := rs.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get total referrals count (you might want to implement this in the repository)
	// For now, we'll return basic stats
	stats := map[string]interface{}{
		"referral_code":     user.ReferralCode,
		"referral_earnings": user.ReferralEarnings,
		"total_earnings":    user.Balance,
	}

	return stats, nil
}

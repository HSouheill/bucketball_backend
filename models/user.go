package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email           string             `json:"email" bson:"email" validate:"required,email"`
	Username        string             `json:"username" bson:"username" validate:"required,min=3,max=20"`
	Password        string             `json:"-" bson:"password" validate:"required,min=6"`
	FirstName       string             `json:"first_name" bson:"first_name" validate:"required,min=2,max=50"`
	LastName        string             `json:"last_name" bson:"last_name" validate:"required,min=2,max=50"`
	ProfilePic      string             `json:"profile_pic" bson:"profile_pic"`
	DOB             *time.Time         `json:"dob" bson:"dob" validate:"omitempty"`
	PhoneNumber     string             `json:"phone_number" bson:"phone_number" validate:"omitempty,min=10,max=15"`
	Location        Location           `json:"location" bson:"location"`
	Balance         float64            `json:"balance" bson:"balance" validate:"min=0"`
	Withdraw        float64            `json:"withdraw" bson:"withdraw" validate:"min=0"`
	Role            string             `json:"role" bson:"role" validate:"required,oneof=user admin"`
	IsActive        bool               `json:"is_active" bson:"is_active"`
	IsEmailVerified bool               `json:"is_email_verified" bson:"is_email_verified"`
	ReferralCode    string             `json:"referral_code" bson:"referral_code" validate:"required"`
	ReferredBy      *primitive.ObjectID `json:"referred_by" bson:"referred_by,omitempty"`
	ReferralEarnings float64           `json:"referral_earnings" bson:"referral_earnings" validate:"min=0"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// AddToBalance adds amount to user's balance
func (u *User) AddToBalance(amount float64) {
	u.Balance += amount
}

// SubtractFromBalance subtracts amount from user's balance
func (u *User) SubtractFromBalance(amount float64) error {
	if u.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}
	u.Balance -= amount
	return nil
}

// CanWithdraw checks if user can withdraw the specified amount
func (u *User) CanWithdraw(amount float64) bool {
	return u.Balance >= amount && u.Balance > 0
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsValidLocation checks if the user's location is complete
func (u *User) IsValidLocation() bool {
	return u.Location.Country != "" && u.Location.State != "" && u.Location.City != ""
}

// Location represents user's location information
type Location struct {
	Country    string `json:"country" bson:"country" validate:"omitempty,min=2,max=50"`
	State      string `json:"state" bson:"state" validate:"omitempty,min=2,max=50"`
	City       string `json:"city" bson:"city" validate:"omitempty,min=2,max=50"`
	PostalCode string `json:"postal_code" bson:"postal_code" validate:"omitempty,min=3,max=10"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID              primitive.ObjectID `json:"id"`
	Email           string             `json:"email"`
	Username        string             `json:"username"`
	FirstName       string             `json:"first_name"`
	LastName        string             `json:"last_name"`
	ProfilePic      string             `json:"profile_pic"`
	DOB             *time.Time         `json:"dob"`
	PhoneNumber     string             `json:"phone_number"`
	Location        Location           `json:"location"`
	Balance         float64            `json:"balance"`
	Withdraw        float64            `json:"withdraw"`
	Role            string             `json:"role"`
	IsActive        bool               `json:"is_active"`
	IsEmailVerified bool               `json:"is_email_verified"`
	ReferralCode    string             `json:"referral_code"`
	ReferredBy      *primitive.ObjectID `json:"referred_by"`
	ReferralEarnings float64           `json:"referral_earnings"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email        string    `json:"email" validate:"required,email"`
	Username     string    `json:"username" validate:"required,min=3,max=20"`
	Password     string    `json:"password" validate:"required,min=6"`
	FirstName    string    `json:"first_name" validate:"required,min=2,max=50"`
	LastName     string    `json:"last_name" validate:"required,min=2,max=50"`
	ProfilePic   string    `json:"profile_pic,omitempty"`
	DOB          string    `json:"dob,omitempty" validate:"omitempty"`
	PhoneNumber  string    `json:"phone_number,omitempty" validate:"omitempty,min=10,max=15"`
	Location     *Location `json:"location,omitempty"`
	ReferralCode string    `json:"referral_code,omitempty" validate:"omitempty,min=6,max=20"`
}

// UpdateUserRequest represents the user update request payload
type UpdateUserRequest struct {
	Username    *string   `json:"username,omitempty" validate:"omitempty,min=3,max=20"`
	FirstName   *string   `json:"first_name,omitempty" validate:"omitempty,min=2,max=50"`
	LastName    *string   `json:"last_name,omitempty" validate:"omitempty,min=2,max=50"`
	ProfilePic  *string   `json:"profile_pic,omitempty"`
	DOB         *string   `json:"dob,omitempty" validate:"omitempty"`
	PhoneNumber *string   `json:"phone_number,omitempty" validate:"omitempty,min=10,max=15"`
	Location    *Location `json:"location,omitempty"`
	Balance     *float64  `json:"balance,omitempty" validate:"omitempty,min=0"`
	Withdraw    *float64  `json:"withdraw,omitempty" validate:"omitempty,min=0"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// BalanceUpdateRequest represents a request to update user balance
type BalanceUpdateRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Type   string  `json:"type" validate:"required,oneof=add subtract"`
	Reason string  `json:"reason,omitempty"`
}

// WithdrawRequest represents a withdrawal request
type WithdrawRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	BankAccount string  `json:"bank_account" validate:"required"`
	Reason      string  `json:"reason,omitempty"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	Type        string             `json:"type" bson:"type" validate:"required,oneof=deposit withdrawal transfer"`
	Amount      float64            `json:"amount" bson:"amount" validate:"required,gt=0"`
	Balance     float64            `json:"balance" bson:"balance"`
	Description string             `json:"description" bson:"description"`
	Status      string             `json:"status" bson:"status" validate:"required,oneof=pending completed failed"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// ReferralCommission represents a referral commission transaction
type ReferralCommission struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ReferrerID      primitive.ObjectID `json:"referrer_id" bson:"referrer_id"`
	ReferredUserID  primitive.ObjectID `json:"referred_user_id" bson:"referred_user_id"`
	OriginalAmount  float64            `json:"original_amount" bson:"original_amount"`
	CommissionRate  float64            `json:"commission_rate" bson:"commission_rate"`
	CommissionAmount float64           `json:"commission_amount" bson:"commission_amount"`
	Description     string             `json:"description" bson:"description"`
	Status          string             `json:"status" bson:"status" validate:"required,oneof=pending completed failed"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

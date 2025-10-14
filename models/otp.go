package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OTP represents an OTP record in the system
type OTP struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Code      string             `json:"code" bson:"code"`
	Type      string             `json:"type" bson:"type"` // registration, password_reset, etc.
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	IsUsed    bool               `json:"is_used" bson:"is_used"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// VerifyOTPRequest represents the OTP verification request payload
type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

// IsExpired checks if the OTP is expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid checks if the OTP is valid (not expired and not used)
func (o *OTP) IsValid() bool {
	return !o.IsExpired() && !o.IsUsed
}

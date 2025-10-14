package repositories

import (
	"context"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OTPRepository handles OTP database operations
type OTPRepository struct {
	collection *mongo.Collection
}

// NewOTPRepository creates a new OTP repository
func NewOTPRepository(db *mongo.Database) *OTPRepository {
	return &OTPRepository{
		collection: db.Collection("otps"),
	}
}

// Create creates a new OTP record
func (r *OTPRepository) Create(ctx context.Context, otp *models.OTP) error {
	_, err := r.collection.InsertOne(ctx, otp)
	return err
}

// GetLatestByEmailAndType gets the latest OTP for an email and type
func (r *OTPRepository) GetLatestByEmailAndType(ctx context.Context, email, otpType string) (*models.OTP, error) {
	var otp models.OTP

	filter := bson.M{
		"email": email,
		"type":  otpType,
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	err := r.collection.FindOne(ctx, filter, opts).Decode(&otp)
	if err != nil {
		return nil, err
	}

	return &otp, nil
}

// MarkAsUsed marks an OTP as used
func (r *OTPRepository) MarkAsUsed(ctx context.Context, email, code, otpType string) error {
	filter := bson.M{
		"email": email,
		"code":  code,
		"type":  otpType,
	}

	update := bson.M{
		"$set": bson.M{
			"is_used": true,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteExpiredOTPs deletes all expired OTPs
func (r *OTPRepository) DeleteExpiredOTPs(ctx context.Context) error {
	filter := bson.M{
		"expires_at": bson.M{
			"$lt": time.Now(),
		},
	}

	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

// InvalidateOldOTPs invalidates all old OTPs for an email and type
func (r *OTPRepository) InvalidateOldOTPs(ctx context.Context, email, otpType string) error {
	filter := bson.M{
		"email":   email,
		"type":    otpType,
		"is_used": false,
	}

	update := bson.M{
		"$set": bson.M{
			"is_used": true,
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Admin represents an admin user in the system
type Admin struct {
	ID           primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Username     string               `json:"username" bson:"username" validate:"required,min=3,max=20"`
	Email        string               `json:"email" bson:"email" validate:"required,email"`
	Password     string               `json:"-" bson:"password" validate:"required,min=6"`
	ProfilePic   string               `json:"profile_pic" bson:"profile_pic"`
	Revenues     float64              `json:"revenues" bson:"revenues" validate:"min=0"`
	Transactions []primitive.ObjectID `json:"transactions" bson:"transactions"`
	Players      []primitive.ObjectID `json:"players" bson:"players"`
	Balance      float64              `json:"balance" bson:"balance" validate:"min=0"`
	Role         string               `json:"role" bson:"role" validate:"required,oneof=admin superadmin"`
	IsActive     bool                 `json:"is_active" bson:"is_active"`
	CreatedAt    time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at" bson:"updated_at"`
}

// AddToBalance adds amount to admin's balance
func (a *Admin) AddToBalance(amount float64) {
	a.Balance += amount
}

// SubtractFromBalance subtracts amount from admin's balance
func (a *Admin) SubtractFromBalance(amount float64) error {
	if a.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}
	a.Balance -= amount
	return nil
}

// AddRevenue adds revenue to the admin's total revenues
func (a *Admin) AddRevenue(amount float64) {
	a.Revenues += amount
}

// AddTransaction adds a transaction ID to the admin's transaction list
func (a *Admin) AddTransaction(transactionID primitive.ObjectID) {
	a.Transactions = append(a.Transactions, transactionID)
}

// AddPlayer adds a player/user ID to the admin's managed players list
func (a *Admin) AddPlayer(playerID primitive.ObjectID) {
	// Check if player already exists
	for _, id := range a.Players {
		if id == playerID {
			return
		}
	}
	a.Players = append(a.Players, playerID)
}

// RemovePlayer removes a player/user ID from the admin's managed players list
func (a *Admin) RemovePlayer(playerID primitive.ObjectID) {
	for i, id := range a.Players {
		if id == playerID {
			a.Players = append(a.Players[:i], a.Players[i+1:]...)
			return
		}
	}
}

// GetTotalPlayers returns the count of players managed by this admin
func (a *Admin) GetTotalPlayers() int {
	return len(a.Players)
}

// GetTotalTransactions returns the count of transactions
func (a *Admin) GetTotalTransactions() int {
	return len(a.Transactions)
}

// IsSuperAdmin checks if the admin has superadmin role
func (a *Admin) IsSuperAdmin() bool {
	return a.Role == "superadmin"
}

// AdminResponse represents the admin data returned in API responses
type AdminResponse struct {
	ID           primitive.ObjectID   `json:"id"`
	Username     string               `json:"username"`
	Email        string               `json:"email"`
	ProfilePic   string               `json:"profile_pic"`
	Revenues     float64              `json:"revenues"`
	Transactions []primitive.ObjectID `json:"transactions"`
	Players      []primitive.ObjectID `json:"players"`
	Balance      float64              `json:"balance"`
	Role         string               `json:"role"`
	IsActive     bool                 `json:"is_active"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// AdminRegisterRequest represents the admin registration request payload
type AdminRegisterRequest struct {
	Username   string `json:"username" validate:"required,min=3,max=20"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=6"`
	ProfilePic string `json:"profile_pic,omitempty"`
	Role       string `json:"role" validate:"required,oneof=admin superadmin"`
}

// AdminUpdateRequest represents the admin update request payload
type AdminUpdateRequest struct {
	Username   *string  `json:"username,omitempty" validate:"omitempty,min=3,max=20"`
	ProfilePic *string  `json:"profile_pic,omitempty"`
	Balance    *float64 `json:"balance,omitempty" validate:"omitempty,min=0"`
	Revenues   *float64 `json:"revenues,omitempty" validate:"omitempty,min=0"`
	Role       *string  `json:"role,omitempty" validate:"omitempty,oneof=admin superadmin"`
}

// AdminLoginRequest represents the admin login request payload
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// AdminAuthResponse represents the admin authentication response
type AdminAuthResponse struct {
	Token string        `json:"token"`
	Admin AdminResponse `json:"admin"`
}

// AdminRevenueUpdateRequest represents a request to update admin revenue
type AdminRevenueUpdateRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Type   string  `json:"type" validate:"required,oneof=add subtract"`
	Reason string  `json:"reason,omitempty"`
}

// AdminPlayerAssignmentRequest represents a request to assign/remove players from admin
type AdminPlayerAssignmentRequest struct {
	PlayerID primitive.ObjectID `json:"player_id" validate:"required"`
	Action   string             `json:"action" validate:"required,oneof=assign remove"`
}

// AdminStats represents statistics for an admin
type AdminStats struct {
	TotalRevenues     float64 `json:"total_revenues"`
	TotalPlayers      int     `json:"total_players"`
	TotalTransactions int     `json:"total_transactions"`
	CurrentBalance    float64 `json:"current_balance"`
}

// GetStats returns statistics for the admin
func (a *Admin) GetStats() AdminStats {
	return AdminStats{
		TotalRevenues:     a.Revenues,
		TotalPlayers:      len(a.Players),
		TotalTransactions: len(a.Transactions),
		CurrentBalance:    a.Balance,
	}
}

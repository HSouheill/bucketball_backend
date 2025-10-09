package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Game represents a basketball game
type Game struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title" validate:"required,min=3,max=100"`
	Description string             `json:"description" bson:"description" validate:"max=500"`
	Location    string             `json:"location" bson:"location" validate:"required,min=3,max=100"`
	Date        time.Time          `json:"date" bson:"date" validate:"required"`
	MaxPlayers  int                `json:"max_players" bson:"max_players" validate:"required,min=2,max=20"`
	CurrentPlayers int             `json:"current_players" bson:"current_players"`
	CreatedBy   primitive.ObjectID `json:"created_by" bson:"created_by" validate:"required"`
	Players     []primitive.ObjectID `json:"players" bson:"players"`
	Status      string             `json:"status" bson:"status" validate:"required,oneof=upcoming ongoing completed cancelled"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateGameRequest represents the game creation request payload
type CreateGameRequest struct {
	Title       string    `json:"title" validate:"required,min=3,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Location    string    `json:"location" validate:"required,min=3,max=100"`
	Date        time.Time `json:"date" validate:"required"`
	MaxPlayers  int       `json:"max_players" validate:"required,min=2,max=20"`
}

// UpdateGameRequest represents the game update request payload
type UpdateGameRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=500"`
	Location    *string    `json:"location,omitempty" validate:"omitempty,min=3,max=100"`
	Date        *time.Time `json:"date,omitempty" validate:"omitempty"`
	MaxPlayers  *int       `json:"max_players,omitempty" validate:"omitempty,min=2,max=20"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=upcoming ongoing completed cancelled"`
}

// JoinGameRequest represents the join game request payload
type JoinGameRequest struct {
	GameID primitive.ObjectID `json:"game_id" validate:"required"`
}

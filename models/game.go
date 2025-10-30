package models

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Ball represents a ball in the game
type Ball struct {
	ID    int    `json:"id" bson:"id"`
	Color string `json:"color" bson:"color"`
	Name  string `json:"name" bson:"name"`
}

// Basket represents a basket with its multiplier
type Basket struct {
	Value float64 `json:"value" bson:"value"`
	Color string  `json:"color" bson:"color"`
}

// Game represents a single game instance
type Game struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RoundNumber     int                `json:"round_number" bson:"round_number"`
	Status          string             `json:"status" bson:"status"` // pending, active, completed
	WinningBallID   *int               `json:"winning_ball_id" bson:"winning_ball_id,omitempty"`
	WinningBasketID *int               `json:"winning_basket_id" bson:"winning_basket_id,omitempty"`
	TotalBets       float64            `json:"total_bets" bson:"total_bets"`
	HouseWallet     float64            `json:"house_wallet" bson:"house_wallet"`
	AdminProfit     float64            `json:"admin_profit" bson:"admin_profit"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
	CompletedAt     *time.Time         `json:"completed_at" bson:"completed_at,omitempty"`
}

// Bet represents a user's bet on a specific ball
type Bet struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	GameID    primitive.ObjectID `json:"game_id" bson:"game_id"`
	BallID    int                `json:"ball_id" bson:"ball_id"`
	Amount    float64            `json:"amount" bson:"amount"`
	Status    string             `json:"status" bson:"status"` // pending, won, lost, pushed
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// GameResult represents the result of a game for a specific user
type GameResult struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID        primitive.ObjectID `json:"user_id" bson:"user_id"`
	GameID        primitive.ObjectID `json:"game_id" bson:"game_id"`
	BallID        int                `json:"ball_id" bson:"ball_id"`
	BallName      string             `json:"ball_name" bson:"ball_name"`
	BallColor     string             `json:"ball_color" bson:"ball_color"`
	BetAmount     float64            `json:"bet_amount" bson:"bet_amount"`
	Multiplier    float64            `json:"multiplier" bson:"multiplier"`
	WinAmount     float64            `json:"win_amount" bson:"win_amount"`
	Profit        float64            `json:"profit" bson:"profit"`
	BasketLanded  int                `json:"basket_landed" bson:"basket_landed"`
	Won           bool               `json:"won" bson:"won"`
	Pushed        bool               `json:"pushed" bson:"pushed"`
	WalletLimited bool               `json:"wallet_limited" bson:"wallet_limited"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
}

// PlaceBetRequest represents a request to place a bet
type PlaceBetRequest struct {
	BallBets map[int]float64 `json:"ball_bets" validate:"required,min=1"`
}

// GameState represents the current state of the game
type GameState struct {
	CurrentGame      *Game        `json:"current_game,omitempty"`
	AvailableBalls   []Ball       `json:"available_balls"`
	AvailableBaskets []Basket     `json:"available_baskets"`
	UserBalance      float64      `json:"user_balance"`
	HouseWallet      float64      `json:"house_wallet"`
	AdminProfit      float64      `json:"admin_profit"`
	TotalBets        float64      `json:"total_bets"`
	GameHistory      []GameResult `json:"game_history,omitempty"`
}

// HouseWallet represents the house wallet state
type HouseWallet struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Balance     float64            `json:"balance" bson:"balance"`
	AdminProfit float64            `json:"admin_profit" bson:"admin_profit"`
	TotalBets   float64            `json:"total_bets" bson:"total_bets"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// GetAvailableBalls returns the available balls for betting
func GetAvailableBalls() []Ball {
	return []Ball{
		{ID: 0, Color: "#FF6B6B", Name: "Red"},
		{ID: 1, Color: "#4ECDC4", Name: "Cyan"},
		{ID: 2, Color: "#FFE66D", Name: "Yellow"},
		{ID: 3, Color: "#95E1D3", Name: "Green"},
	}
}

// GetAvailableBaskets returns the available baskets with their multipliers
func GetAvailableBaskets() []Basket {
	return []Basket{
		{Value: 0.25, Color: "#e74c3c"},
		{Value: 0.50, Color: "#e67e22"},
		{Value: 0.75, Color: "#f39c12"},
		{Value: 1, Color: "#f1c40f"},
		{Value: 2, Color: "#2ecc71"},
		{Value: 4, Color: "#3498db"},
		{Value: 8, Color: "#9b59b6"},
		{Value: 10, Color: "#1abc9c"},
	}
}

// CalculateWinningBasket calculates which basket a ball should land in based on house wallet
func CalculateWinningBasket(playerBets map[int]float64, currentWallet float64) int {
	totalPlayerBets := 0.0
	for _, bet := range playerBets {
		totalPlayerBets += bet
	}

	maxAllowedWin := currentWallet * 0.20
	maxMultiplier := 0.0
	if totalPlayerBets > 0 {
		maxMultiplier = maxAllowedWin / totalPlayerBets
	}

	// Default weights for baskets
	weights := []int{35, 25, 15, 5, 8, 6, 4, 2}

	// Adjust weights based on wallet capacity
	if maxMultiplier < 8 {
		weights[7] = 0 // 8x basket
	}
	if maxMultiplier < 4 {
		weights[6] = 0 // 4x basket
	}
	if maxMultiplier < 2 {
		weights[5] = 0 // 2x basket
		weights[4] = 0 // 1.5x basket
	}

	// Redistribute weights if too many are zeroed out
	availableBaskets := 0
	for _, w := range weights {
		if w > 0 {
			availableBaskets++
		}
	}

	if availableBaskets < 4 {
		// Redistribute weights to available baskets
		newWeights := make([]int, 8)
		weightIndex := 0
		redistributedWeights := []int{40, 30, 20, 10}

		for i, w := range weights {
			if w > 0 {
				if weightIndex < len(redistributedWeights) {
					newWeights[i] = redistributedWeights[weightIndex]
				} else {
					newWeights[i] = 10
				}
				weightIndex++
			}
		}
		weights = newWeights
	}

	// Calculate total weight
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	if totalWeight == 0 {
		return 0 // Fallback to first basket
	}

	// Generate random number and select basket
	random := float64(time.Now().UnixNano() % int64(totalWeight))

	for i, weight := range weights {
		random -= float64(weight)
		if random <= 0 {
			return i
		}
	}

	return 0
}

// IsValidBetAmount checks if the bet amount is valid
func (b *Bet) IsValidBetAmount() bool {
	return b.Amount > 0
}

// CalculateWinAmount calculates the win amount based on multiplier
func (b *Bet) CalculateWinAmount(multiplier float64) float64 {
	return b.Amount * multiplier
}

// CalculateProfit calculates the profit/loss from the bet
func (b *Bet) CalculateProfit(multiplier float64) float64 {
	winAmount := b.CalculateWinAmount(multiplier)
	return winAmount - b.Amount
}

// IsWin checks if the bet is a win (multiplier >= 2)
func (b *Bet) IsWin(multiplier float64) bool {
	return multiplier >= 2.0
}

// IsPush checks if the bet is a push (multiplier == 1)
func (b *Bet) IsPush(multiplier float64) bool {
	return multiplier == 1.0
}

// IsLoss checks if the bet is a loss (multiplier < 1)
func (b *Bet) IsLoss(multiplier float64) bool {
	return multiplier < 1.0
}

// ValidatePlaceBetRequest validates a place bet request
func (req *PlaceBetRequest) Validate() error {
	if len(req.BallBets) == 0 {
		return errors.New("no balls selected for betting")
	}

	if len(req.BallBets) > 4 {
		return errors.New("maximum 4 balls can be selected")
	}

	availableBalls := GetAvailableBalls()
	validBallIDs := make(map[int]bool)
	for _, ball := range availableBalls {
		validBallIDs[ball.ID] = true
	}

	totalAmount := 0.0
	for ballID, amount := range req.BallBets {
		if !validBallIDs[ballID] {
			return fmt.Errorf("invalid ball ID: %d", ballID)
		}

		if amount <= 0 {
			return fmt.Errorf("bet amount must be positive for ball %d", ballID)
		}

		if amount > 1000 {
			return fmt.Errorf("maximum bet amount per ball is $1000, got $%.2f for ball %d", amount, ballID)
		}

		totalAmount += amount
	}

	if totalAmount < 10 {
		return errors.New("minimum total bet amount is $10")
	}

	if totalAmount > 5000 {
		return errors.New("maximum total bet amount is $5000")
	}

	return nil
}

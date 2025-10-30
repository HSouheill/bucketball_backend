package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameService struct {
	gameRepo    *repositories.GameRepository
	userRepo    *repositories.UserRepository
	houseWallet *models.HouseWallet
}

// NewGameService creates a new game service
func NewGameService(gameRepo *repositories.GameRepository, userRepo *repositories.UserRepository) *GameService {
	return &GameService{
		gameRepo: gameRepo,
		userRepo: userRepo,
	}
}

// GetGameState gets the current game state
func (s *GameService) GetGameState(ctx context.Context, userID primitive.ObjectID) (*models.GameState, error) {
	// Get current game
	currentGame, err := s.gameRepo.GetCurrentGame(ctx)
	if err != nil {
		return nil, err
	}

	// Get house wallet
	houseWallet, err := s.gameRepo.GetHouseWallet(ctx)
	if err != nil {
		return nil, err
	}
	s.houseWallet = houseWallet

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user's recent game history
	gameHistory, err := s.gameRepo.GetGameResultsByUserID(ctx, userID, 10)
	if err != nil {
		return nil, err
	}

	gameState := &models.GameState{
		CurrentGame:      currentGame,
		AvailableBalls:   models.GetAvailableBalls(),
		AvailableBaskets: models.GetAvailableBaskets(),
		UserBalance:      user.Balance,
		HouseWallet:      houseWallet.Balance,
		AdminProfit:      houseWallet.AdminProfit,
		TotalBets:        houseWallet.TotalBets,
		GameHistory:      gameHistory,
	}

	return gameState, nil
}

// PlaceBet places a bet for a user
func (s *GameService) PlaceBet(ctx context.Context, userID primitive.ObjectID, req *models.PlaceBetRequest) (*models.Game, error) {
	// Validate bet request
	if len(req.BallBets) == 0 {
		return nil, errors.New("no balls selected for betting")
	}

	// Validate ball IDs
	availableBalls := models.GetAvailableBalls()
	validBallIDs := make(map[int]bool)
	for _, ball := range availableBalls {
		validBallIDs[ball.ID] = true
	}

	totalBetAmount := 0.0
	for ballID, amount := range req.BallBets {
		// Validate ball ID
		if !validBallIDs[ballID] {
			return nil, fmt.Errorf("invalid ball ID: %d", ballID)
		}

		// Validate bet amount
		if amount <= 0 {
			return nil, fmt.Errorf("invalid bet amount for ball %d", ballID)
		}

		// Validate maximum bet amount per ball
		if amount > 1000 {
			return nil, fmt.Errorf("maximum bet amount per ball is $1000, got $%.2f for ball %d", amount, ballID)
		}

		totalBetAmount += amount
	}

	// Validate total bet amount
	if totalBetAmount > 5000 {
		return nil, errors.New("maximum total bet amount is $5000")
	}

	// Validate minimum bet amount
	if totalBetAmount < 10 {
		return nil, errors.New("minimum total bet amount is $10")
	}

	// Get user and check balance
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Balance < totalBetAmount {
		return nil, errors.New("insufficient balance")
	}

	// Get or create current game
	currentGame, err := s.gameRepo.GetCurrentGame(ctx)
	if err != nil {
		return nil, err
	}

	if currentGame == nil {
		// Create new game
		currentGame = &models.Game{
			ID:          primitive.NewObjectID(),
			RoundNumber: 1,
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Get house wallet for the game
		houseWallet, err := s.gameRepo.GetHouseWallet(ctx)
		if err != nil {
			return nil, err
		}

		currentGame.HouseWallet = houseWallet.Balance
		currentGame.AdminProfit = houseWallet.AdminProfit
		currentGame.TotalBets = houseWallet.TotalBets

		if err := s.gameRepo.CreateGame(ctx, currentGame); err != nil {
			return nil, err
		}
	}

	// Check if house wallet has enough funds
	if currentGame.HouseWallet <= 0 {
		return nil, errors.New("house wallet is empty")
	}

	// Create bets
	for ballID, amount := range req.BallBets {
		bet := &models.Bet{
			ID:     primitive.NewObjectID(),
			UserID: userID,
			GameID: currentGame.ID,
			BallID: ballID,
			Amount: amount,
			Status: "pending",
		}

		if err := s.gameRepo.CreateBet(ctx, bet); err != nil {
			return nil, err
		}
	}

	// Deduct bet amount from user balance
	if err := user.SubtractFromBalance(totalBetAmount); err != nil {
		return nil, err
	}

	// Update user in database
	updateData := map[string]interface{}{
		"balance":    user.Balance,
		"updated_at": time.Now(),
	}
	if err := s.userRepo.Update(ctx, userID, updateData); err != nil {
		return nil, err
	}

	// Update game total bets
	updateGameData := map[string]interface{}{
		"total_bets": currentGame.TotalBets + totalBetAmount,
		"updated_at": time.Now(),
	}
	if err := s.gameRepo.UpdateGame(ctx, currentGame.ID, updateGameData); err != nil {
		return nil, err
	}

	return currentGame, nil
}

// PlayGame executes the game and determines results
func (s *GameService) PlayGame(ctx context.Context, gameID primitive.ObjectID) error {
	// Get game
	game, err := s.gameRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("game is not active")
	}

	// Check if game has been active for too long (prevent stale games)
	if time.Since(game.CreatedAt) > 30*time.Minute {
		// Mark game as completed due to timeout
		updateGameData := map[string]interface{}{
			"status":       "completed",
			"completed_at": time.Now(),
			"updated_at":   time.Now(),
		}
		s.gameRepo.UpdateGame(ctx, gameID, updateGameData)
		return errors.New("game has expired due to inactivity")
	}

	// Get all bets for this game
	bets, err := s.gameRepo.GetBetsByGameID(ctx, gameID)
	if err != nil {
		return err
	}

	if len(bets) == 0 {
		return errors.New("no bets found for this game")
	}

	// Group bets by ball
	ballBets := make(map[int][]models.Bet)
	for _, bet := range bets {
		ballBets[bet.BallID] = append(ballBets[bet.BallID], bet)
	}

	// Calculate winning baskets for each ball
	ballTargets := make(map[int]int)
	playerBets := make(map[int]float64)

	for ballID, ballBetList := range ballBets {
		// Calculate total player bets for this ball
		totalPlayerBets := 0.0
		for _, bet := range ballBetList {
			totalPlayerBets += bet.Amount
		}
		playerBets[ballID] = totalPlayerBets

		// Calculate winning basket for this ball
		ballTargets[ballID] = models.CalculateWinningBasket(map[int]float64{ballID: totalPlayerBets}, game.HouseWallet)
	}

	// Select winning ball randomly from balls with bets
	ballIDs := make([]int, 0, len(ballBets))
	for ballID := range ballBets {
		ballIDs = append(ballIDs, ballID)
	}

	rand.Seed(time.Now().UnixNano())
	winningBallID := ballIDs[rand.Intn(len(ballIDs))]
	winningBasketID := ballTargets[winningBallID]

	// Update game with results
	now := time.Now()
	updateGameData := map[string]interface{}{
		"status":            "completed",
		"winning_ball_id":   winningBallID,
		"winning_basket_id": winningBasketID,
		"completed_at":      now,
		"updated_at":        now,
	}

	if err := s.gameRepo.UpdateGame(ctx, gameID, updateGameData); err != nil {
		return err
	}

	// Process results for each bet
	availableBaskets := models.GetAvailableBaskets()
	maxAllowedWin := game.HouseWallet * 0.20
	totalWins := 0.0
	var results []models.GameResult

	// First pass: calculate all wins to check wallet limits
	tempResults := make([]models.GameResult, 0)
	for ballID, ballBetList := range ballBets {
		basketIndex := ballTargets[ballID]
		multiplier := availableBaskets[basketIndex].Value
		ball := models.GetAvailableBalls()[ballID]

		for _, bet := range ballBetList {
			winAmount := bet.CalculateWinAmount(multiplier)
			profit := bet.CalculateProfit(multiplier)

			if bet.IsWin(multiplier) && profit > 0 {
				totalWins += profit
			}

			result := models.GameResult{
				ID:           primitive.NewObjectID(),
				UserID:       bet.UserID,
				GameID:       gameID,
				BallID:       ballID,
				BallName:     ball.Name,
				BallColor:    ball.Color,
				BetAmount:    bet.Amount,
				Multiplier:   multiplier,
				WinAmount:    winAmount,
				Profit:       profit,
				BasketLanded: basketIndex,
				Won:          bet.IsWin(multiplier),
				Pushed:       bet.IsPush(multiplier),
			}

			tempResults = append(tempResults, result)
		}
	}

	// Apply wallet limit if necessary
	walletLimitFactor := 1.0
	if totalWins > maxAllowedWin {
		walletLimitFactor = maxAllowedWin / totalWins
	}

	// Second pass: apply wallet limits and update user balances
	for _, result := range tempResults {
		finalProfit := result.Profit
		finalWinAmount := result.WinAmount

		if result.Won && result.Profit > 0 {
			finalProfit = result.Profit * walletLimitFactor
			finalWinAmount = result.BetAmount + finalProfit
			result.WalletLimited = walletLimitFactor < 1.0
		} else if result.Pushed {
			finalProfit = 0
			finalWinAmount = result.BetAmount
		}

		result.Profit = finalProfit
		result.WinAmount = finalWinAmount
		result.Multiplier = finalWinAmount / result.BetAmount

		// Update user balance
		user, err := s.userRepo.GetByID(ctx, result.UserID)
		if err == nil {
			user.AddToBalance(finalProfit)
			updateUserData := map[string]interface{}{
				"balance":    user.Balance,
				"updated_at": time.Now(),
			}
			s.userRepo.Update(ctx, result.UserID, updateUserData)
		}

		// Update bet status
		betStatus := "lost"
		if result.Won {
			betStatus = "won"
		} else if result.Pushed {
			betStatus = "pushed"
		}

		updateBetData := map[string]interface{}{
			"status":     betStatus,
			"updated_at": time.Now(),
		}
		s.gameRepo.UpdateBet(ctx, primitive.NewObjectID(), updateBetData)

		// Create game result
		if err := s.gameRepo.CreateGameResult(ctx, &result); err != nil {
			// Log error but continue processing
			fmt.Printf("Error creating game result: %v\n", err)
		}

		results = append(results, result)
	}

	// Update house wallet
	houseWallet, err := s.gameRepo.GetHouseWallet(ctx)
	if err == nil {
		netHouseChange := 0.0
		for _, result := range results {
			if result.Profit > 0 {
				netHouseChange -= result.Profit
			} else {
				netHouseChange += math.Abs(result.Profit)
			}
		}

		// Add admin profit (2-4% of total bets)
		adminProfitRate := 0.02 + rand.Float64()*0.02
		adminProfit := game.TotalBets * adminProfitRate
		netHouseChange -= adminProfit

		updateWalletData := map[string]interface{}{
			"balance":      houseWallet.Balance + netHouseChange,
			"admin_profit": houseWallet.AdminProfit + adminProfit,
			"total_bets":   houseWallet.TotalBets + game.TotalBets,
		}
		s.gameRepo.UpdateHouseWallet(ctx, updateWalletData)
	}

	return nil
}

// GetGameHistory gets game history for a user
func (s *GameService) GetGameHistory(ctx context.Context, userID primitive.ObjectID, limit int64) ([]models.GameResult, error) {
	return s.gameRepo.GetGameResultsByUserID(ctx, userID, limit)
}

// GetGameStats gets overall game statistics
func (s *GameService) GetGameStats(ctx context.Context) (map[string]interface{}, error) {
	return s.gameRepo.GetGameStats(ctx)
}

// GetUserGameStats gets game statistics for a specific user
func (s *GameService) GetUserGameStats(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	return s.gameRepo.GetUserGameStats(ctx, userID)
}

// SimulateOtherPlayers simulates other players placing bets (for testing)
func (s *GameService) SimulateOtherPlayers(ctx context.Context, gameID primitive.ObjectID, numPlayers int) error {
	availableBalls := models.GetAvailableBalls()
	betAmounts := []float64{50, 100, 200}

	for i := 0; i < numPlayers; i++ {
		// Create a temporary user for simulation
		simUserID := primitive.NewObjectID()

		// Random ball selection
		ballID := availableBalls[rand.Intn(len(availableBalls))].ID
		amount := betAmounts[rand.Intn(len(betAmounts))]

		bet := &models.Bet{
			ID:     primitive.NewObjectID(),
			UserID: simUserID,
			GameID: gameID,
			BallID: ballID,
			Amount: amount,
			Status: "pending",
		}

		if err := s.gameRepo.CreateBet(ctx, bet); err != nil {
			return err
		}
	}

	return nil
}

// GetHouseWallet gets the current house wallet state
func (s *GameService) GetHouseWallet(ctx context.Context) (*models.HouseWallet, error) {
	return s.gameRepo.GetHouseWallet(ctx)
}

package repositories

import (
	"context"
	"time"

	"github.com/HSouheil/bucketball_backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GameRepository struct {
	db *mongo.Database
}

// NewGameRepository creates a new game repository
func NewGameRepository(db *mongo.Database) *GameRepository {
	return &GameRepository{
		db: db,
	}
}

// CreateGame creates a new game
func (r *GameRepository) CreateGame(ctx context.Context, game *models.Game) error {
	collection := r.db.Collection("games")
	game.CreatedAt = time.Now()
	game.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, game)
	return err
}

// GetGameByID gets a game by ID
func (r *GameRepository) GetGameByID(ctx context.Context, gameID primitive.ObjectID) (*models.Game, error) {
	collection := r.db.Collection("games")

	var game models.Game
	err := collection.FindOne(ctx, bson.M{"_id": gameID}).Decode(&game)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

// GetCurrentGame gets the current active game
func (r *GameRepository) GetCurrentGame(ctx context.Context) (*models.Game, error) {
	collection := r.db.Collection("games")

	var game models.Game
	err := collection.FindOne(ctx, bson.M{"status": "active"}, options.FindOne().SetSort(bson.M{"created_at": -1})).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &game, nil
}

// UpdateGame updates a game
func (r *GameRepository) UpdateGame(ctx context.Context, gameID primitive.ObjectID, updateData map[string]interface{}) error {
	collection := r.db.Collection("games")
	updateData["updated_at"] = time.Now()

	_, err := collection.UpdateOne(ctx, bson.M{"_id": gameID}, bson.M{"$set": updateData})
	return err
}

// CreateBet creates a new bet
func (r *GameRepository) CreateBet(ctx context.Context, bet *models.Bet) error {
	collection := r.db.Collection("bets")
	bet.CreatedAt = time.Now()
	bet.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, bet)
	return err
}

// GetBetsByGameID gets all bets for a specific game
func (r *GameRepository) GetBetsByGameID(ctx context.Context, gameID primitive.ObjectID) ([]models.Bet, error) {
	collection := r.db.Collection("bets")

	cursor, err := collection.Find(ctx, bson.M{"game_id": gameID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bets []models.Bet
	if err = cursor.All(ctx, &bets); err != nil {
		return nil, err
	}

	return bets, nil
}

// GetBetsByUserID gets all bets for a specific user
func (r *GameRepository) GetBetsByUserID(ctx context.Context, userID primitive.ObjectID, limit int64) ([]models.Bet, error) {
	collection := r.db.Collection("bets")

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bets []models.Bet
	if err = cursor.All(ctx, &bets); err != nil {
		return nil, err
	}

	return bets, nil
}

// UpdateBet updates a bet
func (r *GameRepository) UpdateBet(ctx context.Context, betID primitive.ObjectID, updateData map[string]interface{}) error {
	collection := r.db.Collection("bets")
	updateData["updated_at"] = time.Now()

	_, err := collection.UpdateOne(ctx, bson.M{"_id": betID}, bson.M{"$set": updateData})
	return err
}

// CreateGameResult creates a new game result
func (r *GameRepository) CreateGameResult(ctx context.Context, result *models.GameResult) error {
	collection := r.db.Collection("game_results")
	result.CreatedAt = time.Now()

	_, err := collection.InsertOne(ctx, result)
	return err
}

// GetGameResultsByUserID gets game results for a specific user
func (r *GameRepository) GetGameResultsByUserID(ctx context.Context, userID primitive.ObjectID, limit int64) ([]models.GameResult, error) {
	collection := r.db.Collection("game_results")

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.GameResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetGameResultsByGameID gets all game results for a specific game
func (r *GameRepository) GetGameResultsByGameID(ctx context.Context, gameID primitive.ObjectID) ([]models.GameResult, error) {
	collection := r.db.Collection("game_results")

	cursor, err := collection.Find(ctx, bson.M{"game_id": gameID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.GameResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetHouseWallet gets the current house wallet state
func (r *GameRepository) GetHouseWallet(ctx context.Context) (*models.HouseWallet, error) {
	collection := r.db.Collection("house_wallet")

	var wallet models.HouseWallet
	err := collection.FindOne(ctx, bson.M{}).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create initial house wallet if it doesn't exist
			wallet = models.HouseWallet{
				ID:          primitive.NewObjectID(),
				Balance:     1000.0, // Initial house wallet
				AdminProfit: 0.0,
				TotalBets:   0.0,
				UpdatedAt:   time.Now(),
			}
			_, err = collection.InsertOne(ctx, wallet)
			if err != nil {
				return nil, err
			}
			return &wallet, nil
		}
		return nil, err
	}

	return &wallet, nil
}

// UpdateHouseWallet updates the house wallet
func (r *GameRepository) UpdateHouseWallet(ctx context.Context, updateData map[string]interface{}) error {
	collection := r.db.Collection("house_wallet")
	updateData["updated_at"] = time.Now()

	_, err := collection.UpdateOne(ctx, bson.M{}, bson.M{"$set": updateData})
	return err
}

// GetGameStats gets game statistics
func (r *GameRepository) GetGameStats(ctx context.Context) (map[string]interface{}, error) {
	collection := r.db.Collection("games")

	// Get total games count
	totalGames, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Get completed games count
	completedGames, err := collection.CountDocuments(ctx, bson.M{"status": "completed"})
	if err != nil {
		return nil, err
	}

	// Get active games count
	activeGames, err := collection.CountDocuments(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, err
	}

	// Get total bets amount
	betsCollection := r.db.Collection("bets")
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$amount"},
			},
		},
	}

	cursor, err := betsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	totalBetsAmount := 0.0
	if len(result) > 0 {
		if total, ok := result[0]["total"].(float64); ok {
			totalBetsAmount = total
		}
	}

	return map[string]interface{}{
		"total_games":       totalGames,
		"completed_games":   completedGames,
		"active_games":      activeGames,
		"total_bets_amount": totalBetsAmount,
	}, nil
}

// GetUserGameStats gets game statistics for a specific user
func (r *GameRepository) GetUserGameStats(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	betsCollection := r.db.Collection("bets")
	resultsCollection := r.db.Collection("game_results")

	// Get user's total bets
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{
			"$group": bson.M{
				"_id":        nil,
				"total_bets": bson.M{"$sum": "$amount"},
				"bet_count":  bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := betsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var betResult []bson.M
	if err = cursor.All(ctx, &betResult); err != nil {
		return nil, err
	}

	totalBets := 0.0
	betCount := 0
	if len(betResult) > 0 {
		if total, ok := betResult[0]["total_bets"].(float64); ok {
			totalBets = total
		}
		if count, ok := betResult[0]["bet_count"].(int32); ok {
			betCount = int(count)
		}
	}

	// Get user's total winnings
	winPipeline := []bson.M{
		{"$match": bson.M{"user_id": userID, "won": true}},
		{
			"$group": bson.M{
				"_id":        nil,
				"total_wins": bson.M{"$sum": "$profit"},
				"win_count":  bson.M{"$sum": 1},
			},
		},
	}

	winCursor, err := resultsCollection.Aggregate(ctx, winPipeline)
	if err != nil {
		return nil, err
	}
	defer winCursor.Close(ctx)

	var winResult []bson.M
	if err = winCursor.All(ctx, &winResult); err != nil {
		return nil, err
	}

	totalWins := 0.0
	winCount := 0
	if len(winResult) > 0 {
		if wins, ok := winResult[0]["total_wins"].(float64); ok {
			totalWins = wins
		}
		if count, ok := winResult[0]["win_count"].(int32); ok {
			winCount = int(count)
		}
	}

	// Calculate win rate
	winRate := 0.0
	if betCount > 0 {
		winRate = float64(winCount) / float64(betCount) * 100
	}

	return map[string]interface{}{
		"total_bets": totalBets,
		"bet_count":  betCount,
		"total_wins": totalWins,
		"win_count":  winCount,
		"win_rate":   winRate,
		"net_profit": totalWins - totalBets,
	}, nil
}

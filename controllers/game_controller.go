package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/HSouheil/bucketball_backend/models"
	"github.com/HSouheil/bucketball_backend/services"
	"github.com/HSouheil/bucketball_backend/utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameController struct {
	gameService *services.GameService
}

// NewGameController creates a new game controller
func NewGameController(gameService *services.GameService) *GameController {
	return &GameController{
		gameService: gameService,
	}
}

// GetGameState gets the current game state
func (gc *GameController) GetGameState(c echo.Context) error {
	// Get user ID from JWT token
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid token")
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID")
	}

	ctx := c.Request().Context()
	gameState, err := gc.gameService.GetGameState(ctx, objectID)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get game state", err)
	}

	return utils.SuccessResponse(c, "Game state retrieved successfully", gameState)
}

// PlaceBet places a bet for the current user
func (gc *GameController) PlaceBet(c echo.Context) error {
	// Get user ID from JWT token
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid token")
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID")
	}

	var req models.PlaceBetRequest
	if err := c.Bind(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request data", err)
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Validation failed", err)
	}

	// Additional validation using the model's validation method
	if err := req.Validate(); err != nil {
		return utils.ValidationErrorResponse(c, "Bet validation failed", err)
	}

	ctx := c.Request().Context()
	game, err := gc.gameService.PlaceBet(ctx, objectID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient balance") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "no balls selected") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "invalid bet amount") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "house wallet is empty") {
			return utils.ErrorResponse(c, http.StatusServiceUnavailable, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to place bet", err)
	}

	return utils.SuccessResponse(c, "Bet placed successfully", game)
}

// PlayGame executes the current game
func (gc *GameController) PlayGame(c echo.Context) error {
	// Get user ID from JWT token
	_, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid token")
	}

	// Get game ID from URL parameter
	gameID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(gameID)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid game ID")
	}

	ctx := c.Request().Context()
	if err := gc.gameService.PlayGame(ctx, objectID); err != nil {
		if strings.Contains(err.Error(), "game is not active") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "no bets found") {
			return utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		}
		return utils.InternalServerErrorResponse(c, "Failed to play game", err)
	}

	return utils.SuccessResponse(c, "Game played successfully", nil)
}

// GetGameHistory gets game history for the current user
func (gc *GameController) GetGameHistory(c echo.Context) error {
	// Get user ID from JWT token
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid token")
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID")
	}

	// Parse limit parameter
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)
	if limit <= 0 {
		limit = 10
	}

	ctx := c.Request().Context()
	history, err := gc.gameService.GetGameHistory(ctx, objectID, limit)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get game history", err)
	}

	return utils.SuccessResponse(c, "Game history retrieved successfully", history)
}

// GetGameStats gets overall game statistics
func (gc *GameController) GetGameStats(c echo.Context) error {
	ctx := c.Request().Context()
	stats, err := gc.gameService.GetGameStats(ctx)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get game statistics", err)
	}

	return utils.SuccessResponse(c, "Game statistics retrieved successfully", stats)
}

// GetUserGameStats gets game statistics for the current user
func (gc *GameController) GetUserGameStats(c echo.Context) error {
	// Get user ID from JWT token
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid token")
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID")
	}

	ctx := c.Request().Context()
	stats, err := gc.gameService.GetUserGameStats(ctx, objectID)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user game statistics", err)
	}

	return utils.SuccessResponse(c, "User game statistics retrieved successfully", stats)
}

// SimulateOtherPlayers simulates other players (admin only)
func (gc *GameController) SimulateOtherPlayers(c echo.Context) error {
	// Check if user is admin
	role, err := utils.GetUserRoleFromToken(c)
	if err != nil || role != "admin" {
		return utils.ForbiddenResponse(c, "Admin access required")
	}

	// Get game ID from URL parameter
	gameID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(gameID)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid game ID")
	}

	// Parse number of players parameter
	numPlayers, _ := strconv.Atoi(c.QueryParam("players"))
	if numPlayers <= 0 {
		numPlayers = 2
	}
	if numPlayers > 10 {
		numPlayers = 10
	}

	ctx := c.Request().Context()
	if err := gc.gameService.SimulateOtherPlayers(ctx, objectID, numPlayers); err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to simulate other players", err)
	}

	return utils.SuccessResponse(c, "Other players simulated successfully", map[string]interface{}{
		"players_added": numPlayers,
	})
}

// GetAvailableBalls gets available balls for betting
func (gc *GameController) GetAvailableBalls(c echo.Context) error {
	balls := models.GetAvailableBalls()
	return utils.SuccessResponse(c, "Available balls retrieved successfully", balls)
}

// GetAvailableBaskets gets available baskets with multipliers
func (gc *GameController) GetAvailableBaskets(c echo.Context) error {
	baskets := models.GetAvailableBaskets()
	return utils.SuccessResponse(c, "Available baskets retrieved successfully", baskets)
}

// GetHouseWallet gets current house wallet state (admin only)
func (gc *GameController) GetHouseWallet(c echo.Context) error {
	// Check if user is admin
	role, err := utils.GetUserRoleFromToken(c)
	if err != nil || role != "admin" {
		return utils.ForbiddenResponse(c, "Admin access required")
	}

	ctx := c.Request().Context()
	houseWallet, err := gc.gameService.GetHouseWallet(ctx)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get house wallet", err)
	}

	return utils.SuccessResponse(c, "House wallet retrieved successfully", houseWallet)
}

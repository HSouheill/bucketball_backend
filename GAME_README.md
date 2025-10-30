# Bucket Ball Betting Game Implementation

This document describes the implementation of the Bucket Ball betting game in Go, translated from the original JavaScript React implementation.

## Overview

The Bucket Ball game is a betting game where players can bet on colored balls that fall into baskets with different multipliers. The game includes:

- 4 colored balls (Red, Cyan, Yellow, Green)
- 8 baskets with multipliers ranging from 0.25x to 10x
- House wallet management
- Admin profit tracking
- User balance management
- Game history and statistics

## Game Rules

- **Win**: Ball lands in 2x+ basket (multiplier ≥ 2)
- **Push**: Ball lands in 1x basket (multiplier = 1) - break even
- **Loss**: Ball lands in 0.25x-0.75x basket (multiplier < 1)
- **Multiple balls can win in the same round**
- **Maximum win**: 20% of house wallet per game
- **House pays winners from wallet, collects from losers**

## API Endpoints

### Game State & Information
- `GET /api/games/state` - Get current game state
- `GET /api/games/balls` - Get available balls
- `GET /api/games/baskets` - Get available baskets

### Betting
- `POST /api/games/bet` - Place a bet
- `POST /api/games/:id/play` - Execute the game

### History & Statistics
- `GET /api/games/history` - Get user's game history
- `GET /api/games/stats` - Get user's game statistics

### Admin Endpoints
- `GET /api/admin/games/stats` - Get overall game statistics
- `GET /api/admin/games/house-wallet` - Get house wallet state
- `POST /api/admin/games/:id/simulate` - Simulate other players

## Request/Response Examples

### Place Bet Request
```json
{
  "ball_bets": {
    "0": 100.0,
    "1": 50.0
  }
}
```

### Game State Response
```json
{
  "success": true,
  "message": "Game state retrieved successfully",
  "data": {
    "current_game": {
      "id": "64f8a1b2c3d4e5f6a7b8c9d0",
      "round_number": 1,
      "status": "active",
      "total_bets": 150.0,
      "house_wallet": 1000.0,
      "admin_profit": 0.0
    },
    "available_balls": [
      {"id": 0, "color": "#FF6B6B", "name": "Red"},
      {"id": 1, "color": "#4ECDC4", "name": "Cyan"},
      {"id": 2, "color": "#FFE66D", "name": "Yellow"},
      {"id": 3, "color": "#95E1D3", "name": "Green"}
    ],
    "available_baskets": [
      {"value": 0.25, "color": "#e74c3c"},
      {"value": 0.50, "color": "#e67e22"},
      {"value": 0.75, "color": "#f39c12"},
      {"value": 1, "color": "#f1c40f"},
      {"value": 2, "color": "#2ecc71"},
      {"value": 4, "color": "#3498db"},
      {"value": 8, "color": "#9b59b6"},
      {"value": 10, "color": "#1abc9c"}
    ],
    "user_balance": 500.0,
    "house_wallet": 1000.0,
    "admin_profit": 0.0,
    "total_bets": 0.0,
    "game_history": []
  }
}
```

## Database Collections

### games
- `_id`: Game ID
- `round_number`: Round number
- `status`: Game status (pending, active, completed)
- `winning_ball_id`: ID of winning ball
- `winning_basket_id`: ID of winning basket
- `total_bets`: Total amount bet in this game
- `house_wallet`: House wallet at game start
- `admin_profit`: Admin profit from this game
- `created_at`: Game creation timestamp
- `updated_at`: Last update timestamp
- `completed_at`: Game completion timestamp

### bets
- `_id`: Bet ID
- `user_id`: User who placed the bet
- `game_id`: Game this bet belongs to
- `ball_id`: Ball being bet on
- `amount`: Bet amount
- `status`: Bet status (pending, won, lost, pushed)
- `created_at`: Bet creation timestamp
- `updated_at`: Last update timestamp

### game_results
- `_id`: Result ID
- `user_id`: User this result belongs to
- `game_id`: Game this result belongs to
- `ball_id`: Ball that was bet on
- `ball_name`: Name of the ball
- `ball_color`: Color of the ball
- `bet_amount`: Original bet amount
- `multiplier`: Final multiplier applied
- `win_amount`: Amount won
- `profit`: Profit/loss amount
- `basket_landed`: Basket index where ball landed
- `won`: Whether this was a win
- `pushed`: Whether this was a push
- `wallet_limited`: Whether win was limited by house wallet
- `created_at`: Result creation timestamp

### house_wallet
- `_id`: Wallet ID
- `balance`: Current house wallet balance
- `admin_profit`: Total admin profit
- `total_bets`: Total amount bet across all games
- `updated_at`: Last update timestamp

## Game Logic Implementation

### Basket Selection Algorithm
The game uses a weighted random selection for basket landing:

1. Calculate maximum allowed win (20% of house wallet)
2. Calculate maximum multiplier based on total bets
3. Adjust basket weights based on wallet capacity
4. Use weighted random selection to choose basket

### Win Calculation
- **Win**: Multiplier ≥ 2.0
- **Push**: Multiplier = 1.0
- **Loss**: Multiplier < 1.0

### House Wallet Management
- House wallet starts at $1000
- Admin takes 2-4% of total bets as profit
- Winners are paid from house wallet
- Losers' money goes to house wallet
- Maximum win per game is 20% of house wallet

## Security Features

### Input Validation
- Ball ID validation (must be 0-3)
- Bet amount validation ($10-$5000 total, $1-$1000 per ball)
- Maximum 4 balls per bet
- Minimum bet amount validation

### Rate Limiting
- Game endpoints: 50 requests per hour per user
- Admin endpoints: 200 requests per hour per admin

### Authentication
- JWT token required for all game endpoints
- Admin role required for admin endpoints
- Token validation on every request

### Game Security
- Game timeout after 30 minutes of inactivity
- House wallet balance checks
- User balance validation before betting
- Atomic bet processing

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Business logic conflict (e.g., insufficient balance)
- `500 Internal Server Error`: Server error

## Usage Examples

### 1. Get Game State
```bash
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/games/state
```

### 2. Place a Bet
```bash
curl -X POST \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"ball_bets": {"0": 100, "1": 50}}' \
     http://localhost:8080/api/games/bet
```

### 3. Play Game
```bash
curl -X POST \
     -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/games/{game_id}/play
```

### 4. Get Game History
```bash
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8080/api/games/history?limit=10"
```

## Testing

The implementation includes simulation capabilities for testing:

- `SimulateOtherPlayers()`: Add simulated players to a game
- Admin endpoint to trigger simulations
- Configurable number of simulated players (1-10)

## Future Enhancements

Potential improvements for the game system:

1. **Real-time Updates**: WebSocket support for live game updates
2. **Tournament Mode**: Multi-round tournaments
3. **Achievement System**: User achievements and badges
4. **Leaderboards**: Top players and statistics
5. **Mobile App**: Native mobile application
6. **Advanced Analytics**: Detailed game analytics and reporting
7. **Custom Games**: User-created game variations
8. **Social Features**: Friend lists and social interactions

## Configuration

The game can be configured through environment variables or config files:

- House wallet initial balance
- Bet limits (min/max per ball, total)
- Game timeout duration
- Admin profit percentage range
- Rate limiting settings

## Monitoring

Key metrics to monitor:

- House wallet balance
- Admin profit trends
- User betting patterns
- Game completion rates
- Error rates and types
- API response times

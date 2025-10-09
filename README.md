# BucketBall Backend

A secure and professional backend API for a basketball game management system built with Go, Echo framework, MongoDB, and Redis.

## Features

- **Authentication & Authorization**: JWT-based authentication with role-based access control
- **User Management**: Complete CRUD operations for user management
- **Security**: Password hashing, rate limiting, CORS protection
- **Database**: MongoDB for data persistence with Redis for caching and session management
- **Docker**: Fully containerized application with Docker Compose
- **API Documentation**: RESTful API with proper HTTP status codes and error handling

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Echo v4
- **Database**: MongoDB
- **Cache**: Redis
- **Authentication**: JWT
- **Containerization**: Docker & Docker Compose

## Project Structure

```
bucketball_backend/
├── cmd/
│   └── main.go                 # Application entry point
├── controllers/
│   ├── auth_controller.go      # Authentication endpoints
│   └── user_controller.go      # User management endpoints
├── middlewares/
│   ├── auth.go                 # JWT authentication middleware
│   └── rate_limit.go           # Rate limiting middleware
├── models/
│   ├── user.go                 # User data models
│   └── game.go                 # Game data models
├── repositories/
│   ├── mongodb.go              # MongoDB connection
│   ├── redis.go                # Redis connection
│   ├── user_repository.go      # User data access layer
│   └── auth_repository.go      # Authentication data access layer
├── routes/
│   └── routes.go               # API route definitions
├── security/
│   ├── jwt.go                  # JWT token management
│   └── password.go             # Password hashing utilities
├── utils/
│   ├── env.go                  # Environment variable utilities
│   ├── response.go             # API response utilities
│   └── validation.go           # Input validation utilities
├── Dockerfile                  # Docker configuration
├── docker-compose.yml          # Multi-container setup
├── env.example                 # Environment variables template
└── go.mod                      # Go module dependencies
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd bucketball_backend
   ```

2. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Check if services are running**
   ```bash
   docker-compose ps
   ```

5. **View logs**
   ```bash
   docker-compose logs -f backend
   ```

### Local Development

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Start MongoDB and Redis**
   ```bash
   # Using Docker
   docker run -d -p 27017:27017 --name mongodb mongo:7.0
   docker run -d -p 6379:6379 --name redis redis:7.2-alpine
   ```

4. **Run the application**
   ```bash
   go run cmd/main.go
   ```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

### User Management
- `GET /api/v1/users/profile` - Get current user profile
- `PUT /api/v1/users/profile` - Update current user profile

### Admin Endpoints
- `GET /api/v1/admin/users` - Get all users (paginated)
- `GET /api/v1/admin/users/:id` - Get user by ID
- `PUT /api/v1/admin/users/:id` - Update user
- `DELETE /api/v1/admin/users/:id` - Delete user
- `PATCH /api/v1/admin/users/:id/toggle-status` - Toggle user status

### Health Check
- `GET /health` - Health check endpoint

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGODB_DB` | MongoDB database name | `bucketball` |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_DB` | Redis database number | `0` |
| `JWT_SECRET` | JWT signing secret | (required) |
| `ENV` | Environment | `development` |

## Security Features

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: bcrypt for password security
- **Rate Limiting**: Prevents abuse with configurable limits
- **CORS Protection**: Cross-origin resource sharing protection
- **Input Validation**: Comprehensive request validation
- **Token Blacklisting**: Secure logout with token invalidation
- **Role-based Access**: Admin and user role separation

## Development

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o main cmd/main.go
```

### Code Formatting
```bash
go fmt ./...
```

### Linting
```bash
golangci-lint run
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

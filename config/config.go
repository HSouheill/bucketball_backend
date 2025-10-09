package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	App      AppConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string
	Database string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// AppConfig holds general app configuration
type AppConfig struct {
	Environment string
	Name        string
	Version     string
}

var cfg *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	if cfg != nil {
		return cfg
	}

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg = &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DB", "bucketball"),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret: requiredEnv("JWT_SECRET"),
		},
		App: AppConfig{
			Environment: getEnv("ENV", "development"),
			Name:        getEnv("APP_NAME", "BucketBall Backend"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
		},
	}

	return cfg
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	if cfg == nil {
		return LoadConfig()
	}
	return cfg
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// requiredEnv gets a required environment variable or panics if not found
func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}


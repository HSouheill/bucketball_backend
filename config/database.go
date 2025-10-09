package config

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	redisClient *redis.Client
)

// InitMongoDB initializes MongoDB connection
func InitMongoDB(cfg *Config) *mongo.Client {
	if mongoClient != nil {
		return mongoClient
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	mongoClient = client
	log.Printf("Connected to MongoDB database: %s", cfg.MongoDB.Database)
	return mongoClient
}

// GetMongoDB returns the MongoDB client
func GetMongoDB() *mongo.Client {
	return mongoClient
}

// GetMongoDatabase returns the database instance
func GetMongoDatabase(cfg *Config) *mongo.Database {
	return mongoClient.Database(cfg.MongoDB.Database)
}

// InitRedis initializes Redis connection
func InitRedis(cfg *Config) *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	redisClient = client
	log.Printf("Connected to Redis at %s", cfg.Redis.Address)
	return redisClient
}

// GetRedis returns the Redis client
func GetRedis() *redis.Client {
	return redisClient
}

// CloseDatabases closes all database connections
func CloseDatabases() {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}
}


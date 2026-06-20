package main

import (
	"database/sql"
	"log"
	"os"
	"wallet-service/pkg/database"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.Println("Starting Wallet Service...")

	//  Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: No .env file found, reading from system environment")
	}

	// Read the variables using standard os.Getenv
	postgresDSN := os.Getenv("DB_DSN")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Quick validation to make sure we aren't passing empty strings
	if postgresDSN == "" || redisAddr == "" {
		log.Fatalf("❌ Configuration error: DB_DSN or REDIS_ADDR is missing in environment")
	}

	// Connect to PostgreSQL
	db, err := database.SetupPostgres(postgresDSN)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	// Connect to Redis
	redisClient, err := database.SetupRedis(redisAddr, redisPassword)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer func(redisClient *redis.Client) {
		err := redisClient.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(redisClient)

	log.Println("🚀 App configured via Environment Variables. Ready to go!")
}

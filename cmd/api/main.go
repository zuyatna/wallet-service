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
	log.Print("starting wallet service...")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	postgresDSN := os.Getenv("DB_DSN")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if postgresDSN == "" || redisAddr == "" {
		log.Fatal("Configuration error: DB_DSN or REDIS_ADDR environment variables not set")
	}

	// connect to postgreSQL
	db, err := database.SetupPostgres(postgresDSN)
	if err != nil {
		log.Fatal("Error connecting to database")
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	// connect to redis
	redisClient, err := database.SetupRedis(redisAddr, redisPassword)
	if err != nil {
		log.Fatal("Error connecting to redis")
	}
	defer func(redisClient *redis.Client) {
		err := redisClient.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(redisClient)

	log.Print("App configured via env, ready to start server")
}

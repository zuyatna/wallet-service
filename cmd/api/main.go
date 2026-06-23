package main

import (
	"database/sql"
	"log"
	"os"
	"reflect"
	"wallet-service/internal/handler"
	"wallet-service/internal/repository"
	"wallet-service/internal/service"
	"wallet-service/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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

	// Dependency injection
	walletRepo := repository.NewPostgresWalletRepository(db)
	transactionService := service.NewTransactionService(walletRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	router := gin.Default()

	// Register custom validation for decimal.Decimal
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
			if valuer, ok := field.Interface().(decimal.Decimal); ok {
				f, _ := valuer.Float64()
				return f
			}
			return nil
		}, decimal.Decimal{})
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Wallet service is running!",
		})
	})

	// Grouping api routes
	api := router.Group("/api/v1")
	{
		api.POST("/transactions/topup", transactionHandler.TopUp)
	}

	// Start Server
	log.Printf("🚀 Server is running on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

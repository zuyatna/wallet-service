package main

import (
	"database/sql"
	"log"
	"os"
	"reflect"
	"wallet-service/internal/handler"
	"wallet-service/internal/repository"
	"wallet-service/internal/service"
	"wallet-service/pkg/broker"
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

	db, err := database.SetupPostgres(os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Error closing database connection: %v", err)
		}
	}(db)

	redisClient, err := database.SetupRedis(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PASSWORD"))
	if err != nil {
		log.Fatalf("Error connecting to redis: %v", err)
	}
	defer func(redisClient *redis.Client) {
		err := redisClient.Close()
		if err != nil {
			log.Fatalf("Error closing redis connection: %v", err)
		}
	}(redisClient)

	rabbitPublisher, err := broker.NewRabbitMQPublisher(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Error creating rabbitmq publisher: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Dependency injection
	walletRepo := repository.NewPostgresWalletRepository(db)
	transactionService := service.NewTransactionService(walletRepo, redisClient, rabbitPublisher)
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

	router.POST("/api/v1/transactions/topup", transactionHandler.TopUp)

	log.Println("🚀 Server is running on port 8080")
	err = router.Run(":" + port)
	if err != nil {
		return
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

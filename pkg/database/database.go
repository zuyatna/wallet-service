package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func SetupPostgres(dsn string) (*sql.DB, error) {
	// open connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// configure the database connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// ping the database
	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")
	return db, nil
}

func SetupRedis(addr string, password string) (*redis.Client, error) {
	// create the client
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// ping redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to redis")
	return client, nil
}

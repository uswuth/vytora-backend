package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect(databaseURL string) error {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("unable to connect connection pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	Pool = pool
	log.Println("Connected to PostgreSQL Database")
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
		log.Println("Database connection pool closed")
	}
}

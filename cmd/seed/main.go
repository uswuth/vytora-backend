package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:%40aswath@localhost:5432/vrmp_dev?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	_, err = pool.Exec(context.Background(), `
		INSERT INTO users (code, email, password_hash, full_name, role)
		VALUES ('USR001', 'admin@vrmp.com', $1, 'System Administrator', 'system_admin')
		ON CONFLICT (code) DO NOTHING;
	`, string(hash))
	if err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	_, err = pool.Exec(context.Background(), `
		UPDATE entity_sequences SET next_value = 2 WHERE entity_name = 'user';
	`)
	if err != nil {
		log.Fatalf("Failed to update user sequence: %v", err)
	}

	fmt.Println("Seed data inserted: admin user (USR001) created.")
}

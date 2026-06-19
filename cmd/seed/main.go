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

	// Ensure category sequence exists (for fresh DBs or missed migrations)
	_, err = pool.Exec(context.Background(), `
		INSERT INTO entity_sequences (entity_name, next_value) VALUES ('category', 1) ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		log.Fatalf("Failed to seed category sequence: %v", err)
	}

	// Force-correct admin role if user already exists with wrong role
	res, err := pool.Exec(context.Background(), `
		UPDATE users SET role = 'system_admin', is_active = true WHERE email = 'admin@vrmp.com';
	`)
	if err != nil {
		log.Fatalf("Failed to correct admin role: %v", err)
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("✅ Corrected admin user role to 'system_admin' (%d row(s) updated)\n", rowsAffected)
	} else {
		fmt.Println("✅ Admin user already has correct role 'system_admin'")
	}

	fmt.Println("Done. Admin: admin@vrmp.com / admin123 (role: system_admin)")
}

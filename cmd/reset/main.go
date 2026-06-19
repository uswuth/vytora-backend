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

	fmt.Println("Dropping all tables...")

	// Drop in reverse dependency order
	statements := []string{
		"DROP TABLE IF EXISTS categories CASCADE;",
		"DROP TABLE IF EXISTS audit_trail CASCADE;",
		"DROP TABLE IF EXISTS contracts CASCADE;",
		"DROP TABLE IF EXISTS compliance_records CASCADE;",
		"DROP TABLE IF EXISTS risk_assessments CASCADE;",
		"DROP TABLE IF EXISTS vendor_contacts CASCADE;",
		"DROP TABLE IF EXISTS vendors CASCADE;",
		"DROP TABLE IF EXISTS users CASCADE;",
		"DROP TABLE IF EXISTS entity_sequences;",
		"DROP EXTENSION IF EXISTS \"pgcrypto\";",
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(context.Background(), stmt); err != nil {
			log.Printf("Warning: %s -> %v", stmt, err)
		}
	}
	fmt.Println("All tables dropped.")

	fmt.Println("Creating tables...")

	// Create entity_sequences
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS entity_sequences (
			entity_name VARCHAR(50) PRIMARY KEY,
			next_value INTEGER NOT NULL DEFAULT 1
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create entity_sequences: %v", err)
	}

	// Create pgcrypto extension
	_, err = pool.Exec(context.Background(), `CREATE EXTENSION IF NOT EXISTS "pgcrypto";`)
	if err != nil {
		log.Fatalf("Failed to create pgcrypto extension: %v", err)
	}

	// Create users table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL CHECK (
				role IN (
					'system_admin', 'risk_manager', 'compliance_officer',
					'department_manager', 'auditor'
				)
			),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// Create vendors table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS vendors (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			category VARCHAR(100) NOT NULL,
			contact_person VARCHAR(255),
			contact_email VARCHAR(255),
			country VARCHAR(100),
			contract_start_date DATE,
			contract_end_date DATE,
			risk_level VARCHAR(20) NOT NULL CHECK (
				risk_level IN ('Low', 'Medium', 'High', 'Critical')
			),
			status VARCHAR(50) NOT NULL DEFAULT 'Draft' CHECK (
				status IN (
					'Draft', 'Submitted', 'RiskReview', 'ComplianceReview',
					'Approved', 'Rejected', 'Active', 'Inactive'
				)
			),
			assigned_dept_manager_id UUID REFERENCES users(id),
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create vendors table: %v", err)
	}

	// Create vendor_contacts table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS vendor_contacts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(50),
			created_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create vendor_contacts table: %v", err)
	}

	// Create risk_assessments table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS risk_assessments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
			assessment_date DATE NOT NULL,
			assessor_id UUID REFERENCES users(id),
			overall_risk_score DECIMAL(5,2) NOT NULL,
			risk_level VARCHAR(20) NOT NULL CHECK (
				risk_level IN ('Low', 'Medium', 'High', 'Critical')
			),
			security_risk_score DECIMAL(5,2) NOT NULL CHECK (security_risk_score BETWEEN 0 AND 100),
			financial_risk_score DECIMAL(5,2) NOT NULL CHECK (financial_risk_score BETWEEN 0 AND 100),
			operational_risk_score DECIMAL(5,2) NOT NULL CHECK (operational_risk_score BETWEEN 0 AND 100),
			legal_risk_score DECIMAL(5,2) NOT NULL CHECK (legal_risk_score BETWEEN 0 AND 100),
			status VARCHAR(50) NOT NULL DEFAULT 'Draft' CHECK (status IN ('Draft', 'Reviewed', 'Approved')),
			notes TEXT,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create risk_assessments table: %v", err)
	}

	// Create compliance_records table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS compliance_records (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
			certification_type VARCHAR(50) NOT NULL CHECK (
				certification_type IN ('ISO27001', 'SOC2', 'GDPR', 'PCI_DSS')
			),
			status VARCHAR(20) NOT NULL DEFAULT 'Pending' CHECK (status IN ('Pending', 'Approved', 'Expired')),
			valid_from DATE,
			valid_until DATE,
			issued_by VARCHAR(255),
			evidence_url TEXT,
			reviewed_by UUID REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now(),
			CONSTRAINT unique_vendor_cert UNIQUE (vendor_id, certification_type, status)
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create compliance_records table: %v", err)
	}

	// Create contracts table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS contracts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
			contract_number VARCHAR(100) NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			contract_value DECIMAL(15,2),
			renewal_status VARCHAR(50) NOT NULL DEFAULT 'Manual' CHECK (
				renewal_status IN ('Auto-Renew', 'Manual', 'Expiring')
			),
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create contracts table: %v", err)
	}

	// Create audit_trail table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS audit_trail (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			table_name VARCHAR(100) NOT NULL,
			record_id UUID NOT NULL,
			action VARCHAR(20) NOT NULL CHECK (action IN ('CREATE', 'UPDATE', 'DELETE')),
			field_name VARCHAR(100),
			old_value TEXT,
			new_value TEXT,
			changed_by UUID REFERENCES users(id),
			changed_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create audit_trail table: %v", err)
	}

	// Create categories table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(20) UNIQUE NOT NULL,
			name VARCHAR(100) UNIQUE NOT NULL,
			display_name VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(20) NOT NULL DEFAULT 'Active' CHECK (status IN ('Draft', 'Active', 'Inactive')),
			created_by UUID REFERENCES users(id),
			updated_by UUID REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create categories table: %v", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_vendors_status ON vendors(status);",
		"CREATE INDEX IF NOT EXISTS idx_vendors_risk_level ON vendors(risk_level);",
		"CREATE INDEX IF NOT EXISTS idx_vendors_category ON vendors(category);",
		"CREATE INDEX IF NOT EXISTS idx_vendors_assigned_manager ON vendors(assigned_dept_manager_id);",
		"CREATE INDEX IF NOT EXISTS idx_risk_assessments_vendor ON risk_assessments(vendor_id);",
		"CREATE INDEX IF NOT EXISTS idx_risk_assessments_date ON risk_assessments(assessment_date);",
		"CREATE INDEX IF NOT EXISTS idx_compliance_records_vendor ON compliance_records(vendor_id);",
		"CREATE INDEX IF NOT EXISTS idx_compliance_records_cert_type ON compliance_records(certification_type);",
		"CREATE INDEX IF NOT EXISTS idx_contracts_vendor ON contracts(vendor_id);",
		"CREATE INDEX IF NOT EXISTS idx_contracts_end_date ON contracts(end_date);",
		"CREATE INDEX IF NOT EXISTS idx_audit_trail_table_name ON audit_trail(table_name);",
		"CREATE INDEX IF NOT EXISTS idx_audit_trail_record_id ON audit_trail(record_id);",
		"CREATE INDEX IF NOT EXISTS idx_audit_trail_changed_by ON audit_trail(changed_by);",
	}
	for _, idx := range indexes {
		if _, err := pool.Exec(context.Background(), idx); err != nil {
			log.Printf("Warning: %s -> %v", idx, err)
		}
	}
	fmt.Println("Indexes created.")

	// Seed entity sequences
	sequences := []struct {
		name  string
		value int
	}{
		{"user", 1},
		{"vendor", 1},
		{"vendor_contact", 1},
		{"risk_assessment", 1},
		{"compliance_record", 1},
		{"contract", 1},
		{"audit_trail", 1},
		{"category", 1},
	}
	for _, seq := range sequences {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO entity_sequences (entity_name, next_value)
			VALUES ($1, $2) ON CONFLICT DO NOTHING;
		`, seq.name, seq.value)
		if err != nil {
			log.Printf("Warning: failed to seed sequence %s: %v", seq.name, err)
		}
	}
	fmt.Println("Entity sequences seeded.")

	// Seed admin user — force-correct role if it already exists with wrong role
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	_, err = pool.Exec(context.Background(), `
		INSERT INTO users (code, email, password_hash, full_name, role)
		VALUES ('USR001', 'admin@vrmp.com', $1, 'System Administrator', 'system_admin')
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			full_name = EXCLUDED.full_name,
			is_active = true;
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

	fmt.Println("")
	fmt.Println("✅ Database reset complete!")
	fmt.Println("   Admin user: admin@vrmp.com / admin123")
	fmt.Println("   Start server: go run cmd/server/main.go")
}
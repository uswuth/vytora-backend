CREATE EXTENSION IF NOT EXISTS "pgcrypto"
CREATE TABLE
    users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        code VARCHAR(20) UNIQUE NOT NULL,
        email VARCHAR(225) UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        full_name VARCHAR(225) NOT NULL,
        role VARCHAR(50) NOT NULL CHECK (
            role IN (
                'system_admin',
                'risk_manager',
                'compliance_officer',
                'department_manager',
                'auditor'
            )
        ),
        is_active BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMP DEFAULT now (),
        created_at TIMESTAMPTZ DEFAULT now (),
        updated_at TIMESTAMPTZ DEFAULT now ()
    );

CREATE TABLE
    vendors (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
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
                'Draft',
                'Submitted',
                'RiskReview',
                'ComplianceReview',
                'Approved',
                'Rejected',
                'Active',
                'Inactive'
            )
        ),
        assigned_dept_manager_id UUID REFERENCES users (id),
        created_by UUID REFERENCES users (id),
        created_at TIMESTAMPTZ DEFAULT now (),
        updated_at TIMESTAMPTZ DEFAULT now ()
    );

CREATE TABLE
    vendor_contacts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        code VARCHAR(20) UNIQUE NOT NULL,
        vendor_id UUID NOT NULL REFERENCES vendors (id) ON DELETE CASCADE,
        name VARCHAR(225) NOT NULL,
        email VARCHAR(225),
        phone VARCHAR(50),
        created_at TIMESTAMPTZ DEFAULT now ()
    );

CREATE TABLE
    risk_assessments (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        code VARCHAR(20) UNIQUE NOT NULL,
        vendor_id UUID NOT NULL REFERENCES vendors (id) ON DELETE CASCADE,
        assessment_date DATE NOT NULL,
        assessor_id UUID REFERENCES users (id),
        overall_risk_score DECIMAL(5, 2) NOT NULL,
        risk_level VARCHAR(20) NOT NULL CHECK (
            risk_level IN ('Low', 'Medium', 'High', 'Critical')
        ),
        security_risk_score DECIMAL(5, 2) NOT NULL CHECK (security_risk_score BETWEEN 0 AND 100),
        financial_risk_score DECIMAL(5, 2) NOT NULL CHECK (security_risk_score BETWEEN 0 AND 100),
        operational_risk_score DECIMAL(5, 2) NOT NULL CHECK (security_risk_score BETWEEN 0 AND 100),
        legal_risk_score DECIMAL(5, 2) NOT NULL CHECK (security_risk_score BETWEEN 0 AND 100),
        status VARCHAR(50) NOT NULL DEFAULT 'Draft' CHECK (status IN ('Draft', 'Reviewed', 'Approved',)),
        note TEXT,
        created_at TIMESTAMPTZ DEFAULT now (),
        updated_at TIMESTAMPTZ DEFAULT now ()
    );

CREATE TABLE
    compliance_records (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        code VARCHAR(20) UNIQUE NOT NULL,
        vendor_id UUID NOT NULL REFERENCES vendors (id) ON DELETE CASCADE,
        certification_type VARCHAR(50) NOT NULL CHECK (
            certification_type IN ('ISO27001', 'SOC2', 'GDPR', 'PCI_DSS')
        ),
        status VARCHAR(50) NOT NULL DEFAULT 'Pending' CHECK (status IN ('Pending', 'Expired', 'Approved',)),
        valid_from DATE,
        valid_until DATE,
        issued_by VARCHAR(50),
        evidence_url TEXT,
        review_by UUID REFERENCES users (id),
        created_at TIMESTAMPTZ DEFAULT now (),
        updated_at TIMESTAMPTZ DEFAULT now (),
        CONSTRAINT unique_vendor_cert UNIQUE (vendor_id, certification_type, status)
    );

CREATE TABLE
    contracts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        code VARCHAR(20) UNIQUE NOT NULL,
        vendor_id UUID NOT NULL REFERENCES vendors (id) ON DELETE CASCADE,
        contract_number VARCHAR(100) NOT NULL,
        start_date DATE NOT NULL,
        end_date DATE NOT NULL,
        contract_value DECIMAL(5, 2),
        renewal_status VARCHAR(50) NOT NULL DEFAULT 'Manual' CHECK (
            renewal_status IN ('Manual', 'Auto-Renew', 'Expiring')
        ),
        created_at TIMESTAMPTZ DEFAULT now (),
        updated_at TIMESTAMPTZ DEFAULT now (),
    );

CREATE TABLE
    audit_trail (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        code VARCHAR(20) UNIQUE NOT NULL,
        table_name VARCHAR(100) NOT NULL,
        record_id UUID NOT NULL,
        action VARCHAR(20) NOT NULL CHECK (action IN ('CREATE', 'UPDATE', 'Delete')),
        field_name VARCHAR(100),
        old_value TEXT,
        new_value TEXT,
        changed_by UUID REFERENCES users (id),
        changed_at TIMESTAMPTZ DEFAULT now ()
    );

CREATE INDEX idx_vendors_status ON vendors (status);

CREATE INDEX idx_vendors_risk_level ON vendors (risk_level);

CREATE INDEX idx_vendors_category ON vendors (category);

CREATE INDEX idx_vendors_assigned_manager ON vendors (assigned_dept_manager_id);

CREATE INDEX idx_risk_assessments_vendor ON risk_assessments (vendor_id);

CREATE INDEX idx_risk_assessments_date ON risk_assessments (assessment_date);

CREATE INDEX idx_compliance_records_vendor ON compliance_records (vendor_id);

CREATE INDEX idx_compliance_records_cert_type ON compliance_records (certification_type);

CREATE INDEX idx_contracts_vendor ON contracts (vendor_id);

CREATE INDEX idx_contracts_end_date ON contracts (end_date);

CREATE INDEX idx_audit_trail_table_name ON audit_trail (table_name);

CREATE INDEX idx_audit_trail_record_id ON audit_trail (record_id);

CREATE INDEX idx_audit_trail_changed_by ON audit_trail (changed_by);
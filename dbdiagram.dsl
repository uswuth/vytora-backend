// Vytora Backend - Database Schema
// Paste this into https://dbdiagram.io/ to visualize

Table users {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  email VARCHAR(255) [unique, not null]
  password_hash TEXT [not null]
  full_name VARCHAR(255) [not null]
  role VARCHAR(50) [not null]
  is_active BOOLEAN [default: true]
  created_at TIMESTAMPTZ [default: 'now()']
  updated_at TIMESTAMPTZ [default: 'now()']
}

Table categories {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  name VARCHAR(100) [unique, not null]
  display_name VARCHAR(255) [not null]
  description TEXT
  status VARCHAR(20) [not null]
  created_by UUID
  updated_by UUID
  created_at TIMESTAMPTZ [default: 'now()']
  updated_at TIMESTAMPTZ [default: 'now()']
}

Table vendors {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  name VARCHAR(255) [not null]
  category VARCHAR(100) [not null]
  contact_person VARCHAR(255)
  contact_email VARCHAR(255)
  country VARCHAR(100)
  contract_start_date DATE
  contract_end_date DATE
  risk_level VARCHAR(20) [not null]
  status VARCHAR(50) [not null, default: 'Draft']
  assigned_dept_manager_id UUID
  created_by UUID
  created_at TIMESTAMPTZ [default: 'now()']
  updated_at TIMESTAMPTZ [default: 'now()']
}

Table vendor_contacts {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  vendor_id UUID [not null]
  name VARCHAR(255) [not null]
  email VARCHAR(255)
  phone VARCHAR(50)
  created_at TIMESTAMPTZ [default: 'now()']
}

Table risk_assessments {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  vendor_id UUID [not null]
  assessment_date DATE [not null]
  assessor_id UUID
  overall_risk_score DECIMAL(5,2) [not null]
  risk_level VARCHAR(20) [not null]
  security_risk_score DECIMAL(5,2) [not null]
  financial_risk_score DECIMAL(5,2) [not null]
  operational_risk_score DECIMAL(5,2) [not null]
  legal_risk_score DECIMAL(5,2) [not null]
  status VARCHAR(50) [not null, default: 'Draft']
  notes TEXT
  created_at TIMESTAMPTZ [default: 'now()']
  updated_at TIMESTAMPTZ [default: 'now()']
}

Table compliance_records {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  vendor_id UUID [not null]
  certification_type VARCHAR(50) [not null]
  status VARCHAR(20) [not null, default: 'Pending']
  valid_from DATE
  valid_until DATE
  issued_by VARCHAR(255)
  evidence_url TEXT
  reviewed_by UUID
  created_at TIMESTAMPTZ [default: 'now()']
  updated_at TIMESTAMPTZ [default: 'now()']
}

Table contracts {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  vendor_id UUID [not null]
  contract_number VARCHAR(100) [not null]
  start_date DATE [not null]
  end_date DATE [not null]
  contract_value DECIMAL(15,2)
  renewal_status VARCHAR(50) [not null, default: 'Manual']
  created_at TIMESTAMPTZ [default: 'now()']
  updated_at TIMESTAMPTZ [default: 'now()']
}

Table audit_trail {
  id UUID [pk, default: 'gen_random_uuid()']
  code VARCHAR(20) [unique, not null]
  table_name VARCHAR(100) [not null]
  record_id UUID [not null]
  action VARCHAR(20) [not null]
  field_name VARCHAR(100)
  old_value TEXT
  new_value TEXT
  changed_by UUID
  changed_at TIMESTAMPTZ [default: 'now()']
}

// Relationships
Ref: vendors.assigned_dept_manager_id > users.id
Ref: vendors.created_by > users.id
Ref: categories.created_by > users.id
Ref: categories.updated_by > users.id
Ref: vendor_contacts.vendor_id > vendors.id [delete: cascade]
Ref: risk_assessments.vendor_id > vendors.id [delete: cascade]
Ref: risk_assessments.assessor_id > users.id
Ref: compliance_records.vendor_id > vendors.id [delete: cascade]
Ref: compliance_records.reviewed_by > users.id
Ref: contracts.vendor_id > vendors.id [delete: cascade]
Ref: audit_trail.changed_by > users.id
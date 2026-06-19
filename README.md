# VRMP — Vendor Risk Management Platform

**Backend API for managing vendor onboarding, risk assessment, compliance tracking, and contract lifecycle.**

---

## What Is This System

VRMP is a role-based backend system that digitizes vendor risk management for organizations. It replaces spreadsheets and ad-hoc emails with a structured workflow: vendors are registered, risk-assessed, compliance-certified, contracted, and approved through defined stages. Every action is audited, and reports are generated for leadership.

---

## Who It Is For

| Role | What They Do |
|------|-------------|
| **System Admin** | Full control — creates users, manages categories, vendors, all configurations |
| **Risk Manager** | Creates and reviews risk assessments, generates reports |
| **Compliance Officer** | Uploads and tracks certifications (ISO 27001, SOC 2, GDPR, PCI DSS) |
| **Department Manager** | Submits vendor requests, views assigned vendors |
| **Auditor** | Reads audit logs and reports (no write access) |
| **Viewer / Editor** | Limited read/write based on assigned permissions |

---

## Tech Stack

- **Language:** Go 1.24
- **Framework:** Fiber v2
- **Database:** PostgreSQL (via pgxpool)
- **Auth:** JWT HS256 (golang-jwt/jwt/v5)
- **Validation:** go-playground/validator/v10
- **Logging:** zerolog (structured JSON)
- **Config:** Viper + .env file
- **Metrics:** Prometheus exposition format
- **Docs:** frontendDoc.md + endpoint.md

---

## What Was Built (Session Summary)

### Core Entities (CRUD + Workflow)
1. **Users** — admin-managed, bcrypt-hashed passwords, JWT auth
2. **Categories** — vendor categorization (unique name per category)
3. **Vendors** — full lifecycle with workflow state machine
4. **Risk Assessments** — scored 0–100 across 4 dimensions + overall
5. **Compliance Records** — certification tracking with auto-status from dates
6. **Contracts** — contract value, dates, renewal status
7. **Audit Trail** — immutable log of all state-changing actions
8. **Reports** — summary and monthly onboarding dashboards

### Vendor Workflow (State Machine)
```
Draft → Submitted → RiskReview → ComplianceReview → Approved → Active
                                          ↘ Rejected
```
Each transition is a protected endpoint with permission checks.

### Observability & Security Added in This Session
- **Request ID middleware** — `X-Request-ID` header, propagated to all log lines
- **Structured logging** — zerolog JSON logs with `request_id`, `service=VRMP`, latency, status
- **CORS** — locked to `ALLOWED_ORIGINS` from .env (no hardcoded fallback)
- **Rate limiting** — 100 requests/minute per IP (token bucket)
- **Health checks** — `/healthz` (liveness) and `/readyz` (readiness) with IP allowlist
- **Prometheus metrics** — `/metrics` endpoint with request count, latency histograms
- **JWT session extension** — `POST /api/v1/auth/extend` for frontend countdown + extend button
- **IP-restricted health/metrics** — configurable via `HEALTH_ALLOWED_IPS` in .env

### Configuration (.env)
All config values panic if missing — no hardcoded secrets:
- `DATABASE_URL` — PostgreSQL connection string
- `JWT_SECRET` — HMAC signing key
- `JWT_EXPIRY_HOURS` — token lifetime (default 24h)
- `ALLOWED_ORIGINS` — comma-separated CORS origins
- `HEALTH_ALLOWED_IPS` — comma-separated IPs for health endpoints

### Frontend Support Files
- **`frontendDoc.md`** — full API reference with TypeScript types, auth flow, token countdown instructions
- **`endpoint.md`** — concise route table with method, path, auth, permission, request/response shapes

---

## Database Schema

### Tables
- `users` — id, code, email, password_hash, full_name, role, is_active, timestamps
- `categories` — id, code, name (unique), display_name, description, status, created_by, timestamps
- `vendors` — id, code, name, category, contact fields, contract dates, risk_level, status, assigned_dept_manager_id, created_by, timestamps
- `risk_assessments` — id, code, vendor_id, assessment_date, assessor_id, scores (overall/security/financial/operational/legal), risk_level, status, notes
- `compliance_records` — id, code, vendor_id, certification_type, status, valid_from, valid_until, issued_by, evidence_url, reviewed_by
- `contracts` — id, code, vendor_id, contract_number, start_date, end_date, contract_value, renewal_status
- `audit_trail` — id, entity_type, entity_code, action, performed_by, performed_at, details (JSONB)

### Sequences
- `user_code_seq`, `vendor_code_seq`, `risk_assessment_code_seq`, `compliance_record_code_seq`, `contract_code_seq`, `category_code_seq` — auto-increment codes like `USR-001`, `VEN-001`, `RA-001`, `CR-001`, `CT-001`, `CAT-001`

---

## Permissions Matrix

| Permission | Roles |
|------------|-------|
| `canManageUsers` | system_admin |
| `canCreateVendor` | system_admin |
| `canEditVendor` | system_admin |
| `canDeleteVendor` | system_admin |
| `canSubmitVendorRequest` | system_admin, department_manager |
| `canReviewRisk` | system_admin, risk_manager |
| `canReviewCompliance` | system_admin, compliance_officer |
| `canApproveRisk` | system_admin, risk_manager |
| `canManageCategories` | system_admin |
| `canViewCategories` | system_admin, risk_manager, compliance_officer, department_manager, auditor |
| `canViewAuditHistory` | system_admin, auditor |
| `canAccessAllReports` | system_admin, auditor |
| `canViewAssignedVendors` | department_manager |

---

## Quick Start

```bash
# 1. Set up .env
cp .env.example .env   # then fill in values

# 2. Run migrations
go run cmd/seed/main.go

# 3. Start server
go run cmd/server/main.go

# 4. Health check
curl http://localhost:8080/healthz

# 5. Login
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@vrmp.com","password":"admin123"}'
```

---

## File Structure

```
cmd/
  server/main.go         — entry point, wires middleware + routes
  seed/main.go            — runs migrations + seed data
  reset/main.go           — drops and recreates DB
internal/
  config/config.go        — Viper .env loader, panic on missing keys
  database/               — pgxpool connect/close
  middleware/
    auth.go               — JWT Bearer validation
    rbac.go               — permission-based gate
    rate_limiter.go       — per-IP token bucket (100/min)
    structured_logger.go  — zerolog with request_id
    metrics.go            — Prometheus text format exposition
    health.go             — /healthz, /readyz with IP allowlist
  services/
    jwt_service.go        — GenerateToken, ValidateToken, TokenTTL, ExtendToken
    sequence_service.go   — next code generator per entity
  handlers/
    auth_handler.go       — Login, ExtendSession
  entity/
    user/                 — CRUD, login, role management
    vendor/               — CRUD + workflow transitions
    category/             — CRUD
    risk_assessment/      — CRUD + approve
    compliance_record/    — CRUD + expiring query
    contract/             — CRUD + expiring query
    audit_trail/          — read-only log
    report/               — summary + monthly onboarding
migrations/              — SQL migration files
endpoint.md              — concise API reference
frontendDoc.md           — detailed API docs with TypeScript types
```

---

## Key Design Decisions

1. **Stateless JWT** — no server-side session storage; tokens self-contained
2. **Code-based IDs** — human-readable codes (VEN-001) alongside UUIDs
3. **Audit by default** — all state changes go through repositories that write audit entries
4. **IP allowlisting** — health/metrics endpoints hidden from public internet by default
5. **Panic on missing config** — fails fast at startup if .env is incomplete
6. **Per-IP rate limiting** — prevents abuse without affecting multi-user deployments

---

## Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `DATABASE_URL` | PostgreSQL connection | `postgres://user:pass@localhost:5432/vrmp?sslmode=disable` |
| `JWT_SECRET` | HMAC signing key | `supersecretkey123!@#` |
| `JWT_EXPIRY_HOURS` | Token lifetime | `24` |
| `ALLOWED_ORIGINS` | CORS whitelist | `http://localhost:3000` |
| `HEALTH_ALLOWED_IPS` | Health endpoint IPs | `127.0.0.1` |

---

## Default Admin Credentials

Created by seed script:
- **Email:** `admin@vrmp.com`
- **Password:** `admin123`
- **Role:** `system_admin`

> Change password after first login in production.

---

## License

Internal use — Uswuth Projects.
# VRMP Backend API Documentation

**Base URL:** `http://localhost:8080`  
**Content-Type:** `application/json`  
**Auth:** Bearer JWT in `Authorization` header

---

## Table of Contents
1. [Authentication](#authentication)
2. [Users](#users)
3. [Vendors](#vendors)
4. [Risk Assessments](#risk-assessments)
5. [Compliance Records](#compliance-records)
6. [Contracts](#contracts)
7. [Audit Trail](#audit-trail)
8. [Reports](#reports)
9. [Categories](#categories)
10. [Health & Metrics](#health--metrics)
11. [TypeScript Types](#typescript-types)
12. [Error Handling](#error-handling)
13. [Token Management & Session Extension](#token-management--session-extension)

---

## Authentication

### `POST /api/v1/login`
Login with email + password. Returns JWT token + user profile.

**Request:**
```json
{
  "email": "admin@vytora.com",
  "password": "SecurePass123!"
}
```

**Response 200:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiJ9...",
  "expires_in": 86400,
  "user": {
    "id": "a1b2c3d4-...",
    "code": "USR-001",
    "email": "admin@vytora.com",
    "full_name": "Admin User",
    "role": "admin",
    "is_active": true,
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-01-15T10:00:00Z"
  }
}
```

**Errors:**
- `400` — invalid email format or password < 8 chars
- `401` — wrong credentials
- `403` — account deactivated

---

### `POST /api/v1/auth/extend`
Extend the current JWT session. Requires valid (not fully expired) token.

**Headers:** `Authorization: Bearer <token>`

**Request:** (empty body)

**Response 200:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiJ9...",
  "expires_in": 86400
}
```

**Errors:**
- `401` — missing/invalid auth header or token fully expired

---

## Users

### `POST /api/v1/users`
Create a new user. **Requires:** `canManageUsers`

**Request:**
```json
{
  "email": "newuser@vytora.com",
  "password": "SecurePass123!",
  "full_name": "New User",
  "role": "viewer"
}
```

**Response 201:**
```json
{
  "id": "...",
  "code": "USR-002",
  "email": "newuser@vytora.com",
  "full_name": "New User",
  "role": "viewer",
  "is_active": true,
  "created_at": "...",
  "updated_at": "..."
}
```

### `GET /api/v1/users`
List all users. **Requires:** `canManageUsers`

**Response 200:**
```json
[
  {
    "id": "...",
    "code": "USR-001",
    "email": "admin@vytora.com",
    "full_name": "Admin User",
    "role": "admin",
    "is_active": true,
    "created_at": "...",
    "updated_at": "..."
  }
]
```
*(password_hash never returned)*

### `GET /api/v1/users/:id`
Get single user. **Requires:** `canManageUsers`

### `PUT /api/v1/users/:id/role`
Update user role. **Requires:** `canManageUsers`

**Request:**
```json
{ "role": "editor" }
```

### `PUT /api/v1/users/:id/deactivate`
Deactivate user. **Requires:** `canManageUsers`

### `PUT /api/v1/users/:id/activate`
Activate user. **Requires:** `canManageUsers`

---

## Vendors

### `POST /api/v1/vendors`
Create vendor. **Requires:** `canCreateVendor`

**Request:**
```json
{
  "name": "Acme Corp",
  "category": "Technology",
  "contact_person": "John Doe",
  "contact_email": "john@acme.com",
  "country": "USA",
  "contract_start_date": "2026-01-01",
  "contract_end_date": "2027-12-31",
  "risk_level": "Medium",
  "status": "Pending",
  "assigned_dept_manager_id": "manager-uuid-here"
}
```
*(Dates: `YYYY-MM-DD`. IDs: null or omit if not assigned.)*

**Response 201:** Full Vendor object.

### `GET /api/v1/vendors`
List vendors (editable view). **Requires:** `canEditVendor`

**Query params:** none required  
**Response 200:** `Vendor[]`

### `GET /api/v1/vendors/:code`
Get vendor by code. **Requires:** `canEditVendor`

### `PUT /api/v1/vendors/:code`
Update vendor. **Requires:** `canEditVendor`

**Request:** Same shape as Create (all fields required).

### `DELETE /api/v1/vendors/:code`
Delete vendor. **Requires:** `canDeleteVendor`

### `PUT /api/v1/vendors/:code/submit`
Submit vendor for review. **Requires:** `canSubmitVendorRequest`

### `PUT /api/v1/vendors/:code/review-risk`
Review vendor risk. **Requires:** `canReviewRisk`

### `PUT /api/v1/vendors/:code/review-compliance`
Review vendor compliance. **Requires:** `canReviewCompliance`

### `PUT /api/v1/vendors/:code/approve`
Approve vendor. **Requires:** `canEditVendor`

### `PUT /api/v1/vendors/:code/reject`
Reject vendor. **Requires:** `canEditVendor`

---

## Risk Assessments

### `POST /api/v1/risk-assessments`
Create risk assessment. **Requires:** `canCreateRiskAssessment`

**Request:**
```json
{
  "vendor_code": "VEN-001",
  "assessment_date": "2026-06-01",
  "overall_risk_score": 65.5,
  "risk_level": "High",
  "security_risk_score": 80.0,
  "financial_risk_score": 50.0,
  "operational_risk_score": 60.0,
  "legal_risk_score": 45.0,
  "status": "Draft",
  "notes": "Key findings..."
}
```
*(Scores 0–100. Status: `Draft` | `Reviewed` | `Approved`)*

**Response 201:** Full RiskAssessment object.

### `GET /api/v1/risk-assessments`
List risk assessments. **Requires:** `canReviewRisk`

**Query params:** `vendor_code`, `risk_level`, `status`, `limit`, `offset`

**Response 200:**
```json
{
  "data": [ /* RiskAssessment[] */ ],
  "total": 42
}
```

### `GET /api/v1/risk-assessments/:code`
Get by code. **Requires:** `canReviewRisk`

### `PUT /api/v1/risk-assessments/:code`
Update. **Requires:** `canReviewRisk`

**Request:** Same as Create (all fields required).

### `DELETE /api/v1/risk-assessments/:code`
Delete. **Requires:** `canReviewRisk`

### `PUT /api/v1/risk-assessments/:code/approve`
Approve. **Requires:** `canApproveRisk`

---

## Compliance Records

### `POST /api/v1/compliance`
Create compliance record. **Requires:** `canReviewCompliance`

**Request:**
```json
{
  "vendor_code": "VEN-001",
  "certification_type": "ISO 27001",
  "valid_from": "2026-01-01",
  "valid_until": "2027-01-01",
  "issued_by": "BSI Group",
  "evidence_url": "https://..."
}
```
*(Dates: `YYYY-MM-DD`. Status auto-calculated from dates.)*

**Response 201:** Full ComplianceRecord object.

### `GET /api/v1/compliance`
List by vendor. **Requires:** `canReviewCompliance`

**Query param:** `vendor_code` (required)

### `GET /api/v1/compliance/:code`
Get by code. **Requires:** `canReviewCompliance`

### `PUT /api/v1/compliance/:code`
Update. **Requires:** `canReviewCompliance`

**Request:**
```json
{
  "certification_type": "SOC 2",
  "status": "Approved",
  "valid_from": "2026-02-01",
  "valid_until": "2027-02-01",
  "issued_by": "Deloitte",
  "evidence_url": "https://..."
}
```
*(All fields required. Partial update not supported yet.)*

### `DELETE /api/v1/compliance/:code`
Delete. **Requires:** `canReviewCompliance`

### `GET /api/v1/compliance/expiring`
Get expiring certifications (default 30 days). **Requires:** auth

**Query param:** `days` (optional, default `30`)

---

## Contracts

### `POST /api/v1/contracts`
Create contract. **Requires:** `canEditVendor`

**Request:**
```json
{
  "vendor_code": "VEN-001",
  "contract_number": "CTR-2026-001",
  "start_date": "2026-01-01",
  "end_date": "2027-01-01",
  "contract_value": 150000.00,
  "renewal_status": "Active"
}
```
*(Dates: `YYYY-MM-DD`. `contract_value`: number or null.)*

**Response 201:** Full Contract object.

### `GET /api/v1/contracts`
List by vendor. **Requires:** `canEditVendor`

**Query param:** `vendor_code` (required)

### `GET /api/v1/contracts/:code`
Get by code. **Requires:** `canEditVendor`

### `PUT /api/v1/contracts/:code`
Update. **Requires:** `canEditVendor`

**Request:** Same as Create (all fields required).

### `DELETE /api/v1/contracts/:code`
Delete. **Requires:** `canEditVendor`

### `GET /api/v1/contracts/expiring`
Get expiring contracts (default 30 days). **Requires:** auth

**Query param:** `days` (optional, default `30`)

---

## Audit Trail

### `GET /api/v1/audit`
List audit trail entries. **Requires:** `canViewAuditHistory`

**Response 200:**
```json
[
  {
    "id": "...",
    "entity_type": "vendor",
    "entity_code": "VEN-001",
    "action": "update",
    "performed_by": "USR-001",
    "performed_at": "2026-06-19T20:00:00Z",
    "details": { "field": "status", "old": "Pending", "new": "Approved" }
  }
]
```

---

## Reports

### `GET /api/v1/reports/summary`
Dashboard summary. **Requires:** `canAccessAllReports`

**Response 200:**
```json
{
  "total_vendors": 120,
  "active_vendors": 85,
  "pending_approvals": 12,
  "high_risk_count": 15,
  "expiring_contracts_30d": 8,
  "expiring_compliance_30d": 5
}
```

### `GET /api/v1/reports/monthly-onboarding`
Monthly vendor onboarding stats. **Requires:** `canAccessAllReports`

**Response 200:**
```json
{
  "months": ["2026-01", "2026-02", "2026-03", "2026-04", "2026-05", "2026-06"],
  "onboarded": [5, 8, 3, 10, 6, 4],
  "approved": [4, 7, 3, 9, 5, 3]
}
```

### `GET /api/v1/reports/summary-2`
Same as summary but scoped to assigned vendors. **Requires:** `canViewAssignedVendors`

### `GET /api/v1/reports/monthly-onboarding-2`
Same as monthly-onboarding but scoped. **Requires:** `canViewAssignedVendors`

> Routes ending in `-2` filter results to vendors assigned to the current user's department.

---

## Categories

### `POST /api/v1/categories`
Create category. **Requires:** `canManageCategories`

**Request:**
```json
{
  "name": "technology",
  "display_name": "Technology",
  "description": "IT and software vendors",
  "status": "Active"
}
```
*(`name` must be unique, lowercase, no spaces.)*

**Response 201:** Full Category object.

### `GET /api/v1/categories`
List categories. **Requires:** `canViewCategories`

**Query params:** `search`, `status`

**Response 200:**
```json
{
  "data": [ /* Category[] */ ],
  "total": 25
}
```

### `GET /api/v1/categories/:code`
Get by code. **Requires:** `canViewCategories`

### `PUT /api/v1/categories/:code`
Update. **Requires:** `canManageCategories`

**Request:**
```json
{
  "display_name": "Technology & Software",
  "description": "Updated description",
  "status": "Active"
}
```

### `DELETE /api/v1/categories/:code`
Delete. **Requires:** `canManageCategories`

---

## Health & Metrics

### `GET /healthz`
Liveness probe. Returns `200 OK` if server is running.

**Response:** `{ "status": "healthy" }`

### `GET /readyz`
Readiness probe. Checks database connection.

**Response:** `{ "status": "ready", "database": "connected" }`

> Both endpoints are IP-allowlisted (only `127.0.0.1` by default). Configure `HEALTH_ALLOWED_IPS` in `.env` for remote access.

### `GET /metrics`
Prometheus metrics exposition format.

**Response:** Text/plain prometheus metrics.

---

## TypeScript Types

```typescript
// --- Core Auth ---
interface LoginRequest {
  email: string;
  password: string;
}

interface LoginResponse {
  token: string;
  expires_in: number; // seconds until expiry
  user: User;
}

interface ExtendSessionResponse {
  token: string;
  expires_in: number;
}

// --- User ---
interface User {
  id: string;
  code: string;
  email: string;
  full_name: string;
  role: 'admin' | 'editor' | 'viewer';
  is_active: boolean;
  created_at: string; // ISO 8601
  updated_at: string; // ISO 8601
}

// --- Vendor ---
interface Vendor {
  id: string;
  code: string;
  name: string;
  category: string;
  contact_person?: string;
  contact_email?: string;
  country?: string;
  contract_start_date?: string; // YYYY-MM-DD or null
  contract_end_date?: string;   // YYYY-MM-DD or null
  risk_level: string;
  status: string;
  assigned_dept_manager_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

// --- Risk Assessment ---
interface RiskAssessment {
  id: string;
  code: string;
  vendor_id: string;
  vendor_code?: string;
  assessment_date: string; // ISO 8601
  assessor_id?: string;
  overall_risk_score: number; // 0-100
  risk_level: 'Low' | 'Medium' | 'High' | 'Critical';
  security_risk_score: number;
  financial_risk_score: number;
  operational_risk_score: number;
  legal_risk_score: number;
  status: 'Draft' | 'Reviewed' | 'Approved';
  notes?: string;
  created_at: string;
  updated_at: string;
}

interface CreateRiskAssessmentRequest {
  vendor_code: string;
  assessment_date: string; // YYYY-MM-DD
  overall_risk_score: number;
  risk_level: string;
  security_risk_score: number;
  financial_risk_score: number;
  operational_risk_score: number;
  legal_risk_score: number;
  status: string;
  notes?: string;
}

// --- Compliance Record ---
interface ComplianceRecord {
  id: string;
  code: string;
  vendor_id: string;
  vendor_code?: string;
  certification_type: string;
  status: string;
  valid_from?: string; // YYYY-MM-DD or null
  valid_until?: string; // YYYY-MM-DD or null
  issued_by: string;
  evidence_url: string;
  reviewed_by?: string;
  created_at: string;
  updated_at: string;
}

interface CreateComplianceRequest {
  vendor_code: string;
  certification_type: string;
  valid_from: string; // YYYY-MM-DD
  valid_until: string; // YYYY-MM-DD
  issued_by: string;
  evidence_url: string;
}

// --- Contract ---
interface Contract {
  id: string;
  code: string;
  vendor_id: string;
  vendor_code?: string;
  contract_number: string;
  start_date: string; // ISO 8601
  end_date: string;   // ISO 8601
  contract_value?: number; // nullable
  renewal_status: string;
  created_at: string;
  updated_at: string;
}

interface CreateContractRequest {
  vendor_code: string;
  contract_number: string;
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  contract_value?: number;
  renewal_status: string;
}

// --- Category ---
interface Category {
  id: string;
  code: string;
  name: string;
  display_name: string;
  description: string;
  status: string;
  created_by: string;
  created_at: string;
  updated_at?: string;
}

interface CreateCategoryRequest {
  name: string;
  display_name: string;
  description: string;
  status: string;
}

// --- Audit ---
interface AuditEntry {
  id: string;
  entity_type: string;
  entity_code: string;
  action: string;
  performed_by: string;
  performed_at: string;
  details: Record<string, any>;
}

// --- Reports ---
interface ReportSummary {
  total_vendors: number;
  active_vendors: number;
  pending_approvals: number;
  high_risk_count: number;
  expiring_contracts_30d: number;
  expiring_compliance_30d: number;
}

interface MonthlyOnboardingReport {
  months: string[];
  onboarded: number[];
  approved: number[];
}
```

---

## Error Handling

All errors follow this shape:

```json
{
  "error": "short description",
  "details": "optional longer message"
}
```

**Common status codes:**
| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad request (validation / parse error) |
| `401` | Unauthorized (missing/invalid/expired token) |
| `403` | Forbidden (valid token but insufficient permissions) |
| `404` | Not found |
| `409` | Conflict (e.g., duplicate category name) |
| `429` | Rate limited (100 req/min per IP) |
| `500` | Server error |

---

## Token Management & Session Extension

### Overview
- JWT tokens expire after `JWT_EXPIRY_HOURS` (default: 24h) from `.env`
- Frontend should track `expires_in` returned from login/extend
- When countdown reaches near 0, call `/api/v1/auth/extend` to get a fresh token

### Frontend Countdown Flow
```
1. Login → store token + expires_in + timestamp
2. Every second: remaining = expires_in - (now - loginTimestamp)
3. If remaining < 60s: show "Extend Session" button
4. Click button → POST /api/v1/auth/extend with current token
   → swap stored token + reset timer
5. If extend fails (401) or token fully expires → redirect to login
```

### Extend Session Call (frontend example)
```typescript
async function extendSession() {
  const res = await fetch('/api/v1/auth/extend', {
    headers: {
      'Authorization': `Bearer ${currentToken}`,
    },
  });
  if (!res.ok) {
    // Token expired — redirect to login
    logout();
    return;
  }
  const data: ExtendSessionResponse = await res.json();
  saveToken(data.token, data.expires_in);
}
```

### Important Notes
- The `Authorization` header must be `Bearer <token>` (case-insensitive `bearer`)
- `expires_in` from extend is the **new total seconds**, not the remaining from old token
- Tokens are **stateless** — no server-side session storage
- If user is deactivated after login, next request will fail with `403`
- Rate limit: 100 requests/minute per IP (includes auth endpoints)

---

## Base Response Wrapper
All list endpoints that support pagination return:
```json
{
  "data": [...],
  "total": 42
}
```

Non-paginated list endpoints (like `/api/v1/audit`, `/api/v1/vendors`) return raw arrays.
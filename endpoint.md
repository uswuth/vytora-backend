# API Endpoints

Base URL: `http://localhost:8080`  
Auth: `Authorization: Bearer <token>`  
Content-Type: `application/json`

---

## Auth

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| POST | /api/v1/login | Public | — |
| POST | /api/v1/auth/extend | Bearer | — |

### POST /api/v1/login
**Request**
```json
{ "email": "string", "password": "string" }
```
**Response 200**
```json
{ "token": "string", "expires_in": 86400, "user": { "id": "uuid", "code": "string", "email": "string", "full_name": "string", "role": "string", "is_active": true, "created_at": "ISO8601", "updated_at": "ISO8601" } }
```

### POST /api/v1/auth/extend
**Headers:** `Authorization: Bearer <token>`  
**Request:** empty body  
**Response 200**
```json
{ "token": "string", "expires_in": 86400 }
```

---

## Users

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| POST | /api/v1/users | Bearer | canManageUsers |
| GET | /api/v1/users | Bearer | canManageUsers |
| GET | /api/v1/users/:id | Bearer | canManageUsers |
| PUT | /api/v1/users/:id/role | Bearer | canManageUsers |
| PUT | /api/v1/users/:id/deactivate | Bearer | canManageUsers |
| PUT | /api/v1/users/:id/activate | Bearer | canManageUsers |

### POST /api/v1/users
**Request**
```json
{ "email": "string", "password": "string", "full_name": "string", "role": "admin|editor|viewer" }
```
**Response 201** — User object (no `password_hash`)

### GET /api/v1/users
**Response 200** — User[]

### GET /api/v1/users/:id
**Response 200** — User

### PUT /api/v1/users/:id/role
**Request**
```json
{ "role": "string" }
```
**Response 204**

### PUT /api/v1/users/:id/deactivate
**Response 204**

### PUT /api/v1/users/:id/activate
**Response 204**

---

## Vendors

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| POST | /api/v1/vendors | Bearer | canCreateVendor |
| GET | /api/v1/vendors | Bearer | canEditVendor |
| GET | /api/v1/vendors/:code | Bearer | canEditVendor |
| PUT | /api/v1/vendors/:code | Bearer | canEditVendor |
| DELETE | /api/v1/vendors/:code | Bearer | canDeleteVendor |
| PUT | /api/v1/vendors/:code/submit | Bearer | canSubmitVendorRequest |
| PUT | /api/v1/vendors/:code/review-risk | Bearer | canReviewRisk |
| PUT | /api/v1/vendors/:code/review-compliance | Bearer | canReviewCompliance |
| PUT | /api/v1/vendors/:code/approve | Bearer | canEditVendor |
| PUT | /api/v1/vendors/:code/reject | Bearer | canEditVendor |

### POST /api/v1/vendors
**Request**
```json
{ "name": "string", "category": "string", "contact_person": "string?", "contact_email": "string?", "country": "string?", "contract_start_date": "YYYY-MM-DD?", "contract_end_date": "YYYY-MM-DD?", "risk_level": "Low|Medium|High|Critical", "status": "string", "assigned_dept_manager_id": "uuid?" }
```
**Response 201** — Vendor object

### GET /api/v1/vendors
**Response 200** — Vendor[]

### GET /api/v1/vendors/:code
**Response 200** — Vendor

### PUT /api/v1/vendors/:code
**Request** — same as Create (all fields required)  
**Response 200** — Vendor

### DELETE /api/v1/vendors/:code
**Response 204**

### Workflow endpoints (PUT /api/v1/vendors/:code/*)
**Request:** empty body  
**Response 204**

---

## Risk Assessments

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| POST | /api/v1/risk-assessments | Bearer | canCreateRiskAssessment |
| GET | /api/v1/risk-assessments | Bearer | canReviewRisk |
| GET | /api/v1/risk-assessments/:code | Bearer | canReviewRisk |
| PUT | /api/v1/risk-assessments/:code | Bearer | canReviewRisk |
| DELETE | /api/v1/risk-assessments/:code | Bearer | canReviewRisk |
| PUT | /api/v1/risk-assessments/:code/approve | Bearer | canApproveRisk |

### POST /api/v1/risk-assessments
**Request**
```json
{ "vendor_code": "string", "assessment_date": "YYYY-MM-DD", "overall_risk_score": 0-100, "risk_level": "Low|Medium|High|Critical", "security_risk_score": 0-100, "financial_risk_score": 0-100, "operational_risk_score": 0-100, "legal_risk_score": 0-100, "status": "Draft|Reviewed|Approved", "notes": "string?" }
```
**Response 201** — RiskAssessment object

### GET /api/v1/risk-assessments
**Query params:** `vendor_code?`, `risk_level?`, `status?`, `limit?`, `offset?`  
**Response 200**
```json
{ "data": [RiskAssessment], "total": 42 }
```

### GET /api/v1/risk-assessments/:code
**Response 200** — RiskAssessment

### PUT /api/v1/risk-assessments/:code
**Request** — same as Create  
**Response 200** — RiskAssessment

### DELETE /api/v1/risk-assessments/:code
**Response 204**

### PUT /api/v1/risk-assessments/:code/approve
**Response 204**

---

## Compliance Records

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| POST | /api/v1/compliance | Bearer | canReviewCompliance |
| GET | /api/v1/compliance | Bearer | canReviewCompliance |
| GET | /api/v1/compliance/:code | Bearer | canReviewCompliance |
| PUT | /api/v1/compliance/:code | Bearer | canReviewCompliance |
| DELETE | /api/v1/compliance/:code | Bearer | canReviewCompliance |
| GET | /api/v1/compliance/expiring | Bearer | — |

### POST /api/v1/compliance
**Request**
```json
{ "vendor_code": "string", "certification_type": "string", "valid_from": "YYYY-MM-DD", "valid_until": "YYYY-MM-DD", "issued_by": "string", "evidence_url": "string" }
```
**Response 201** — ComplianceRecord object

### GET /api/v1/compliance
**Query param:** `vendor_code` (required)  
**Response 200** — ComplianceRecord[]

### GET /api/v1/compliance/:code
**Response 200** — ComplianceRecord

### PUT /api/v1/compliance/:code
**Request**
```json
{ "certification_type": "string", "status": "string", "valid_from": "YYYY-MM-DD", "valid_until": "YYYY-MM-DD", "issued_by": "string", "evidence_url": "string" }
```
**Response 200** — ComplianceRecord

### DELETE /api/v1/compliance/:code
**Response 204**

### GET /api/v1/compliance/expiring
**Query param:** `days?` (default 30)  
**Response 200** — ComplianceRecord[]

---

## Contracts

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| POST | /api/v1/contracts | Bearer | canEditVendor |
| GET | /api/v1/contracts | Bearer | canEditVendor |
| GET | /api/v1/contracts/:code | Bearer | canEditVendor |
| PUT | /api/v1/contracts/:code | Bearer | canEditVendor |
| DELETE | /api/v1/contracts/:code | Bearer | canEditVendor |
| GET | /api/v1/contracts/expiring | Bearer | — |

### POST /api/v1/contracts
**Request**
```json
{ "vendor_code": "string", "contract_number": "string", "start_date": "YYYY-MM-DD", "end_date": "YYYY-MM-DD", "contract_value": 1000.00?, "renewal_status": "string" }
```
**Response 201** — Contract object

### GET /api/v1/contracts
**Query param:** `vendor_code`  
**Response 200** — Contract[]

### GET /api/v1/contracts/:code
**Response 200** — Contract

### PUT /api/v1/contracts/:code
**Request** — same as Create  
**Response 200** — Contract

### DELETE /api/v1/contracts/:code
**Response 204**

### GET /api/v1/contracts/expiring
**Query param:** `days?` (default 30)  
**Response 200** — Contract[]

---

## Audit

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| GET | /api/v1/audit | Bearer | canViewAuditHistory |

### GET /api/v1/audit
**Response 200** — AuditEntry[]

---

## Reports

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| GET | /api/v1/reports/summary | Bearer | canAccessAllReports |
| GET | /api/v1/reports/monthly-onboarding | Bearer | canAccessAllReports |
| GET | /api/v1/reports/summary-2 | Bearer | canViewAssignedVendors |
| GET | /api/v1/reports/monthly-onboarding-2 | Bearer | canViewAssignedVendors |

### GET /api/v1/reports/summary
**Response 200**
```json
{ "total_vendors": 120, "active_vendors": 85, "pending_approvals": 12, "high_risk_count": 15, "expiring_contracts_30d": 8, "expiring_compliance_30d": 5 }
```

### GET /api/v1/reports/monthly-onboarding
**Response 200**
```json
{ "months": ["2026-01", ...], "onboarded": [5, ...], "approved": [4, ...] }
```

### GET /api/v1/reports/summary-2
Same as summary, scoped to assigned vendors.

### GET /api/v1/reports/monthly-onboarding-2
Same as monthly-onboarding, scoped to assigned vendors.

---

## Categories

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| POST | /api/v1/categories | Bearer | canManageCategories |
| GET | /api/v1/categories | Bearer | canViewCategories |
| GET | /api/v1/categories/:code | Bearer | canViewCategories |
| PUT | /api/v1/categories/:code | Bearer | canManageCategories |
| DELETE | /api/v1/categories/:code | Bearer | canManageCategories |

### POST /api/v1/categories
**Request**
```json
{ "name": "string", "display_name": "string", "description": "string", "status": "Active|Inactive" }
```
**Response 201** — Category object

### GET /api/v1/categories
**Query params:** `search?`, `status?`  
**Response 200**
```json
{ "data": [Category], "total": 25 }
```

### GET /api/v1/categories/:code
**Response 200** — Category

### PUT /api/v1/categories/:code
**Request**
```json
{ "display_name": "string", "description": "string", "status": "string" }
```
**Response 200** — Category

### DELETE /api/v1/categories/:code
**Response 204**

---

## Health & Metrics

| Method | Path | Auth | Permission |
|--------|------|------|------------|
| GET | /healthz | Public | IP allowlist |
| GET | /readyz | Public | IP allowlist |
| GET | /metrics | Public | IP allowlist |

### GET /healthz
**Response 200**
```json
{ "status": "healthy" }
```

### GET /readyz
**Response 200**
```json
{ "status": "ready", "database": "connected" }
```

### GET /metrics
**Response 200** — Prometheus exposition format (text/plain)

---

## Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 204 | No content (delete/update success) |
| 400 | Validation / parse error |
| 401 | Missing/invalid/expired token |
| 403 | Insufficient permissions |
| 404 | Not found |
| 409 | Conflict |
| 429 | Rate limited |
| 500 | Server error |

## Error Format
```json
{ "error": "string", "details": "string?" }
```

## Paginated Response
```json
{ "data": [...], "total": 42 }
```
Used by: categories, risk-assessments. All other list endpoints return raw arrays.

## Auth Header
`Authorization: Bearer <token>`  
Prefix `Bearer ` is case-insensitive.
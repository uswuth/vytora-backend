# VRMP API — Clear Reference

> **Base URL:** `http://localhost:8080/api/v1`  
> **Auth:** Add `Authorization: Bearer <token>` header to every request (except Health & Login)  
> **Steps:** Login → copy `token` from response → paste in every request as `Bearer <token>`

---

## How to Use the Token

1. Send `POST /api/v1/login` with admin credentials
2. Copy the `"token"` value from the JSON response
3. In every other request, add this header:
   ```
   Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
   ```
4. Keep using the same token until it expires

---

## Role — Permission Mapping

| Role | Has Permissions |
|------|----------------|
| **system_admin** | Everything |
| **risk_manager** | Create/List/Get/Update/Delete/Approve risk assessments, Generate reports, View categories |
| **compliance_officer** | Full CRUD compliance records, View categories |
| **department_manager** | Submit vendor for review, View assigned vendors, View categories |
| **auditor** | View audit logs, View reports, View categories |

---

## 1. Health

| Method | URL | Auth | Permission | Roles |
|--------|-----|------|------------|-------|
| `GET` | `/health` | ❌ None | Public | All (no auth) |

**Body:** None

---

## 2. Auth

### POST /api/v1/login

| Auth | Permission | Roles |
|------|------------|-------|
| ❌ None | Public | All |

```json
{
  "email": "admin@vrmp.com",
  "password": "admin123"
}
```

> **Response:** `{ "token": "eyJ...", "user": { ... } }`  
> Copy the `token` value and use in `Authorization: Bearer <token>` for all other requests.

### GET /api/v1/me

| Auth | Permission | Roles |
|------|------------|-------|
| ✅ Bearer token | None (any authenticated user) | All |

**Body:** None

---

## 3. Users

All user endpoints are behind the **`canManageUsers`** permission.  
**Only `system_admin` has this permission.**

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `POST` | `/api/v1/users` | canManageUsers | system_admin |
| `GET` | `/api/v1/users` | canManageUsers | system_admin |
| `GET` | `/api/v1/users/{id}` | canManageUsers | system_admin |
| `PUT` | `/api/v1/users/{id}/role` | canManageUsers | system_admin |
| `PUT` | `/api/v1/users/{id}/deactivate` | canManageUsers | system_admin |
| `PUT` | `/api/v1/users/{id}/activate` | canManageUsers | system_admin |

**Request body templates:**

### POST /api/v1/users
```json
{
  "email": "risk.manager@vrmp.com",
  "password": "password123",
  "full_name": "Risk Manager",
  "role": "risk_manager"
}
```
> Roles: `system_admin`, `risk_manager`, `compliance_officer`, `department_manager`, `auditor`

### POST /api/v1/users (another role)
```json
{
  "email": "compliance@vrmp.com",
  "password": "password123",
  "full_name": "Compliance Officer",
  "role": "compliance_officer"
}
```

### PUT /api/v1/users/{id}/role
```json
{
  "role": "risk_manager"
}
```

### PUT /api/v1/users/{id}/deactivate
**Body:** None

### PUT /api/v1/users/{id}/activate
**Body:** None

---

## 4. Categories

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `POST` | `/api/v1/categories` | canManageCategories | system_admin |
| `PUT` | `/api/v1/categories/{code}` | canManageCategories | system_admin |
| `DELETE` | `/api/v1/categories/{code}` | canManageCategories | system_admin |
| `GET` | `/api/v1/categories` | canViewCategories | system_admin, risk_manager, compliance_officer, department_manager, auditor |
| `GET` | `/api/v1/categories/{code}` | canViewCategories | system_admin, risk_manager, compliance_officer, department_manager, auditor |

### POST /api/v1/categories
```json
{
  "name": "technology",
  "display_name": "Technology",
  "description": "Technology and software vendors",
  "status": "Active"
}
```
> Status: `Draft`, `Active`, `Inactive`

### PUT /api/v1/categories/{code}
```json
{
  "display_name": "Technology Vendors",
  "description": "Updated description",
  "status": "Active"
}
```

---

## 5. Vendors

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `POST` | `/api/v1/vendors` | canCreateVendor | system_admin |
| `GET` | `/api/v1/vendors` | canEditVendor | system_admin |
| `GET` | `/api/v1/vendors/{code}` | canEditVendor | system_admin |
| `PUT` | `/api/v1/vendors/{code}` | canEditVendor | system_admin |
| `DELETE` | `/api/v1/vendors/{code}` | canDeleteVendor | system_admin |

### POST /api/v1/vendors
```json
{
  "name": "Acme Corporation",
  "category": "technology",
  "contact_person": "John Doe",
  "contact_email": "john@acme.com",
  "country": "USA",
  "risk_level": "Medium"
}
```
> Risk levels: `Low`, `Medium`, `High`, `Critical`  
> Category must match an existing category `name`.

### PUT /api/v1/vendors/{code}
```json
{
  "name": "Acme Corporation Updated",
  "category": "technology",
  "contact_person": "John Doe Jr",
  "contact_email": "john.jr@acme.com",
  "country": "USA",
  "risk_level": "High",
  "status": "Draft"
}
```
> Status values: `Draft`, `Submitted`, `RiskReview`, `ComplianceReview`, `Approved`, `Rejected`, `Active`, `Inactive`

---

## 6. Vendor Workflow (Status Transitions)

| Method | URL | Transition | Permission | Roles |
|--------|-----|------------|------------|-------|
| `PUT` | `/api/v1/vendors/{code}/submit` | Draft → Submitted | canSubmitVendorRequest | system_admin, department_manager |
| `PUT` | `/api/v1/vendors/{code}/review-risk` | Submitted → RiskReview | canReviewRisk | system_admin, risk_manager |
| `PUT` | `/api/v1/vendors/{code}/review-compliance` | RiskReview → ComplianceReview | canReviewCompliance | system_admin, compliance_officer |
| `PUT` | `/api/v1/vendors/{code}/approve` | ComplianceReview → Approved | canEditVendor | system_admin |
| `PUT` | `/api/v1/vendors/{code}/reject` | Any → Rejected | canEditVendor | system_admin |

**Body:** None for all workflow endpoints.

---

## 7. Risk Assessments

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `POST` | `/api/v1/risk-assessments` | canCreateRiskAssessment | system_admin, risk_manager |
| `GET` | `/api/v1/risk-assessments` | canReviewRisk | system_admin, risk_manager |
| `GET` | `/api/v1/risk-assessments/{code}` | canReviewRisk | system_admin, risk_manager |
| `PUT` | `/api/v1/risk-assessments/{code}` | canReviewRisk | system_admin, risk_manager |
| `DELETE` | `/api/v1/risk-assessments/{code}` | canReviewRisk | system_admin, risk_manager |
| `PUT` | `/api/v1/risk-assessments/{code}/approve` | canApproveRisk | system_admin, risk_manager |

### POST /api/v1/risk-assessments
```json
{
  "vendor_code": "VEN001",
  "assessment_date": "2026-06-14",
  "overall_risk_score": 65,
  "risk_level": "Medium",
  "security_risk_score": 70,
  "financial_risk_score": 60,
  "operational_risk_score": 50,
  "legal_risk_score": 80,
  "notes": "Initial risk assessment review"
}
```
> Scores must be 0–100. Risk level: `Low`, `Medium`, `High`, `Critical`

### PUT /api/v1/risk-assessments/{code}
```json
{
  "assessment_date": "2026-06-15",
  "overall_risk_score": 45,
  "risk_level": "Low",
  "security_risk_score": 40,
  "financial_risk_score": 50,
  "operational_risk_score": 30,
  "legal_risk_score": 60,
  "status": "Reviewed",
  "notes": "Updated after mitigation review"
}
```
> Status: `Draft`, `Reviewed`, `Approved`

### GET /api/v1/risk-assessments?vendor_code=VEN001
**Query params:** `vendor_code`, `risk_level`, `status`, `limit`, `offset`

---

## 8. Compliance Records

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `POST` | `/api/v1/compliance` | canReviewCompliance | system_admin, compliance_officer |
| `GET` | `/api/v1/compliance` | canReviewCompliance | system_admin, compliance_officer |
| `GET` | `/api/v1/compliance/{code}` | canReviewCompliance | system_admin, compliance_officer |
| `PUT` | `/api/v1/compliance/{code}` | canReviewCompliance | system_admin, compliance_officer |
| `DELETE` | `/api/v1/compliance/{code}` | canReviewCompliance | system_admin, compliance_officer |
| `GET` | `/api/v1/compliance/expiring` | canReviewCompliance | system_admin, compliance_officer |

### POST /api/v1/compliance
```json
{
  "vendor_code": "VEN001",
  "certification_type": "ISO27001",
  "valid_from": "2026-01-01",
  "valid_until": "2026-12-31",
  "issued_by": "BSI Group",
  "evidence_url": "https://storage.example.com/certs/iso27001.pdf"
}
```
> Cert types: `ISO27001`, `SOC2`, `GDPR`, `PCI_DSS`

### PUT /api/v1/compliance/{code}
```json
{
  "certification_type": "ISO27001",
  "status": "Approved",
  "valid_from": "2026-01-01",
  "valid_until": "2026-12-31",
  "issued_by": "BSI Group",
  "evidence_url": "https://storage.example.com/certs/iso27001.pdf"
}
```
> Status: `Pending`, `Approved`, `Expired`

### GET /api/v1/compliance?vendor_code=VEN001
**Query param:** `vendor_code` (required)

### GET /api/v1/compliance/expiring?days=90
**Query param:** `days` (optional, default 30)

---

## 9. Contracts

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `POST` | `/api/v1/contracts` | canEditVendor | system_admin |
| `GET` | `/api/v1/contracts` | canEditVendor | system_admin |
| `GET` | `/api/v1/contracts/{code}` | canEditVendor | system_admin |
| `PUT` | `/api/v1/contracts/{code}` | canEditVendor | system_admin |
| `DELETE` | `/api/v1/contracts/{code}` | canEditVendor | system_admin |
| `GET` | `/api/v1/contracts/expiring` | canEditVendor | system_admin |

### POST /api/v1/contracts
```json
{
  "vendor_code": "VEN001",
  "contract_number": "CN-2026-001",
  "start_date": "2026-01-01",
  "end_date": "2026-12-31",
  "contract_value": 500000,
  "renewal_status": "Auto-Renew"
}
```
> Renewal status: `Auto-Renew`, `Manual`, `Expiring`

### PUT /api/v1/contracts/{code}
```json
{
  "vendor_code": "VEN001",
  "contract_number": "CN-2026-001-UPDATED",
  "start_date": "2026-01-01",
  "end_date": "2027-06-30",
  "contract_value": 750000,
  "renewal_status": "Manual"
}
```

### GET /api/v1/contracts?vendor_code=VEN001
**Query param:** `vendor_code` (required)

### GET /api/v1/contracts/expiring?days=90
**Query param:** `days` (optional, default 30)

---

## 10. Audit Trail

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `GET` | `/api/v1/audit` | canViewAuditHistory | system_admin, auditor |

### GET /api/v1/audit
**Query params:** `table`, `record_code`, `action`, `changed_by`, `date_from`, `date_to`, `limit`, `offset` (all optional)

---

## 11. Reports

| Method | URL | Permission | Roles |
|--------|-----|------------|-------|
| `GET` | `/api/v1/reports/summary` | canAccessAllReports | system_admin, auditor |
| `GET` | `/api/v1/reports/monthly-onboarding` | canAccessAllReports | system_admin, auditor |
| `GET` | `/api/v1/reports/summary-2` | canViewAssignedVendors | department_manager |
| `GET` | `/api/v1/reports/monthly-onboarding-2` | canViewAssignedVendors | department_manager |

**Body:** None for all report endpoints.

---

## Quick Test Flow (All Entities in 2 Minutes)

```
1. POST /api/v1/login                        → copy token
2. POST /api/v1/categories                   → copy code
3. POST /api/v1/vendors                      → copy code
4. PUT  /api/v1/vendors/{code}/submit
5. PUT  /api/v1/vendors/{code}/review-risk
6. PUT  /api/v1/vendors/{code}/review-compliance
7. PUT  /api/v1/vendors/{code}/approve
8. POST /api/v1/risk-assessments             → copy code
9. POST /api/v1/compliance                   → copy code
10. POST /api/v1/contracts                   → copy code
11. GET  /api/v1/audit
12. GET  /api/v1/reports/summary
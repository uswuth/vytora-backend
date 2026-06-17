VRMP Backend API Documentation v1.0

Base URL: http://localhost:8080
Content-Type: application/json (for requests with body)
Authentication: Bearer token (JWT) in Authorization header for protected endpoints.
1. Authentication
1.1 Login

Endpoint
POST /api/v1/login

Purpose
Authenticate a user and receive a JWT token.

Authentication
None (public).

Request Body
Field	Type	Required	Validation	Description
email	string	Yes	Valid email	User's email
password	string	Yes	Min 6 characters	User's password

Sample Request
json

{
  "email": "admin@vrmp.com",
  "password": "admin123"
}

Success Response (200)
json

{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "code": "USR001",
    "email": "admin@vrmp.com",
    "full_name": "System Administrator",
    "role": "system_admin",
    "is_active": true,
    "created_at": "2026-06-14T03:08:31+05:30",
    "updated_at": "2026-06-14T03:08:31+05:30"
  }
}

Error Responses

    400 – Invalid request body

    401 – {"error":"invalid email or password"}

    403 – {"error":"account is deactivated"}

2. User Profile
2.1 Get Current User

Endpoint
GET /api/v1/me

Purpose
Return authenticated user’s details from the JWT.

Authentication
Required – any valid role.

Headers
Authorization: Bearer <token>

Response
json

{
  "user_id": "uuid",
  "code": "USR001",
  "email": "admin@vrmp.com",
  "role": "system_admin"
}

Error Responses

    401 – Missing/invalid token

3. Vendors
3.1 Create Vendor

Endpoint
POST /api/v1/vendors

Purpose
Add a new vendor (status = Draft).

Authentication
canCreateVendor (system_admin, risk_manager, department_manager)

Request Body
Field	Type	Required	Validation	Description
name	string	Yes	2-255 characters	Vendor name
category	string	Yes	Max 100 characters	e.g., "Cloud", "Payroll"
contact_person	string	No	–	Name of contact person
contact_email	string	No	Valid email (if present)	Contact email
country	string	No	–	Country
risk_level	string	Yes	One of: Low, Medium, High, Critical	Initial risk level

Sample Request
json

{
  "name": "Acme Corp",
  "category": "Cloud",
  "contact_person": "John Doe",
  "contact_email": "john@acme.com",
  "country": "India",
  "risk_level": "Low"
}

Success Response (201)
json

{
  "id": "uuid",
  "code": "VEN001",
  "name": "Acme Corp",
  "category": "Cloud",
  "contact_person": "John Doe",
  "contact_email": "john@acme.com",
  "country": "India",
  "contract_start_date": null,
  "contract_end_date": null,
  "risk_level": "Low",
  "status": "Draft",
  "assigned_dept_manager_id": null,
  "created_by": "uuid",
  "created_at": "2026-06-17T22:11:53+05:30",
  "updated_at": "2026-06-17T22:11:53+05:30"
}

Error Responses

    400 – Validation error

    401 – Token missing/invalid

    403 – No permission

    500 – Server error

3.2 List Vendors

Endpoint
GET /api/v1/vendors

Purpose
Search and paginate vendors with filters.

Authentication
canEditVendor (system_admin, risk_manager, department_manager)

Query Parameters
Parameter	Type	Required	Description
search	string	No	Search in name or code (case‑insensitive)
category	string	No	Exact match
risk_level	string	No	Exact match (Low/Medium/High/Critical)
status	string	No	Exact match (Draft/Submitted/…)
country	string	No	Exact match
limit	integer	No	Rows per page (default 20)
offset	integer	No	Offset for pagination
sort_by	string	No	Column name (e.g., created_at)
sort_order	string	No	ASC or DESC (default DESC)

Sample Request
GET /api/v1/vendors?search=Acme&risk_level=Low&limit=10

Response (200)
json

{
  "data": [
    {
      "id": "uuid",
      "code": "VEN001",
      "name": "Acme Corp",
      "category": "Cloud",
      "contact_person": "John Doe",
      "contact_email": "john@acme.com",
      "country": "India",
      "contract_start_date": null,
      "contract_end_date": null,
      "risk_level": "Low",
      "status": "Draft",
      "assigned_dept_manager_id": null,
      "created_by": "uuid",
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "total": 1
}

Error Responses

    401, 403

3.3 Get Vendor by Code

Endpoint
GET /api/v1/vendors/{code}

Purpose
Retrieve a single vendor by its human‑readable code (e.g., VEN001).

Authentication
canEditVendor

Response (200)
Same vendor object as above.

Error Responses

    404 – {"error":"vendor not found"}

3.4 Update Vendor

Endpoint
PUT /api/v1/vendors/{code}

Purpose
Modify vendor details.

Authentication
canEditVendor

Request Body
Same fields as create, plus status (required) – can only change to valid workflow states.
Field	Type	Required	Validation
status	string	Yes	One of: Draft, Submitted, RiskReview, ComplianceReview, Approved, Rejected, Active, Inactive

Sample Request
json

{
  "name": "Acme Corp Updated",
  "category": "Finance",
  "contact_person": "Jane Doe",
  "contact_email": "jane@acme.com",
  "country": "India",
  "risk_level": "Medium",
  "status": "Draft"
}

Response (200)
Updated vendor object.

Error Responses

    400 – Validation

    404 – Vendor not found

    403

3.5 Delete Vendor

Endpoint
DELETE /api/v1/vendors/{code}

Purpose
Remove a vendor permanently.

Authentication
canDeleteVendor

Response
204 No Content

Error Responses

    404

4. Vendor Workflow Transitions

These endpoints advance the vendor’s status.
4.1 Submit (Draft → Submitted)

Endpoint
PUT /api/v1/vendors/{code}/submit

Authentication
canSubmitVendorRequest

Response
204 No Content

Errors
400 – Not in Draft status
404
4.2 Risk Review (Submitted → RiskReview)

Endpoint
PUT /api/v1/vendors/{code}/review-risk

Authentication
canReviewRisk
4.3 Compliance Review (RiskReview → ComplianceReview)

Endpoint
PUT /api/v1/vendors/{code}/review-compliance

Authentication
canReviewCompliance
4.4 Approve (ComplianceReview → Approved)

Endpoint
PUT /api/v1/vendors/{code}/approve

Authentication
canEditVendor (admin role typically)
4.5 Reject (Any → Rejected)

Endpoint
PUT /api/v1/vendors/{code}/reject

Authentication
canEditVendor

Errors for all transitions
400 – Invalid current status
404 – Not found
403 – No permission
5. Risk Assessments
5.1 Create Risk Assessment

Endpoint
POST /api/v1/risk-assessments

Purpose
Submit a new risk assessment; overall risk score and level are auto‑calculated.

Authentication
canCreateRiskAssessment

Request Body
Field	Type	Required	Validation
vendor_code	string	Yes	Existing vendor code
security_risk_score	float	Yes	0‑100
financial_risk_score	float	Yes	0‑100
operational_risk_score	float	Yes	0‑100
legal_risk_score	float	Yes	0‑100
assessment_date	string	Yes	Format: YYYY-MM-DD
notes	string	No	Free text

Sample Request
json

{
  "vendor_code": "VEN001",
  "security_risk_score": 30,
  "financial_risk_score": 40,
  "operational_risk_score": 20,
  "legal_risk_score": 10,
  "assessment_date": "2026-06-17",
  "notes": "Initial review"
}

Response (201)
json

{
  "id": "uuid",
  "code": "RAK00001",
  "vendor_id": "uuid",
  "vendor_code": "VEN001",
  "assessment_date": "2026-06-17T00:00:00Z",
  "assessor_id": "uuid",
  "assessor_code": "USR001",
  "overall_risk_score": 25.0,
  "risk_level": "Low",
  "security_risk_score": 30,
  "financial_risk_score": 40,
  "operational_risk_score": 20,
  "legal_risk_score": 10,
  "status": "Draft",
  "notes": "Initial review",
  "created_at": "...",
  "updated_at": "..."
}

*(overall_score = average of four scores, risk_level mapping: ≤25 Low, 26‑50 Medium, 51‑75 High, ≥76 Critical)*

Errors
400 – Invalid date, vendor not found
403 – No permission
500
5.2 List Risk Assessments

Endpoint
GET /api/v1/risk-assessments

Authentication
canReviewRisk

Query Parameters
Parameter	Type	Description
vendor_code	string	Filter by vendor code
risk_level	string	Low/Medium/High/Critical
status	string	Draft/Reviewed/Approved
date_from	string	YYYY-MM-DD
date_to	string	YYYY-MM-DD
limit/offset	int	Pagination

Response (200)
json

{
  "data": [ /* array of risk assessment objects with vendor_code and assessor_code */ ],
  "total": 5
}

5.3 Get Risk Assessment

Endpoint
GET /api/v1/risk-assessments/{code}

Authentication
canReviewRisk

Response (200)
Single risk assessment object (with vendor_code and assessor_code).
5.4 Approve Risk Assessment

Endpoint
PUT /api/v1/risk-assessments/{code}/approve

Authentication
canApproveRisk

Response
204 No Content

Errors
400 – Only Draft/Reviewed can be approved
6. Compliance Records
6.1 Create Compliance Record

Endpoint
POST /api/v1/compliance

Authentication
canReviewCompliance

Request Body
Field	Type	Required	Validation
vendor_code	string	Yes	Existing vendor
certification_type	string	Yes	ISO27001, SOC2, GDPR, PCI_DSS
valid_from	string	Yes	YYYY-MM-DD
valid_until	string	Yes	YYYY-MM-DD
issued_by	string	No	Issuing authority
evidence_url	string	No	URL to evidence file

Sample Request
json

{
  "vendor_code": "VEN001",
  "certification_type": "ISO27001",
  "valid_from": "2026-01-01",
  "valid_until": "2026-12-31",
  "issued_by": "Certifier Inc",
  "evidence_url": "https://files.example.com/iso.pdf"
}

Response (201)
json

{
  "id": "uuid",
  "code": "CMP00001",
  "vendor_id": "uuid",
  "vendor_code": "VEN001",
  "certification_type": "ISO27001",
  "status": "Approved",         // auto‑computed: Approved/Expired/Pending
  "valid_from": "2026-01-01T00:00:00Z",
  "valid_until": "2026-12-31T00:00:00Z",
  "issued_by": "Certifier Inc",
  "evidence_url": "https://files.example.com/iso.pdf",
  "reviewed_by": "uuid",
  "created_at": "...",
  "updated_at": "..."
}

Status: Approved if today between valid_from and valid_until, Expired if after valid_until, else Pending.

Errors
409 – Duplicate certification type for this vendor (unique vendor+certification+status)
404 – Vendor not found
400 – Validation
6.2 List Compliance Records

Endpoint
GET /api/v1/compliance?vendor_code=VEN001

Authentication
canReviewCompliance

Response (200)
json

[
  { /* compliance record objects with vendor_code */ }
]

6.3 Get Compliance Record

Endpoint
GET /api/v1/compliance/{code}

Authentication
canReviewCompliance

Response
Single compliance record.
6.4 Update Compliance Record

Endpoint
PUT /api/v1/compliance/{code}

Authentication
canReviewCompliance

Request Body
Same fields as create, but all optional; only provided fields are updated. Can also update status manually if needed.
6.5 Expiring Certifications

Endpoint
GET /api/v1/compliance/expiring?days=90

Authentication
Any authenticated user.

Response (200)
Array of compliance records that expire within the given number of days (status=Approved and valid_until ≤ today+days).
7. Contracts
7.1 Create Contract

Endpoint
POST /api/v1/contracts

Authentication
canEditVendor

Request Body
Field	Type	Required	Validation
vendor_code	string	Yes	Existing vendor
contract_number	string	Yes	Max 100 characters
start_date	string	Yes	YYYY-MM-DD
end_date	string	Yes	YYYY-MM-DD
contract_value	float	No	Monetary value
renewal_status	string	Yes	Auto-Renew, Manual, Expiring

Sample Request
json

{
  "vendor_code": "VEN001",
  "contract_number": "CT-001",
  "start_date": "2026-01-01",
  "end_date": "2027-01-01",
  "contract_value": 500000,
  "renewal_status": "Manual"
}

Response (201)
json

{
  "id": "uuid",
  "code": "CTR00001",
  "vendor_id": "uuid",
  "vendor_code": "VEN001",
  "contract_number": "CT-001",
  "start_date": "2026-01-01T00:00:00Z",
  "end_date": "2027-01-01T00:00:00Z",
  "contract_value": 500000,
  "renewal_status": "Manual",
  "created_at": "...",
  "updated_at": "..."
}

7.2 List Contracts

Endpoint
GET /api/v1/contracts?vendor_code=VEN001

Authentication
canEditVendor
7.3 Get Contract

Endpoint
GET /api/v1/contracts/{code}

Authentication
canEditVendor
7.4 Expiring Contracts

Endpoint
GET /api/v1/contracts/expiring?days=90

Authentication
Any authenticated user.
8. Audit Trail
8.1 List Audit Entries

Endpoint
GET /api/v1/audit

Purpose
Query immutable record of all changes.

Authentication
canViewAuditHistory (system_admin, auditor)

Query Parameters
Parameter	Type	Description
table	string	vendors, risk_assessments, compliance_records, contracts
record_code	string	Entity code (e.g., VEN001)
action	string	CREATE, UPDATE, DELETE
changed_by	string	User code (e.g., USR001)
date_from	string	YYYY-MM-DD
date_to	string	YYYY-MM-DD
limit/offset	int	Pagination

Response (200)
json

{
  "data": [
    {
      "id": "uuid",
      "code": "AUD000001",
      "table_name": "vendors",
      "record_id": "uuid",
      "record_code": "VEN001",
      "action": "UPDATE",
      "field_name": "status",
      "old_value": "Draft",
      "new_value": "Submitted",
      "changed_by": "uuid",
      "changed_at": "2026-06-17T22:11:53+05:30"
    }
  ],
  "total": 1
}

9. Reports / Dashboard
9.1 Summary

Endpoint
GET /api/v1/reports/summary

Authentication
canAccessAllReports (admin, risk manager, compliance officer) or canViewAssignedVendors (department manager – sees only their vendors)

Response (200)
json

{
  "total_vendors": 10,
  "vendors_by_status": {
    "Draft": 4,
    "Submitted": 2,
    "Approved": 4
  },
  "vendors_by_risk_level": {
    "Low": 5,
    "Medium": 3,
    "High": 2
  },
  "expiring_contracts_30_days": 1,
  "expiring_contracts_60_days": 2,
  "expiring_contracts_90_days": 3,
  "expiring_compliance_30_days": 0,
  "expiring_compliance_60_days": 1,
  "expiring_compliance_90_days": 2,
  "pending_risk_assessments": 2,
  "approved_risk_assessments": 5
}

9.2 Monthly Onboarding Trend

Endpoint
GET /api/v1/reports/monthly-onboarding

Authentication
Same as summary.

Response (200)
json

[
  { "month": "2026-05", "count": 3 },
  { "month": "2026-06", "count": 7 }
]

10. Health Check

Endpoint
GET /health

Purpose
Check if the server and database are alive.

Response

    200 – {"status":"healthy"}

    503 – {"status":"unhealthy","error":"database unreachable"}

General Notes

    All timestamps are in ISO 8601 format with timezone (e.g., 2026-06-17T22:11:53+05:30).

    All IDs are UUID v4 strings.

    Entity codes are human‑readable: VEN001, RAK00001, AUD000001, etc.

    Error responses follow the format {"error": "message"}.

    Rate limit: 100 requests/minute per IP. Returns 429 Too Many Requests if exceeded.

    For bulk operations, frontend should respect pagination parameters (limit/offset).

This document gives your frontend team everything they need to start building immediately. If any endpoint requires further clarification, let me know!

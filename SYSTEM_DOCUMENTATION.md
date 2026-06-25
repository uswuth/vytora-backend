# Vytora Vendor Risk Management Platform (VRMP)

## What is VRMP?

VRMP stands for **Vendor Risk Management Platform**. It is a software system designed to help organizations manage, assess, and monitor the risks associated with third-party vendors and suppliers. Every company works with external vendors — from IT service providers and cloud platforms to cleaning services and food suppliers. Each of these vendors introduces potential risks: security breaches, compliance failures, financial instability, or operational disruptions.

VRMP provides a structured, auditable, and automated way to onboard vendors, assess their risk, track compliance certifications, manage contracts, and generate reports for leadership and auditors.

---

## What Does This Web App Do?

This backend powers the VRMP web application. It provides:

- **User Authentication & Authorization** — Role-based access so different people see and do different things.
- **Vendor Onboarding & Lifecycle Management** — From initial registration to approval, rejection, activation, or deactivation.
- **Risk Assessment** — Score vendors across security, financial, operational, and legal dimensions.
- **Compliance Tracking** — Monitor certifications like ISO 27001, SOC 2, GDPR, and PCI DSS with expiry alerts.
- **Contract Management** — Track contract dates, values, and renewal statuses with expiry warnings.
- **Audit Trail** — Every significant change is logged with who changed it, when, and what changed.
- **Reporting & Analytics** — Executive dashboards showing vendor distributions, risk levels, onboarding trends, and compliance health.
- **Contact Management** — Store multiple contacts per vendor for coordinated communication.

Everything is exposed via a modern **GraphQL API**, allowing frontend applications to request exactly the data they need in a single round-trip.

---

## Who Is This Useful For?

| Role | Why They Use VRMP |
|------|-------------------|
| **System Admin** | Manages users, roles, and system configuration. Full access. |
| **Risk Manager** | Evaluates vendor risk scores, approves or escalates risk assessments. |
| **Compliance Officer** | Reviews and maintains compliance records, ensures vendors hold valid certifications. |
| **Department Manager** | Owns specific vendor relationships, reviews vendors assigned to their department. |
| **Auditor** | Reviews audit trails and compliance evidence for regulatory or internal audits. |

---

## System Entities & Their Purpose

### 1. Users
**What it is:** The people who access the system.

**Why it exists:** Every action in VRMP must be traceable to a person. Users have roles that determine what they can see and do.

**Key fields:**
- `role` — Determines permissions (system_admin, risk_manager, compliance_officer, department_manager, auditor)
- `is_active` — Soft-delete flag; deactivated users cannot log in

**What users can do:**
- Authenticate (login) and receive a JWT token
- View their own profile
- (Admins) Create, deactivate, activate users and assign roles

---

### 2. Categories
**What it is:** A structured classification system for vendors (e.g., "Technology", "Logistics", "Healthcare").

**Why it exists:** Organizations need to group vendors by type for filtering, reporting, and policy enforcement.

**Key fields:**
- `name` + `display_name` — Internal code vs. human-readable label
- `status` — Draft/Active/Inactive lifecycle

**What users can do:**
- Create, update, and deactivate categories
- Search and filter categories

---

### 3. Vendors
**What it is:** The core entity — a third-party company or service provider being managed.

**Why it exists:** VRMP's entire purpose is to track and manage vendor-related risk. Every other entity (contacts, assessments, contracts, compliance records) belongs to a vendor.

**Key fields:**
- `risk_level` — Low / Medium / High / Critical (drives prioritization)
- `status` — Draft → Submitted → RiskReview → ComplianceReview → Approved / Rejected → Active / Inactive
- `assigned_dept_manager_id` — Which department manager owns this vendor

**State machine:**
```
Draft → Submitted (by creator)
       → RiskReview (risk manager reviews)
       → ComplianceReview (compliance officer reviews)
       → Approved or Rejected
       → Active or Inactive
```

**What users can do:**
- Create vendor profiles with contact details
- Update vendor information
- Submit vendor for review
- Review risk and compliance
- Approve or reject vendors
- Filter, search, and sort vendors

---

### 4. Vendor Contacts
**What it is:** Individual people associated with a vendor (e.g., account manager, technical contact, billing contact).

**Why it exists:** A vendor organization has multiple stakeholders. Storing contacts ensures the right people are reached for escalations, questions, and communication.

**Key fields:**
- `vendor_id` — Links to parent vendor
- `name`, `email`, `phone` — Contact details

**What users can do:**
- Add, update, or remove contacts for a vendor

---

### 5. Risk Assessments
**What it is:** A structured evaluation of a vendor's risk profile across four dimensions.

**Why it exists:** Risk must be quantified, not subjective. This entity provides a repeatable, auditable scoring method.

**Scoring dimensions:**
| Dimension | Range | What it measures |
|-----------|-------|-----------------|
| Security Risk Score | 0–100 | Data breaches, vulnerabilities, access controls |
| Financial Risk Score | 0–100 | Credit rating, revenue stability, litigation history |
| Operational Risk Score | 0–100 | SLA breaches, downtime history, support quality |
| Legal Risk Score | 0–100 | Regulatory penalties, litigation, governance issues |

**Overall score** drives the `risk_level` label (Low/Medium/High/Critical).

**Key fields:**
- `assessment_date` — When the assessment was performed
- `assessor_id` — Who performed it
- `status` — Draft / Reviewed / Approved (workflow states)

**What users can do:**
- Create risk assessments for vendors
- Update assessments (during draft phase)
- Review and approve assessments
- Delete draft assessments
- Filter by vendor, risk level, and status

---

### 6. Compliance Records
**What it is:** Tracks whether a vendor holds required certifications and whether they are current.

**Why it exists:** Many industries require vendors to maintain specific certifications (ISO 27001 for data security, SOC 2 for cloud services, GDPR for EU data handling, PCI DSS for payment processing). Expired certifications are a compliance risk.

**Certification types supported:**
- **ISO27001** — Information Security Management
- **SOC2** — Service Organization Control 2 (trust & security)
- **GDPR** — EU Data Protection compliance
- **PCI_DSS** — Payment Card Industry Data Security Standard

**Key fields:**
- `valid_from` / `valid_until` — Certification validity window
- `status` — Pending / Approved / Expired
- `evidence_url` — Link to the actual certificate document
- `reviewed_by` — Compliance officer who validated it

**Constraint:** Unique combination of (vendor_id, certification_type, status) — prevents duplicate pending/approved certs for the same vendor.

**What users can do:**
- Create compliance records with evidence URLs
- Update records (renew certifications)
- Delete records
- View expiring certifications (with configurable warning window: 30/60/90 days)

---

### 7. Contracts
**What it is:** Formal agreement documents between the organization and the vendor.

**Why it exists:** Contracts define legal terms, financial obligations, SLAs, and renewal conditions. Expired or missing contracts create operational and legal risk.

**Key fields:**
- `contract_number` — External reference (e.g., CNT-2025-001)
- `start_date` / `end_date` — Contract validity period
- `contract_value` — Monetary value for financial reporting
- `renewal_status` — Auto_Renew / Manual / Expiring

**What users can do:**
- Create, update, and delete contracts
- View contracts expiring within a configurable window (30/60/90 days)
- Track contract values for budget forecasting

---

### 8. Audit Trail
**What it is:** An immutable log of every significant action taken in the system.

**Why it exists:** Regulators (like GDPR, SOX) and internal auditors require proof of who changed what and when. The audit trail provides this evidence.

**What gets logged:**
- CREATE, UPDATE, DELETE actions
- Which table and record was affected
- The specific field that changed
- Old value → New value
- Who made the change and when

**Example entries:**
- "User X changed vendor VEN-001 status from Draft to Submitted"
- "User Y updated risk assessment RA-003 overall score from 60 to 85"
- "User Z deleted contract CNT-2024-015"

**What users can do:**
- Search and filter logs by table, record, action type, user, and date range
- View the complete history of any vendor, assessment, or contract

---

### 9. Reports (Read-Only Analytics)
**What it is:** Pre-built analytical views that aggregate data across all entities for decision-makers.

**Why it exists:** Executives and managers need answers fast — "How many vendors are high risk?", "Which certifications expire soon?", "How many vendors did we onboard this month?"

**Available reports:**

| Report | What it Shows |
|--------|--------------|
| **Summary Report** | Total vendors, breakdown by status/risk level, expiring contracts & compliance counts, pending vs approved risk assessments |
| **Monthly Onboarding** | Trend of new vendors added per month |
| **High Risk Vendors** | Vendors flagged as High/Critical risk with their latest assessment status and expiring contract count |
| **Expiring Contracts** | Contracts ending soon, with days remaining and renewal status |
| **Compliance Summary** | Breakdown of certifications by type (approved/pending/expired counts) |
| **Time Series** | Vendor, risk assessment, and compliance creation trends over time |

---

## User Roles & Permissions Summary

| Permission | System Admin | Risk Manager | Compliance Officer | Dept. Manager | Auditor |
|------------|:---:|:---:|:---:|:---:|:---:|
| Manage Users | ✅ | | | | |
| Manage Categories | ✅ | | | | |
| Create Vendor | ✅ | ✅ | | | |
| Edit Vendor | ✅ | ✅ | | ✅ | |
| Delete Vendor | ✅ | ✅ | | | |
| Submit Vendor for Review | ✅ | | | | |
| Review Risk Assessment | | ✅ | | | |
| Approve Risk Assessment | | ✅ | | | |
| Manage Compliance Records | | | ✅ | | |
| View Audit History | | | | | ✅ |
| Access All Reports | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## How the Pieces Connect

```
Users (with roles)
  ├── own Vendors (created_by)
  ├── manage Vendors (assigned_dept_manager_id)
  ├── perform Risk Assessments (assessor_id)
  ├── review Compliance Records (reviewed_by)
  ├── change categories (created_by / updated_by)
  └── trigger AuditTrail entries (changed_by)

Vendors
  ├── have many VendorContacts
  ├── have many RiskAssessments
  ├── have many ComplianceRecords
  ├── have many Contracts
  └── belong to a Category

Everything → feeds into Reports
Everything → is logged in AuditTrail
```

---

## Authentication Flow

1. User sends `login(email, password)` mutation
2. System validates credentials against stored `password_hash`
3. System returns a JWT token and user profile
4. For subsequent requests, client sends `Authorization: Bearer <token>` header
5. Middleware validates the token, extracts claims (user ID, role)
6. Directive-based authorization (`@isAuthenticated`, `@hasPermission`) protects fields and mutations

---

## Technology Stack

- **Language:** Go (Golang)
- **API Style:** GraphQL (gqlgen)
- **Router:** Chi
- **Database:** PostgreSQL
- **Auth:** JWT tokens
- **Metrics:** Prometheus
- **Migrations:** Raw SQL files

---

## Quick Start

1. Start the server: the app listens on `0.0.0.0:8080`
2. Open browser to `http://localhost:8080/` for GraphQL Playground
3. Login with your first user (create via seed script or direct DB insert)
4. Use the token from login response for all subsequent authenticated requests

For ready-to-use queries and mutations, see `graphql_playground_queries.graphql`.
For database diagram, see `dbdiagram.dsl` and paste into https://dbdiagram.io/.
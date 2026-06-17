# API Endpoints

## 1. Health

<table class="endpoint-table" style="width: 100%; border-collapse: collapse; border: 1px solid #ddd;">
  <thead>
    <tr style="background-color: #f0f0f0; font-weight: bold;">
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Method & Path</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Purpose</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Auth</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Request</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Response (Success)</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Errors</th>
    </tr>
  </thead>
  <tbody>
    <tr style="background-color: #fafafa;">
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><code>GET /health</code></td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">Check API health and database connectivity</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">Public</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">None</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><code>200</code> - <code>{"status":"healthy"}</code></td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><span style="background-color: #fff3cd;">503</span> – database unreachable</td>
    </tr>
  </tbody>
</table>

## 2. Authentication

<table class="endpoint-table" style="width: 100%; border-collapse: collapse; border: 1px solid #ddd;">
  <thead>
    <tr style="background-color: #f0f0f0; font-weight: bold;">
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Method & Path</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Purpose</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Auth</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Request</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Response (Success)</th>
      <th style="padding: 8px 10px; border: 1px solid #ddd;">Errors</th>
    </tr>
  </thead>
  <tbody>
    <tr style="background-color: #fafafa;">
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><code>POST /api/v1/login</code></td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">Authenticate and receive JWT</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">Public</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">
        <ul style="margin: 0; padding-left: 20px;">
          <li><code>email</code> string ✅</li>
          <li><code>password</code> string ✅</li>
        </ul>
      </td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><code>200</code> - <code>{"token": string, "user": User}</code></td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">
        <span style="background-color: #f8d7da;">400</span> – validation failed<br>
        <span style="background-color: #f8d7da;">401</span> – invalid credentials<br>
        <span style="background-color: #f8d7da;">403</span> – account deactivated
      </td>
    </tr>
    <tr style="background-color: #fff;">
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><code>GET /api/v1/me</code></td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">Get current authenticated user</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">Authenticated</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;">None</td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><code>200</code> - <code>{"user_id": string, "code": string, "email": string, "role": string}</code></td>
      <td style="padding: 8px 10px; border: 1px solid #ddd;"><span style="background-color: #f8d7da;">401</span> – missing/invalid token</td>
    </tr>
  </tbody>
</table>

## 3. Users

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>POST /api/v1/users</code> | Create a new user | <code>canManageUsers</code> | <ul style="margin:0;padding-left:20px;"><li><code>email</code> string ✅</li><li><code>password</code> string ✅</li><li><code>full_name</code> string ✅</li><li><code>role</code> string ✅ (system_admin / risk_manager / compliance_officer / department_manager / auditor)</li></ul> | <code>201</code> - User object (without password_hash) | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">409</span> – email exists<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/users</code> | List all users | <code>canManageUsers</code> | None | <code>200</code> - Array of User objects | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/users/{id}</code> | Get user by ID | <code>canManageUsers</code> | None | <code>200</code> - User object | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized |
| <code>PUT /api/v1/users/{id}/role</code> | Update user role | <code>canManageUsers</code> | <ul style="margin:0;padding-left:20px;"><li><code>role</code> string ✅</li></ul> | <code>204 No Content</code> | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>PUT /api/v1/users/{id}/deactivate</code> | Deactivate user | <code>canManageUsers</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>PUT /api/v1/users/{id}/activate</code> | Activate user | <code>canManageUsers</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |

## 4. Categories

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>POST /api/v1/categories</code> | Create category | <code>canManageCategories</code> | <ul style="margin:0;padding-left:20px;"><li><code>name</code> string ✅</li><li><code>display_name</code> string ✅</li><li><code>description</code> string ❌</li><li><code>status</code> string ✅ (Draft/Active/Inactive)</li></ul> | <code>201</code> - Category object | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">409</span> – duplicate name<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/categories</code> | List categories | <code>canViewCategories</code> | <ul style="margin:0;padding-left:20px;"><li><code>?search</code> string ❌</li><li><code>?status</code> string ❌</li><li><code>?limit</code> int ❌</li><li><code>?offset</code> int ❌</li></ul> | <code>200</code> - <code>{"data": Category[], "total": int}</code> | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/categories/{code}</code> | Get category by code | <code>canViewCategories</code> | None | <code>200</code> - Category object | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized |
| <code>PUT /api/v1/categories/{code}</code> | Update category | <code>canManageCategories</code> | <ul style="margin:0;padding-left:20px;"><li><code>display_name</code> string ✅</li><li><code>description</code> string ❌</li><li><code>status</code> string ✅ (Draft/Active/Inactive)</li></ul> | <code>200</code> - Category object | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>DELETE /api/v1/categories/{code}</code> | Delete category | <code>canManageCategories</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |

## 5. Vendors

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>POST /api/v1/vendors</code> | Create vendor | <code>canCreateVendor</code> | <ul style="margin:0;padding-left:20px;"><li><code>name</code> string ✅</li><li><code>category</code> string ✅</li><li><code>contact_person</code> string ❌</li><li><code>contact_email</code> string ❌</li><li><code>country</code> string ❌</li><li><code>risk_level</code> string ✅ (Low/Medium/High/Critical)</li></ul> | <code>201</code> - Vendor object | <span style="background-color:#f8d7da;">400</span> – validation failed / category invalid<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/vendors</code> | List vendors | <code>canEditVendor</code> | <ul style="margin:0;padding-left:20px;"><li><code>?search</code> string ❌</li><li><code>?category</code> string ❌</li><li><code>?risk_level</code> string ❌</li><li><code>?status</code> string ❌</li><li><code>?country</code> string ❌</li><li><code>?sort_by</code> string ❌</li><li><code>?sort_order</code> string ❌</li><li><code>?limit</code> int ❌</li><li><code>?offset</code> int ❌</li></ul> | <code>200</code> - <code>{"data": Vendor[], "total": int}</code> | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/vendors/{code}</code> | Get vendor by code | <code>canEditVendor</code> | None | <code>200</code> - Vendor object | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized |
| <code>PUT /api/v1/vendors/{code}</code> | Update vendor | <code>canEditVendor</code> | <ul style="margin:0;padding-left:20px;"><li><code>name</code> string ✅</li><li><code>category</code> string ✅</li><li><code>contact_person</code> string ❌</li><li><code>contact_email</code> string ❌</li><li><code>country</code> string ❌</li><li><code>risk_level</code> string ✅ (Low/Medium/High/Critical)</li><li><code>status</code> string ✅ (Draft/Submitted/RiskReview/ComplianceReview/Approved/Rejected/Active/Inactive)</li></ul> | <code>200</code> - Vendor object | <span style="background-color:#f8d7da;">400</span> – validation failed / category invalid<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>DELETE /api/v1/vendors/{code}</code> | Delete vendor | <code>canDeleteVendor</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |

### Workflow Transitions

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>PUT /api/v1/vendors/{code}/submit</code> | Submit vendor for review (Draft → Submitted) | <code>canSubmitVendorRequest</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">400</span> – must be Draft<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>PUT /api/v1/vendors/{code}/review-risk</code> | Move to risk review (Submitted → RiskReview) | <code>canReviewRisk</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">400</span> – must be Submitted<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>PUT /api/v1/vendors/{code}/review-compliance</code> | Move to compliance review (RiskReview → ComplianceReview) | <code>canReviewCompliance</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">400</span> – must be RiskReview<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>PUT /api/v1/vendors/{code}/approve</code> | Approve vendor (ComplianceReview → Approved) | <code>canEditVendor</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">400</span> – must be ComplianceReview<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>PUT /api/v1/vendors/{code}/reject</code> | Reject vendor (any → Rejected) | <code>canEditVendor</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">400</span> – already terminal state<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |

## 6. Risk Assessments

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>POST /api/v1/risk-assessments</code> | Create risk assessment | <code>canCreateRiskAssessment</code> | <ul style="margin:0;padding-left:20px;"><li><code>vendor_code</code> string ✅</li><li><code>assessment_date</code> string ✅ (YYYY-MM-DD)</li><li><code>risk_level</code> string ✅ (Low/Medium/High/Critical)</li><li><code>findings</code> string ✅</li><li><code>recommendations</code> string ❌</li></ul> | <code>201</code> - RiskAssessment object | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">404</span> – vendor not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/risk-assessments</code> | List risk assessments | <code>canReviewRisk</code> | <ul style="margin:0;padding-left:20px;"><li><code>?vendor_code</code> string ❌</li><li><code>?risk_level</code> string ❌</li><li><code>?status</code> string ❌</li><li><code>?limit</code> int ❌</li><li><code>?offset</code> int ❌</li></ul> | <code>200</code> - <code>{"data": RiskAssessment[], "total": int}</code> | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/risk-assessments/{code}</code> | Get risk assessment by code | <code>canReviewRisk</code> | None | <code>200</code> - RiskAssessment object | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized |
| <code>PUT /api/v1/risk-assessments/{code}/approve</code> | Approve risk assessment | <code>canApproveRisk</code> | None | <code>204 No Content</code> | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |

## 7. Compliance Records

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>POST /api/v1/compliance</code> | Create compliance record | <code>canReviewCompliance</code> | <ul style="margin:0;padding-left:20px;"><li><code>vendor_code</code> string ✅</li><li><code>certification_type</code> string ✅ (ISO27001/SOC2/GDPR/PCI_DSS)</li><li><code>valid_from</code> string ✅ (YYYY-MM-DD)</li><li><code>valid_until</code> string ✅ (YYYY-MM-DD)</li><li><code>issued_by</code> string ❌</li><li><code>evidence_url</code> string ❌</li></ul> | <code>201</code> - ComplianceRecord object | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">404</span> – vendor not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/compliance</code> | List compliance records by vendor | <code>canReviewCompliance</code> | <ul style="margin:0;padding-left:20px;"><li><code>?vendor_code</code> string ✅</li></ul> | <code>200</code> - Array of ComplianceRecord objects | <span style="background-color:#f8d7da;">404</span> – vendor not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/compliance/{code}</code> | Get compliance record by code | <code>canReviewCompliance</code> | None | <code>200</code> - ComplianceRecord object | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized |
| <code>PUT /api/v1/compliance/{code}</code> | Update compliance record | <code>canReviewCompliance</code> | <ul style="margin:0;padding-left:20px;"><li><code>certification_type</code> string ✅</li><li><code>status</code> string ✅ (Pending/Approved/Expired)</li><li><code>valid_from</code> string ❌ (YYYY-MM-DD)</li><li><code>valid_until</code> string ❌ (YYYY-MM-DD)</li><li><code>issued_by</code> string ❌</li><li><code>evidence_url</code> string ❌</li></ul> | <code>200</code> - ComplianceRecord object | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/compliance/expiring</code> | Get expiring certifications | <code>canReviewCompliance</code> | <ul style="margin:0;padding-left:20px;"><li><code>?days</code> int ❌ (default: 30)</li></ul> | <code>200</code> - Array of ComplianceRecord objects | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |

## 8. Contracts

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>POST /api/v1/contracts</code> | Create contract | <code>canEditVendor</code> | <ul style="margin:0;padding-left:20px;"><li><code>vendor_code</code> string ✅</li><li><code>contract_number</code> string ✅</li><li><code>start_date</code> string ✅ (YYYY-MM-DD)</li><li><code>end_date</code> string ✅ (YYYY-MM-DD)</li><li><code>contract_value</code> float64 ❌</li><li><code>renewal_status</code> string ✅ (Auto-Renew/Manual/Expiring)</li></ul> | <code>201</code> - Contract object | <span style="background-color:#f8d7da;">400</span> – validation failed<br><span style="background-color:#f8d7da;">404</span> – vendor not found<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/contracts</code> | List contracts by vendor | <code>canEditVendor</code> | <ul style="margin:0;padding-left:20px;"><li><code>?vendor_code</code> string ✅</li></ul> | <code>200</code> - Array of Contract objects | <span style="background-color:#f8d7da;">404</span> – vendor not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/contracts/{code}</code> | Get contract by code | <code>canEditVendor</code> | None | <code>200</code> - Contract object | <span style="background-color:#f8d7da;">404</span> – not found<br><span style="background-color:#f8d7da;">401</span> – unauthorized |
| <code>GET /api/v1/contracts/expiring</code> | Get expiring contracts | <code>canEditVendor</code> | <ul style="margin:0;padding-left:20px;"><li><code>?days</code> int ❌ (default: 30)</li></ul> | <code>200</code> - Array of Contract objects | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |

## 9. Audit Trail

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>GET /api/v1/audit</code> | List audit trail entries | <code>canViewAuditHistory</code> | <ul style="margin:0;padding-left:20px;"><li><code>?table</code> string ❌</li><li><code>?record_code</code> string ❌</li><li><code>?action</code> string ❌ (CREATE/UPDATE/DELETE)</li><li><code>?changed_by</code> string ❌</li><li><code>?date_from</code> string ❌ (YYYY-MM-DD)</li><li><code>?date_to</code> string ❌ (YYYY-MM-DD)</li><li><code>?limit</code> int ❌</li><li><code>?offset</code> int ❌</li></ul> | <code>200</code> - <code>{"data": AuditTrail[], "total": int}</code> | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |

## 10. Reports

| Method & Path | Purpose | Auth | Request | Response (Success) | Errors |
|---|---|---|---|---|---|
| <code>GET /api/v1/reports/summary</code> | Get vendor summary dashboard | <code>canAccessAllReports</code> or <code>canViewAssignedVendors</code> | None | <code>200</code> - SummaryResponse object with:<br><ul style="margin:0;padding-left:20px;"><li><code>total_vendors</code> int</li><li><code>vendors_by_status</code> map</li><li><code>vendors_by_risk_level</code> map</li><li><code>expiring_contracts_30_days</code> int</li><li><code>expiring_contracts_60_days</code> int</li><li><code>expiring_contracts_90_days</code> int</li><li><code>expiring_compliance_30_days</code> int</li><li><code>expiring_compliance_60_days</code> int</li><li><code>expiring_compliance_90_days</code> int</li><li><code>pending_risk_assessments</code> int</li><li><code>approved_risk_assessments</code> int</li></ul> | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
| <code>GET /api/v1/reports/monthly-onboarding</code> | Get monthly onboarding stats | <code>canAccessAllReports</code> or <code>canViewAssignedVendors</code> | None | <code>200</code> - Array of MonthlyOnboardingItem objects, each with:<br><ul style="margin:0;padding-left:20px;"><li><code>month</code> string (YYYY-MM)</li><li><code>count</code> int</li></ul> | <span style="background-color:#f8d7da;">401</span> – unauthorized<br><span style="background-color:#fff3cd;">500</span> – server error |
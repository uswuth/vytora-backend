# API Endpoints

Base URL: `http://localhost:8080`

The backend currently exposes **GraphQL** for app operations, plus **health/metrics** endpoints.

- GraphQL: `POST http://localhost:8080/graphql`
- Health: `GET http://localhost:8080/healthz`
- Ready: `GET http://localhost:8080/readyz`
- Metrics: `GET http://localhost:8080/metrics`

Auth: `Authorization: Bearer <token>`  
Content-Type: `application/json`

---

## GraphQL

All core operations are served through the single GraphQL endpoint.

| Method | Path | Auth |
|--------|------|------|
| POST | /graphql | Bearer |

### Request format
```json
{
  "query": "query ... { ... }",
  "variables": {}
}
```

### Playground
Open the browser playground at:
- `http://localhost:8080/`
- `http://localhost:8080/playground`

---

## Queries

**Me**
```graphql
query Me {
  me {
    id
    code
    email
    fullName
    role
    isActive
    createdAt
    updatedAt
  }
}
```

**Users**
```graphql
query Users {
  users {
    id
    code
    email
    fullName
    role
    isActive
    createdAt
    updatedAt
  }
}
```

**User**
```graphql
query User($id: ID!) {
  user(id: $id) {
    id
    code
    email
    fullName
    role
    isActive
  }
}
```

**vendors**
```graphql
query Vendors($filter: VendorListFilter, $limit: Int, $offset: Int, $sortBy: String, $sortOrder: String) {
  vendors(filter: $filter, limit: $limit, offset: $offset, sortBy: $sortBy, sortOrder: $sortOrder) {
    data {
      id
      code
      name
      status
      riskLevel
    }
    total
  }
}
```

**Vendor**
```graphql
query Vendor($code: String!) {
  vendor(code: $code) {
    id
    code
    name
    status
    category
    email
    phone
    address
    riskLevel
  }
}
```

**Compliance Records**
```graphql
query ComplianceRecords($vendorCode: String!) {
  complianceRecords(vendorCode: $vendorCode) {
    id
    code
    vendorCode
    certificationType
    validFrom
    validUntil
    issuedBy
    evidenceUrl
  }
}
```

**Contracts**
```graphql
query Contracts($vendorCode: String!) {
  contracts(vendorCode: $vendorCode) {
    id
    code
    vendorCode
    contractNumber
    startDate
    endDate
    contractValue
    renewalStatus
  }
}
```

**Risk Assessments**
```graphql
query RiskAssessments($vendorCode: String, $riskLevel: RiskLevel, $status: AssessmentStatus, $limit: Int, $offset: Int) {
  riskAssessments(vendorCode: $vendorCode, riskLevel: $riskLevel, status: $status, limit: $limit, offset: $offset) {
    id
    code
    vendorCode
    assessmentDate
    overallRiskScore
    riskLevel
    securityRiskScore
    financialRiskScore
    operationalRiskScore
    legalRiskScore
    status
    notes
  }
}
```

**Categories**
```graphql
query Categories($search: String, $status: String) {
  categories(search: $search, status: $status) {
    data {
      id
      code
      name
      displayName
      description
      status
    }
    total
  }
}
```

---

## Mutations

**Create User**
```graphql
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    code
    email
    fullName
    role
    isActive
  }
}
```

**Update User Role**
```graphql
mutation UpdateUserRole($id: ID!, $input: UpdateRoleInput!) {
  updateUserRole(id: $id, input: $input)
}
```

**Deactivate User**
```graphql
mutation DeactivateUser($id: ID!) {
  deactivateUser(id: $id)
}
```

**Activate User**
```graphql
mutation ActivateUser($id: ID!) {
  activateUser(id: $id)
}
```

**Login**
```graphql
mutation Login($email: String!, $password: String!) {
  login(email: $email, password: $password) {
    token
    expiresIn
    user {
      id
      code
      email
      fullName
      role
    }
  }
}
```

**Extend Session**
```graphql
mutation ExtendSession {
  extendSession {
    token
    expiresIn
  }
}
```

**Create Vendor**
```graphql
mutation CreateVendor($input: CreateVendorInput!) {
  createVendor(input: $input) {
    id
    code
    name
    status
    riskLevel
  }
}
```

**Update Vendor**
```graphql
mutation UpdateVendor($code: String!, $input: UpdateVendorInput!) {
  updateVendor(code: $code, input: $input) {
    id
    code
    name
    status
  }
}
```

**Delete Vendor**
```graphql
mutation DeleteVendor($code: String!) {
  deleteVendor(code: $code)
}
```

**Submit Vendor**
```graphql
mutation SubmitVendor($code: String!) {
  submitVendor(code: $code)
}
```

**Review Risk Vendor**
```graphql
mutation ReviewRiskVendor($code: String!) {
  reviewRiskVendor(code: $code)
}
```

**Review Compliance Vendor**
```graphql
mutation ReviewComplianceVendor($code: String!) {
  reviewComplianceVendor(code: $code)
}
```

**Approve Vendor**
```graphql
mutation ApproveVendor($code: String!) {
  approveVendor(code: $code)
}
```

**Reject Vendor**
```graphql
mutation RejectVendor($code: String!) {
  rejectVendor(code: $code)
}
```

**Create Contract**
```graphql
mutation CreateContract($input: CreateContractInput!) {
  createContract(input: $input) {
    id
    code
    vendorCode
    contractNumber
    startDate
    endDate
    contractValue
    renewalStatus
  }
}
```

**Create Compliance Record**
```graphql
mutation CreateComplianceRecord($vendorCode: String!, $input: CreateComplianceInput!) {
  createComplianceRecord(vendorCode: $vendorCode, input: $input) {
    id
    code
    vendorCode
    certificationType
    validFrom
    validUntil
  }
}
```

**Create Risk Assessment**
```graphql
mutation CreateRiskAssessment($input: CreateRiskAssessmentInput!) {
  createRiskAssessment(input: $input) {
    id
    code
    vendorCode
    assessmentDate
    overallRiskScore
    riskLevel
    status
  }
}
```

**Reject Risk Assessment**
```graphql
mutation RejectRiskAssessment($code: String!) {
  rejectRiskAssessment(code: $code)
}
```

**Create Contact**
```graphql
mutation CreateContact($vendorCode: String!, $input: CreateContactInput!) {
  createContact(vendorCode: $vendorCode, input: $input) {
    id
    code
    vendorId
    name
    email
    phone
    position
  }
}
```

**Update Contact**
```graphql
mutation UpdateContact($id: ID!, $input: UpdateContactInput!) {
  updateContact(id: $id, input: $input) {
    id
    code
    vendorId
    name
    email
    phone
    position
  }
}
```

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

Used by: categories, risk-assessments, vendors. All other list endpoints return raw arrays.

## Auth Header
`Authorization: Bearer <token>`  
Prefix `Bearer ` is case-insensitive.
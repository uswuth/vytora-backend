# VRMP — Implementation Guide

How to run the backend, where to start with the frontend, and phased delivery plan.

---

## Part 1: Backend — How to Use

### Prerequisites
- Go 1.24
- PostgreSQL 16
- curl or Postman for testing

### Step 1: Configuration

Create `.env` in project root:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/vrmp?sslmode=disable
JWT_SECRET=supersecretkey123!@#change_in_production
JWT_EXPIRY_HOURS=24
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
HEALTH_ALLOWED_IPS=127.0.0.1
```

### Step 2: Start PostgreSQL
```bash
# macOS (Homebrew)
brew services start postgresql

# Ubuntu/Debian
sudo systemctl start postgresql

# Windows
# Start PostgreSQL service from Services panel
```

### Step 3: Run Migrations
```bash
go run cmd/seed/main.go
```

This creates all tables and sequences, then inserts:
- Default admin user: `admin@vrmp.com` / `admin123`
- Default categories: Technology, Healthcare, Finance, Logistics

### Step 4: Start Server
```bash
go run cmd/server/main.go
```

Expected log output:
```
{"level":"info","service":"VRMP","request_id":"...","jwt_secret_prefix":"supersecre","msg":"JWT secret loaded"}
{"level":"info","service":"VRMP","msg":"VRMP server starting on http://localhost:8080"}
```

### Step 5: Verify
```bash
# Health check (from localhost only)
curl http://localhost:8080/healthz
# → {"status":"healthy"}

# Login
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@vrmp.com","password":"admin123"}'
# → copy the "token" from response
```

### Step 6: Authenticated Request
```bash
TOKEN="eyJhbGciOiJIUzI1NiJ9..."  # paste token from login

curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN"
# → returns current user info
```

### Common Operations

```bash
# Create category
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"technology","display_name":"Technology","description":"IT vendors","status":"Active"}'

# Create vendor
curl -X POST http://localhost:8080/api/v1/vendors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Corp","category":"technology","risk_level":"Medium","status":"Draft"}'

# Submit vendor for review
curl -X PUT http://localhost:8080/api/v1/vendors/{vendor_code}/submit \
  -H "Authorization: Bearer $TOKEN"

# List vendors
curl "http://localhost:8080/api/v1/vendors" \
  -H "Authorization: Bearer $TOKEN"

# Check metrics (localhost only)
curl http://localhost:8080/metrics
```

### Reset Database
```bash
go run cmd/reset/main.go
```
Drops all tables, recreates schema, re-seeds admin + categories.

---

## Part 2: Frontend — Where to Start

### Recommended Stack

| Layer | Choice | Why |
|-------|--------|-----|
| **Framework** | React 18+ (Vite) or Next.js | Component ecosystem, fast dev |
| **Language** | TypeScript | API types already defined in frontendDoc.md |
| **State** | Zustand or React Context | Simple, enough for JWT + session |
| **Router** | React Router v6 | Standard routing |
| **HTTP Client** | axios or fetch wrapper | Interceptors for auth/errors |
| **UI** | Tailwind CSS + shadcn/ui | Fast, clean, accessible |
| **Forms** | React Hook Form + Zod | Validation mirrors backend |
| **Tables** | TanStack Table | Sort/filter for list views |

### Project Setup
```bash
npm create vite@latest vrmp-frontend -- --template react-ts
cd vrmp-frontend
npm install
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
npm install zustand axios react-router-dom react-hook-form @hookform/resolvers zod
```

### Folder Structure
```
src/
  api/
    client.ts           — axios instance with interceptor
    auth.ts             — login, extend, logout
    vendors.ts          — vendor CRUD + workflow
    risk.ts             — risk assessments
    compliance.ts       — compliance records
    contracts.ts        — contracts
    categories.ts       — categories
    audit.ts            — audit trail
    reports.ts          — reports
    users.ts            — user management
    types.ts            — TypeScript interfaces (from frontendDoc.md)
  store/
    authStore.ts        — token, user, session timer
    vendorStore.ts      — vendor list, selected vendor
  components/
    Layout.tsx          — sidebar, header with timer
    ProtectedRoute.tsx  — auth guard
    SessionTimer.tsx    — countdown + extend button
    DataTable.tsx       — reusable table
    ConfirmDialog.tsx   — delete confirmation
  pages/
    Login.tsx
    Dashboard.tsx
    Vendors/
      List.tsx
      Create.tsx
      Detail.tsx
      Workflow.tsx
    Risk/
      List.tsx
      Create.tsx
      Detail.tsx
    Compliance/
      List.tsx
      Create.tsx
    Contracts/
      List.tsx
      Create.tsx
    Audit/
      List.tsx
    Reports/
      Summary.tsx
      Monthly.tsx
    Admin/
      Users.tsx
      Categories.tsx
  lib/
    utils.ts
```

### First Screen to Build: Login

```tsx
// src/api/client.ts
import axios from 'axios';

export const api = axios.create({ baseURL: 'http://localhost:8080/api/v1' });

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);
```

```tsx
// src/pages/Login.tsx
import { useState } from 'react';
import { api } from '../api/client';
import type { LoginRequest, LoginResponse } from '../api/types';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const { data } = await api.post<LoginResponse>('/login', {
      email,
      password,
    } as LoginRequest);
    localStorage.setItem('token', data.token);
    localStorage.setItem('expiresAt', Date.now() + data.expires_in * 1000);
    window.location.href = '/';
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" />
      <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Password" />
      <button type="submit">Login</button>
    </form>
  );
}
```

### Session Timer (Extend Flow)

```tsx
// src/components/SessionTimer.tsx
import { useState, useEffect } from 'react';

export default function SessionTimer() {
  const [remaining, setRemaining] = useState(0);

  useEffect(() => {
    const expiresAt = Number(localStorage.getItem('expiresAt'));
    const tick = () => {
      const r = Math.max(0, Math.floor((expiresAt - Date.now()) / 1000));
      setRemaining(r);
    };
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, []);

  const extend = async () => {
    await fetch('http://localhost:8080/api/v1/auth/extend', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    });
    // response gives new token + expires_in — update localStorage
  };

  if (remaining < 300) return <button onClick={extend}>Extend Session ({remaining}s)</button>;
  return <span>Session: {Math.floor(remaining / 60)}m {remaining % 60}s</span>;
}
```

---

## Part 3: Phase Implementation Plan

### Phase 1: Foundation (Week 1)
**Goal:** Login works, layout ready, auth flow complete.

| Task | Deliverable |
|------|-------------|
| Vite + React + TS + Tailwind bootstrap | Running dev server |
| Axios client with auth interceptor | Auto-attach token, redirect on 401 |
| Login page | Email + password form |
| Session timer component | Countdown + extend button |
| Protected route wrapper | Redirect to login if no token |
| Layout shell | Sidebar + header + content area |
| Dashboard page (stub) | Placeholder cards for stats |

**Milestone:** Admin logs in → sees dashboard → timer counts down → can extend session.

---

### Phase 2: Vendor Lifecycle (Week 2)
**Goal:** Full vendor CRUD + workflow.

| Task | Deliverable |
|------|-------------|
| Category management (admin) | Create/edit/delete categories |
| Vendor list page | Table with search, filters |
| Vendor create/edit form | All fields with validation |
| Vendor detail page | Read-only view |
| Workflow actions | Submit → Review Risk → Review Compliance → Approve / Reject |
| Status badge component | Color-coded by status |

**Milestone:** Admin creates category → creates vendor → runs through full approval workflow.

---

### Phase 3: Risk & Compliance (Week 3)
**Goal:** Risk assessment + compliance tracking.

| Task | Deliverable |
|------|-------------|
| Risk assessment create form | Score inputs (0–100) for 4 dimensions |
| Risk assessment list | Filterable by vendor, risk level, status |
| Risk approval action | Sets status to Approved |
| Compliance record create form | Certification type, valid_from, valid_until, evidence_url |
| Compliance list by vendor | Shows cert status (Pending/Approved/Expired) |
| Expiring query | Button to show certs expiring in X days |

**Milestone:** Risk manager completes assessment → compliance officer adds ISO cert → approves vendor.

---

### Phase 4: Contracts & Audit (Week 4)
**Goal:** Contracts, audit visibility.

| Task | Deliverable |
|------|-------------|
| Contract create/list | Linked to vendor, shows value + dates |
| Expiring contracts query | 30/60/90 day filter |
| Audit log page | Read-only table of all actions with details |
| Contract timeline view | Visual start/end bars |

**Milestone:** Contract created for approved vendor → audit entry visible → expiring contracts flagged.

---

### Phase 5: Reports & Polish (Week 5)
**Goal:** Dashboards + quality of life.

| Task | Deliverable |
|------|-------------|
| Summary dashboard | Stat cards (total vendors, high risk, pending approvals) |
| Monthly onboarding chart | Bar chart of onboarded vs approved per month |
| User management (admin) | Create/edit/deactivate users |
| Category management UI | Admin CRUD for categories |
| Error boundary + toast notifications | Friendly error messages |
| Loading skeletons | Per page/component |

**Milestone:** Admin sees real stats from queries → manager sees onboarding trends.

---

### Phase 6: Production Hardening (Week 6)
**Goal:** Ship-ready quality.

| Task | Deliverable |
|------|-------------|
| Replace `localhost:3000` CORS | Set production domain in .env |
| Environment-aware API base | `VITE_API_URL` for dev/prod |
| Dark mode toggle | Frontend polish |
| Keyboard shortcuts | Power-user features |
| E2E tests (Playwright) | Login → vendor create → approve → logout |
| Performance audit | Lighthouse / bundle size check |

---

## API Testing Quick Reference

### Login
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@vrmp.com","password":"admin123"}'
```

### Create Category
```bash
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"tech","display_name":"Technology","description":"IT vendors","status":"Active"}'
```

### Create Vendor
```bash
curl -X POST http://localhost:8080/api/v1/vendors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme","category":"tech","risk_level":"Medium","status":"Draft"}'
```

### List Vendors
```bash
curl "http://localhost:8080/api/v1/vendors" \
  -H "Authorization: Bearer $TOKEN"
```

### Workflow: Submit
```bash
curl -X PUT "http://localhost:8080/api/v1/vendors/{code}/submit" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Audit Log
```bash
curl "http://localhost:8080/api/v1/audit" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Troubleshooting

### "database connection failed"
- Check PostgreSQL is running: `pg_isready`
- Verify `DATABASE_URL` in `.env`
- Ensure database `vrmp` exists: `createdb vrmp`

### "missing authorization header"
- Every request except login/health needs `Authorization: Bearer <token>`
- Token may be expired — call `/api/v1/auth/extend` or re-login

### "rate limited"
- Default: 100 requests/minute per IP
- Wait 60 seconds or increase limit in `internal/middleware/rate_limiter.go`

### "forbidden"
- User role lacks the required permission for that endpoint
- Check `instrustion.md` role matrix

---

## Key Reminders

1. **Token lifetime** is set in `.env` as `JWT_EXPIRY_HOURS` (default 24h)
2. **CORS origins** must match frontend URL exactly in `ALLOWED_ORIGINS`
3. **Health endpoints** only respond to `HEALTH_ALLOWED_IPS` (default `127.0.0.1`)
4. **Commands — you should run them.**
5. **Backend compiles cleanly** — `go build ./...` should be silent
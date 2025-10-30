## Repository Restructuring Plan (Auth + Tenant/Owner Flows)

This plan organizes the codebase for clarity and growth. Follow steps in order. Check off boxes as you complete them.

### 1) Move postgres implementations (fix dir typo)
- From: `internal/repository/postrgres/`
- To:   `internal/repository/postgres/`
- Update imports anywhere referencing `postrgres` → `postgres`.

Suggested commands:
```bash
git mv internal/repository/postrgres internal/repository/postgres
rg -n "postrgres" | cut -d: -f1 | sort -u | xargs sed -i '' 's/postrgres/postgres/g'
```

### 2) Introduce HTTP layer with router + middleware
- Add `internal/http/router.go` that receives handlers and registers all routes.
- Add `internal/http/middleware/auth.go` with:
  - `LoadSession(next)` → attaches user to context if cookie present
  - `RequireOwner(next)` → ensures owner
  - `RequireTenant(next)` → ensures tenant

Routes to register in `router.go`:
- Public: `GET /login`, `POST /login`, `POST /logout`
- Owner-only: `GET /dashboard`, `POST /api/payments/mark-paid`, admin CRUD
- Tenant-only: `GET /me`, `POST /api/payments/submit`, `POST /api/me/change-password`

### 3) Split handlers by responsibility
- Keep existing logic but relocate:
  - `internal/http/handlers/auth_handler.go`
  - `internal/http/handlers/owner_handler.go` (dashboard, approve payment)
  - `internal/http/handlers/tenant_handler.go` (/me, submit txn, change password)
  - Keep `rental_handler.go` owner/admin-only; later split if needed.

Checklist:
- [ ] Create `internal/http/router.go`
- [ ] Create `internal/http/middleware/auth.go`
- [ ] Move/rename handlers into `internal/http/handlers`
- [ ] Update imports and registrations

### 4) Services: keep APIs; extract files for clarity
- Files in `internal/service/`:
  - `auth_service.go`
  - `payment_service.go`
  - `tenant_service.go`
  - `unit_service.go`

No behavioral changes expected—only file organization if needed.

### 5) Templates re-organization
- Move to folders:
  - `templates/auth/login.html`
  - `templates/owner/dashboard.html`
  - `templates/owner/unit-detail.html`
  - `templates/tenant/tenant-dashboard.html`

Then update `ExecuteTemplate` calls accordingly (e.g., `owner/dashboard.html`).

Checklist:
- [ ] Create subfolders under `templates/`
- [ ] Move files
- [ ] Update handler render paths

### 6) App bootstrap and main
- Add `internal/app/bootstrap.go` that wires:
  - config, DB, repositories, services, handlers, router
- Keep `cmd/server/main.go` minimal: load config → bootstrap → start server

Checklist:
- [ ] Create `internal/app/bootstrap.go`
- [ ] Refactor `cmd/server/main.go` to call bootstrap + start router

### 7) Config polish
- Add `COOKIE_NAME` (default `sid`) and optional `SESSION_TTL` to config/env.
- Ensure handlers use injected cookie name.

### 8) Build and smoke test
```bash
go build ./cmd/server
```
- Owner flow: login → dashboard → approve payment
- Tenant flow: login → /me → submit TXN → change password

### 9) Git steps (suggested)
```bash
git checkout -b refactor/http-layer
# perform steps 1–8 with small commits per step
git push -u origin refactor/http-layer
```

### Notes
- Keep package names: `domain`, `interfaces`, `postgres` consistent.
- Avoid changing public function signatures during the move.
- Run `rg`/build after each step to catch import path issues early.


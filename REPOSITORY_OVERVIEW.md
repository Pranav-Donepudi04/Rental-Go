# 📋 Repository Overview - Rental Property Management System

## 🎯 Project Purpose

This is a **Rental Property Management System** built in **Go** for managing a 3-floor property with 6 rental units. The system helps property owners manage tenants, track rent payments, and maintain property records.

---

## 🏗️ System Architecture

The application follows a **clean, layered architecture** pattern:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP REQUEST                              │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  HANDLERS LAYER                              │
│        (HTTP Request/Response, Template Rendering)            │
│  • rental_handler.go  • auth_handler.go  • tenant_handler.go│
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   SERVICE LAYER                              │
│              (Business Logic & Validation)                   │
│  • unit_service.go  • tenant_service.go  • payment_service.go│
│  • auth_service.go                                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                REPOSITORY LAYER                              │
│                  (Data Access)                               │
│  • interfaces/ (contracts)                                    │
│  • postgres/ (PostgreSQL implementations)                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  DATABASE (PostgreSQL)                       │
│  Tables: units, tenants, family_members, payments,          │
│          users, sessions                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Core Components

### 1. **Domain Models** (`internal/domain/`)

Business entities that represent the core concepts:

- **`unit.go`** - Rental units (e.g., G1, 1A, 2B)
  - Properties: UnitCode, Floor, UnitType, MonthlyRent, SecurityDeposit, PaymentDueDay, IsOccupied

- **`tenant.go`** - Primary rent payers
  - Properties: Name, Phone, AadharNumber, MoveInDate, NumberOfPeople, UnitID
  - Related: Unit, FamilyMembers

- **`payment.go`** - Rent payment records
  - Properties: TenantID, UnitID, Amount, PaymentDate, DueDate, IsPaid, PaymentMethod, UPIID, Notes
  - Status methods: GetStatus(), GetUserFacingStatus(), GetDaysOverdue()

- **`family_member.go`** - Family members of tenants
  - Properties: Name, Age, Relationship, AadharNumber (optional)

- **`user.go`** - Authentication users
  - Properties: Phone, PasswordHash, UserType (owner/tenant), TenantID (for tenant users)

- **`session.go`** - Active user sessions
  - Properties: Token, UserID, CreatedAt, ExpiresAt (7 days TTL)

### 2. **Services** (`internal/service/`)

Business logic layer:

- **`unit_service.go`** - Unit management
  - `GetAllUnits()`, `GetUnitByID()`, `GetAvailableUnits()`, `GetOccupiedUnits()`
  - `UpdateUnitOccupancy()`, `GetRentalSummary()`

- **`tenant_service.go`** - Tenant management
  - `CreateTenant()` - Creates tenant and updates unit occupancy
  - `GetAllTenants()`, `GetTenantByID()`, `UpdateTenant()`
  - `MoveOutTenant()` - Deletes tenant, payments, and updates unit
  - `AddFamilyMember()`, `GetFamilyMembersByTenantID()`

- **`payment_service.go`** - Payment management
  - `CreateMonthlyPayment()` - Creates payment for a specific month/year
  - `MarkPaymentAsPaid()` - Marks payment as paid with date and notes
  - `GetPaymentsByTenantID()`, `GetOverduePayments()`, `GetPendingPayments()`
  - `SubmitPaymentIntent()` - Allows tenants to submit UPI transaction IDs
  - `GetPaymentSummary()` - Returns payment statistics

- **`auth_service.go`** - Authentication & authorization
  - `Login()` - Validates credentials, creates session
  - `Logout()` - Invalidates session
  - `ValidateSession()` - Checks if session is valid and not expired
  - `CreateTenantCredentials()` - Creates login credentials for new tenants
  - Password hashing (SHA-256)

### 3. **Handlers** (`internal/handlers/`)

HTTP request handlers:

- **`rental_handler.go`** - Owner-facing endpoints
  - `Dashboard()` - Main dashboard view (GET `/dashboard`)
  - `UnitDetails()` - Unit detail page (GET `/unit/{id}`)
  - `GetUnits()` - API: List all units (GET `/api/units`)
  - `GetTenants()` - API: List all tenants (GET `/api/tenants`)
  - `CreateTenant()` - API: Create new tenant (POST `/api/tenants`)
  - `GetPayments()` - API: List all payments (GET `/api/payments`)
  - `MarkPaymentAsPaid()` - API: Mark payment as paid (POST `/api/payments/mark-paid`)
  - `VacateTenant()` - API: Move out tenant (POST `/api/tenants/vacate`)
  - `GetSummary()` - API: Dashboard summary (GET `/api/summary`)

- **`auth_handler.go`** - Authentication endpoints
  - `LoginPage()` - Login form (GET `/login`)
  - `Login()` - Process login (POST `/login`)
  - `Logout()` - Process logout (POST `/logout`)

- **`tenant_handler.go`** - Tenant-facing endpoints
  - `Me()` - Tenant dashboard (GET `/me`)
  - `SubmitPayment()` - Submit payment transaction ID (POST `/api/payments/submit`)
  - `ChangePassword()` - Change password (POST `/api/me/change-password`)

### 4. **Repositories** (`internal/repository/`)

Data access layer with interface-based design:

- **Interfaces** (`interfaces/`):
  - `unit_repository.go`
  - `tenant_repository.go`
  - `payment_repository.go`
  - `user_repository.go`
  - `session_repository.go`

- **PostgreSQL Implementations** (`postgres/`):
  - `postgres_unit_repository.go`
  - `postgres_tenant_repository.go`
  - `postgres_payment_repository.go`
  - `postgres_user_repository.go`
  - `postgres_session_repository.go`

### 5. **Routing & Middleware** (`internal/http/`)

- **`router.go`** - Sets up all HTTP routes with role-based middleware
  - `requireOwner()` - Ensures user is an owner
  - `requireTenant()` - Ensures user is a tenant
  - `loadSessionAndValidateRole()` - Validates session and role

- **`middleware/auth.go`** - Alternative middleware (currently not used in router)

---

## 🔐 Authentication & Authorization

### User Roles

1. **Owner** - Full access to all features
   - Dashboard, tenant management, payment tracking
   - Routes protected by `requireOwner()` middleware

2. **Tenant** - Limited access to own data
   - View own payment history
   - Submit payment transaction IDs
   - Change password
   - Routes protected by `requireTenant()` middleware

### Session Management

- Sessions stored in PostgreSQL `sessions` table
- Session token stored in cookie named `sid`
- Session TTL: **7 days**
- HTTP-only cookies for security

### Password Management

- Passwords hashed using **SHA-256**
- New tenants receive temporary password (random 8-byte hex)
- Tenants can change password via `/api/me/change-password`

---

## 🌐 API Endpoints

### Public Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Redirects to `/login` |
| GET | `/login` | Login page |
| POST | `/login` | Authenticate user |
| POST | `/logout` | Logout user |

### Owner Routes (Require Owner Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/dashboard` | Main dashboard (HTML) |
| GET | `/unit/{id}` | Unit detail page (HTML) |
| GET | `/api/units` | List all units (JSON) |
| GET | `/api/tenants` | List all tenants (JSON) |
| POST | `/api/tenants` | Create new tenant (JSON) |
| GET | `/api/payments` | List all payments (JSON) |
| POST | `/api/payments/mark-paid` | Mark payment as paid (JSON) |
| POST | `/api/tenants/vacate` | Move out tenant (JSON) |
| GET | `/api/summary` | Dashboard summary (JSON) |

### Tenant Routes (Require Tenant Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/me` | Tenant dashboard (HTML) |
| POST | `/api/payments/submit` | Submit payment transaction ID |
| POST | `/api/me/change-password` | Change password (JSON) |

---

## 📊 Property Structure

The system manages a 3-floor property with 6 units:

### Floor Layout

**Ground Floor (G):**
- G1: 1BHK - ₹7,000/month (Due: 10th)
- G2: Single Room - ₹3,000/month (Due: 5th)

**1st Floor:**
- 1A: 1BHK - ₹9,000/month (Due: 10th)
- 1B: Single Room - ₹2,500/month (Due: 5th)

**2nd Floor:**
- 2A: Single Room - ₹1,500/month (Due: 5th)
- 2B: Single Room - ₹1,500/month (Due: 5th)

### Payment Details

- **Payment Method**: UPI (9848790200@ybl)
- **Security Deposit**: 1 month rent for all units
- **Due Dates**: 
  - Single Rooms: 5th of every month
  - 1BHK Units: 10th of every month

---

## 🔑 Key Features

### Owner Features

1. **Dashboard Overview**
   - View all units (occupied/vacant)
   - Payment summaries (paid/pending/overdue)
   - Pending verifications (tenant-submitted transactions)
   - Rental income statistics

2. **Tenant Management**
   - Create new tenants with auto-generated login credentials
   - View tenant details with family members
   - Move out tenants (deletes payments and frees unit)

3. **Payment Tracking**
   - View all payment history
   - Mark payments as paid with date and notes
   - Track overdue payments
   - See pending verifications (transactions submitted by tenants)

4. **Unit Management**
   - View unit details
   - Track occupancy status
   - See payment history per unit

### Tenant Features

1. **Payment Submission**
   - Submit UPI transaction IDs for rent payment
   - Transaction IDs stored in payment notes with "TXN:" prefix
   - Owner sees as "Pending verification" status

2. **Payment History**
   - View own payment history
   - See payment status (Paid/Pending/Overdue/Pending verification)

3. **Account Management**
   - Change password
   - View tenant dashboard

---

## 📁 File Structure

```
pythonProject2/
├── cmd/
│   ├── server/
│   │   └── main.go              # Application entry point
│   └── test/
│       └── main.go              # Test utilities
│
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration loading
│   │   └── env.go               # Environment variables
│   │
│   ├── domain/                  # Business entities
│   │   ├── unit.go
│   │   ├── tenant.go
│   │   ├── payment.go
│   │   ├── family_member.go
│   │   ├── user.go
│   │   └── session.go
│   │
│   ├── handlers/                # HTTP handlers
│   │   ├── rental_handler.go
│   │   ├── auth_handler.go
│   │   └── tenant_handler.go
│   │
│   ├── http/
│   │   ├── router.go            # Route setup & middleware
│   │   └── middleware/
│   │       └── auth.go
│   │
│   ├── repository/
│   │   ├── interfaces/          # Repository contracts
│   │   │   ├── unit_repository.go
│   │   │   ├── tenant_repository.go
│   │   │   ├── payment_repository.go
│   │   │   ├── user_repository.go
│   │   │   └── session_repository.go
│   │   └── postgres/            # PostgreSQL implementations
│   │       ├── postgres_unit_repository.go
│   │       ├── postgres_tenant_repository.go
│   │       ├── postgres_payment_repository.go
│   │       ├── postgres_user_repository.go
│   │       └── postgres_session_repository.go
│   │
│   └── service/                 # Business logic
│       ├── unit_service.go
│       ├── tenant_service.go
│       ├── payment_service.go
│       └── auth_service.go
│
├── templates/                   # HTML templates
│   ├── dashboard.html           # Owner dashboard
│   ├── unit-detail.html         # Unit detail page
│   ├── login.html               # Login page
│   └── tenant-dashboard.html    # Tenant dashboard
│
├── docs/                       # Documentation
│   └── cursor_rules/
│       ├── RENTAL_MANAGEMENT_PRD.md
│       ├── ARCHITECTURE_FLOW.md
│       └── ...
│
├── go.mod                      # Go module dependencies
└── go.sum                      # Dependency checksums
```

---

## 🔄 Request Flow Examples

### Example 1: Owner Views Dashboard

```
Browser → GET /dashboard
  → Router.requireOwner() [validates session & role]
  → RentalHandler.Dashboard()
  → UnitService.GetAllUnits()
  → TenantService.GetAllTenants()
  → PaymentService.GetAllPayments()
  → UnitService.GetRentalSummary()
  → PaymentService.GetPaymentSummary()
  → Render dashboard.html template
  → HTML Response
```

### Example 2: Owner Creates Tenant

```
Browser → POST /api/tenants (JSON)
  → Router.requireOwner() [validates session & role]
  → RentalHandler.CreateTenant()
  → TenantService.CreateTenant()
    → Validate tenant data
    → Check unit availability
    → Create tenant record
    → Update unit occupancy
  → AuthService.CreateTenantCredentials()
    → Generate temp password
    → Create user record (or update existing)
  → Return JSON with tenant & temp password
```

### Example 3: Tenant Submits Payment

```
Browser → POST /api/payments/submit (form: txn_id)
  → Router.requireTenant() [validates session & role]
  → TenantHandler.SubmitPayment()
  → PaymentService.SubmitPaymentIntent()
    → Get current month payment (or create if doesn't exist)
    → Append "TXN:{txn_id}" to payment notes
    → Update payment record
  → 204 No Content response
```

### Example 4: Owner Marks Payment as Paid

```
Browser → POST /api/payments/mark-paid (JSON)
  → Router.requireOwner() [validates session & role]
  → RentalHandler.MarkPaymentAsPaid()
  → PaymentService.MarkPaymentAsPaid()
    → Get payment by ID
    → Validate not already paid
    → Set IsPaid = true
    → Set PaymentDate
    → Update Notes
  → Return JSON success
```

---

## 🗄️ Database Schema

The system uses PostgreSQL with the following tables:

1. **`units`** - Rental unit information
2. **`tenants`** - Primary tenant information
3. **`family_members`** - Family members of tenants
4. **`payments`** - Rent payment records
5. **`users`** - Authentication users (owner/tenant)
6. **`sessions`** - Active user sessions

*(Note: Schema should be provisioned externally - no auto-migrations)*

---

## ⚙️ Configuration

Configuration is loaded from environment variables:

- `SERVER_PORT` (default: "8080")
- `LOG_LEVEL` (default: "info")
- `DATABASE_URL` (full connection string) OR individual:
  - `DB_HOST` (default: "localhost")
  - `DB_PORT` (default: "5432")
  - `DB_USER` (default: "postgres")
  - `DB_PASSWORD`
  - `DB_NAME` (default: "formdb")
  - `DB_SSL_MODE` (default: "require")
- `DB_MAX_CONNECTIONS` (default: 25)
- `DB_CONNECTION_TIMEOUT` (default: 30 seconds)

---

## 🚀 Getting Started

1. **Set up PostgreSQL database** (schema must be provisioned externally)
2. **Configure environment variables** (`.env` file)
3. **Install dependencies**: `go mod download`
4. **Run server**: `go run cmd/server/main.go`

---

## 📝 Key Design Decisions

1. **Interface-based repositories** - Enables easy testing and swapping implementations
2. **Service layer** - Centralizes business logic and validation
3. **Role-based middleware** - Enforces authorization at route level
4. **Session-based auth** - 7-day sessions stored in database
5. **No auto-migrations** - Database schema managed externally
6. **Transaction ID tracking** - Tenants can submit UPI transaction IDs for verification
7. **Automatic credential creation** - New tenants get login credentials automatically

---

## 🔍 Notable Features

- **Payment Verification Workflow**: Tenants submit transaction IDs, owners verify and mark as paid
- **Automatic Unit Occupancy**: Creating tenant updates unit, moving out frees unit
- **Temporary Passwords**: New tenants receive random temporary passwords
- **Payment Status Tracking**: Tracks Paid, Pending, Overdue, and Pending Verification statuses
- **Dashboard Summaries**: Aggregated statistics for units and payments

---

*Last Updated: Based on current codebase structure*
*This document provides a comprehensive overview of the rental property management system.*


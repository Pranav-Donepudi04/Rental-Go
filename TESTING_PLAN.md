# Comprehensive Testing Plan

## 🎯 Testing Strategy

**Approach:** Test pyramid - Unit Tests (most) → Integration Tests → E2E Tests (fewest)

**Test Framework:** Go's built-in `testing` package + `testify` for assertions and mocks

**Coverage Goal:** 
- Services: 80%+ coverage
- Critical paths: 100% coverage
- Edge cases: All identified cases

---

## 📋 Test Structure

```
internal/
├── service/
│   ├── payment_service_test.go
│   ├── payment_query_service_test.go
│   ├── payment_transaction_service_test.go
│   ├── payment_history_service_test.go
│   ├── tenant_service_test.go
│   ├── unit_service_test.go
│   ├── auth_service_test.go
│   └── dashboard_service_test.go
├── handlers/
│   ├── rental_handler_test.go
│   ├── tenant_handler_test.go
│   └── auth_handler_test.go
├── domain/
│   ├── tenant_test.go
│   ├── payment_test.go
│   ├── family_member_test.go
│   └── payment_transaction_test.go
└── repository/
    └── postgres/
        ├── payment_repository_test.go
        └── tenant_repository_test.go (optional - integration tests)
```

---

## 🧪 Phase 1: Domain Validation Tests (Foundation)

### Priority: **HIGH** - Foundation for all other tests

#### `domain/tenant_test.go`
```go
TestTenant_Validate:
  ✅ Valid tenant passes
  ✅ Empty name fails
  ✅ Empty phone fails
  ✅ Empty Aadhar fails
  ✅ Aadhar != 12 digits fails
  ✅ Aadhar = 12 digits passes
  ✅ Zero move-in date fails
  ✅ NumberOfPeople <= 0 fails
  ✅ UnitID <= 0 fails
  ✅ All valid fields passes
```

#### `domain/payment_test.go`
```go
TestPayment_Validate:
  ✅ Valid payment passes
  ✅ TenantID <= 0 fails
  ✅ UnitID <= 0 fails
  ✅ Amount <= 0 fails
  ✅ DueDate zero fails
```

#### `domain/family_member_test.go`
```go
TestFamilyMember_Validate:
  ✅ Valid family member passes
  ✅ Empty name fails
  ✅ Empty relation fails
  ✅ TenantID <= 0 fails
  ✅ Age < 0 fails
  ✅ Age > 150 fails (edge case)
```

---

## 🧪 Phase 2: Service Layer Tests (Core Business Logic)

### 2.1 PaymentService Tests

#### File: `service/payment_service_test.go`

**Setup:** Mock `PaymentRepository`, `TenantRepository`, `UnitRepository`

```go
// Payment Creation
TestPaymentService_CreatePaymentForTenant:
  ✅ Creates payment with correct tenant/unit/amount
  ✅ Sets default values (AmountPaid=0, RemainingBalance=Amount)
  ✅ Returns error if tenant not found
  ✅ Returns error if unit not found
  ✅ Handles date timezone correctly

TestPaymentService_CreateNextPayment:
  ✅ Calculates next due date correctly (current + 1 month)
  ✅ Uses same tenant/unit/amount
  ✅ Returns error if current payment is nil

TestPaymentService_AutoCreateNextPayment:
  ✅ Creates next payment when current is fully paid
  ✅ Skips if current payment not fully paid
  ✅ Skips if next payment already exists
  ✅ Handles month/year rollover (Dec → Jan)

TestPaymentService_getOrCreateCurrentPayment:
  ✅ Returns existing unpaid payment if found
  ✅ Creates new payment if no unpaid payment exists
  ✅ Calculates due date correctly (next due day >= today)
  ✅ Handles month rollover
  ✅ Returns error if tenant not found
  ✅ Returns error if unit not found

// Payment Queries (moved to PaymentQueryService but test deprecated methods)
TestPaymentService_GetPaymentByID:
  ✅ Loads payment with tenant data
  ✅ Loads payment with unit data
  ✅ Loads transactions
  ✅ Returns error if payment not found

TestPaymentService_GetPaymentsByTenantID:
  ✅ Returns all payments for tenant
  ✅ Loads transactions for each payment
  ✅ Returns empty slice if no payments
  ✅ Handles tenant with no payments

// Payment Updates
TestPaymentService_MarkPaymentAsPaid:
  ✅ Marks payment as fully paid
  ✅ Updates AmountPaid = Amount
  ✅ Sets RemainingBalance = 0
  ✅ Sets PaymentDate and FullyPaidDate
  ✅ Returns error if payment already paid
  ✅ Returns error if payment not found
  ✅ Auto-creates next payment after marking paid
```

### 2.2 PaymentQueryService Tests

#### File: `service/payment_query_service_test.go`

**Setup:** Mock `PaymentRepository` only

```go
TestPaymentQueryService_GetAllPayments:
  ✅ Returns all payments from repository
  ✅ Handles empty result
  ✅ Returns error from repository

TestPaymentQueryService_GetOverduePayments:
  ✅ Filters payments where due date < now and not fully paid
  ✅ Excludes fully paid payments
  ✅ Excludes future payments
  ✅ Handles timezone correctly

TestPaymentQueryService_GetPendingPayments:
  ✅ Filters payments where due date >= now and not fully paid
  ✅ Excludes fully paid payments
  ✅ Excludes overdue payments

TestPaymentQueryService_GetPaymentSummary:
  ✅ Calculates total payments count
  ✅ Calculates paid payments count
  ✅ Calculates pending payments count
  ✅ Calculates overdue payments count
  ✅ Calculates total amount correctly
  ✅ Calculates paid amount correctly
  ✅ Uses remaining balance for pending/overdue amounts
  ✅ Handles empty payments list (all zeros)

TestPaymentQueryService_GetUnpaidPaymentsByTenantID:
  ✅ Returns unpaid payments for tenant
  ✅ Returns empty slice if all paid
  ✅ Returns error if tenant not found
```

### 2.3 PaymentTransactionService Tests

#### File: `service/payment_transaction_service_test.go`

**Setup:** Mock `PaymentRepository`, `PaymentService`

```go
TestPaymentTransactionService_SubmitPaymentIntent:
  ✅ Creates transaction with NULL amount
  ✅ Links transaction to payment
  ✅ Uses getOrCreateCurrentPayment from PaymentService
  ✅ Returns error if payment creation fails
  ✅ Returns nil if transaction already exists (idempotent)
  ✅ Returns error if transaction creation fails

TestPaymentTransactionService_VerifyTransaction:
  ✅ Updates transaction amount
  ✅ Updates payment amount_paid
  ✅ Updates payment remaining_balance
  ✅ Marks payment as fully paid if amount_paid >= amount
  ✅ Auto-creates next payment if fully paid
  ✅ Returns error if transaction not found
  ✅ Returns error if transaction already verified
  ✅ Returns error if amount <= 0
  ✅ Handles partial payment correctly

TestPaymentTransactionService_RejectTransaction:
  ✅ Deletes pending transaction
  ✅ Returns error if transaction not found
  ✅ Returns error if transaction already verified
  ✅ Does not affect payment amounts

TestPaymentTransactionService_GetPendingVerifications:
  ✅ Returns pending transactions for tenant
  ✅ Returns all pending if tenantID = 0
  ✅ Excludes verified transactions
  ✅ Returns empty slice if no pending
```

### 2.4 PaymentHistoryService Tests

#### File: `service/payment_history_service_test.go`

**Setup:** Mock `PaymentRepository`, `TenantRepository`, `UnitRepository`, `PaymentService`

```go
TestPaymentHistoryService_CreateHistoricalPaidPayment:
  ✅ Creates payment marked as fully paid
  ✅ Validates payment due date not before move-in date
  ✅ Calculates first valid payment date correctly
  ✅ Handles move-in day after payment due day
  ✅ Handles move-in day before payment due day
  ✅ Returns error if payment before move-in date
  ✅ Updates existing unpaid payment to paid
  ✅ Returns existing payment if already paid
  ✅ Returns error if tenant not found
  ✅ Returns error if unit not found

TestPaymentHistoryService_SyncPaymentHistory:
  ✅ Creates multiple historical payments
  ✅ Validates each payment against move-in date
  ✅ Returns error on first invalid payment
  ✅ Tracks latest paid date correctly
  ✅ Auto-creates next payment after sync
  ✅ Handles empty payments array
  ✅ Handles single payment

TestPaymentHistoryService_AdjustFirstPaymentDueDate:
  ✅ Updates first unpaid payment due date
  ✅ Returns error if no unpaid payments
  ✅ Returns error if tenant not found
  ✅ Updates correct payment (first unpaid)

TestPaymentHistoryService_AutoCreateNextPaymentAfterSync:
  ✅ Creates next payment after latest paid date
  ✅ Skips if payment already exists for next month
  ✅ Returns error if tenant not found
  ✅ Returns error if unit not found
```

### 2.5 TenantService Tests

#### File: `service/tenant_service_test.go`

**Setup:** Mock `TenantRepository`, `UnitRepository`, `PaymentService`

```go
TestTenantService_CreateTenant:
  ✅ Creates tenant successfully
  ✅ Validates tenant data
  ✅ Checks unit exists
  ✅ Checks unit not occupied
  ✅ Updates unit occupancy
  ✅ Creates first payment if skipFirstPayment = false
  ✅ Skips first payment if skipFirstPayment = true
  ✅ Rolls back tenant if unit update fails
  ✅ Returns error if validation fails
  ✅ Returns error if unit not found
  ✅ Returns error if unit occupied
  ✅ Returns error if tenant creation fails

TestTenantService_GetTenantByID:
  ✅ Returns tenant with unit data
  ✅ Returns tenant with family members
  ✅ Returns error if tenant not found
  ✅ Handles tenant with no unit
  ✅ Handles tenant with no family members

TestTenantService_MoveOutTenant:
  ✅ Deletes all payments for tenant
  ✅ Deletes tenant
  ✅ Updates unit occupancy to false
  ✅ Returns error if tenant not found
  ✅ Handles payment deletion failure gracefully

TestTenantService_AddFamilyMember:
  ✅ Creates family member
  ✅ Validates family member data
  ✅ Returns error if validation fails

TestTenantService_GetTenantsByUnitID:
  ✅ Returns tenants for unit
  ✅ Returns empty slice if no tenants
```

### 2.6 AuthService Tests

#### File: `service/auth_service_test.go`

**Setup:** Mock `UserRepository`, `SessionRepository`

```go
TestAuthService_Login:
  ✅ Returns session and user on valid credentials
  ✅ Returns error on invalid phone
  ✅ Returns error on invalid password
  ✅ Returns error if user not active
  ✅ Cleans up expired sessions
  ✅ Creates new session
  ✅ Handles password comparison correctly

TestAuthService_CreateTenantCredentials:
  ✅ Creates new user with tenant link
  ✅ Updates existing user password
  ✅ Links tenant if user.TenantID is nil
  ✅ Updates tenant link if different
  ✅ Keeps tenant link if same
  ✅ Returns temporary password
  ✅ Generates valid temp password format

TestAuthService_ValidateSession:
  ✅ Returns session if valid and not expired
  ✅ Returns error if session not found
  ✅ Returns error if session expired
  ✅ Deletes expired session

TestAuthService_Logout:
  ✅ Deletes session
  ✅ Handles missing cookie gracefully
```

### 2.7 DashboardService Tests

#### File: `service/dashboard_service_test.go`

**Setup:** Mock `UnitService`, `TenantService`, `PaymentQueryService`

```go
TestDashboardService_GetDashboardData:
  ✅ Aggregates units from UnitService
  ✅ Aggregates tenants from TenantService
  ✅ Aggregates payments from PaymentQueryService
  ✅ Gets unit summary
  ✅ Gets payment summary
  ✅ Returns error if any service fails
  ✅ Returns complete dashboard data

TestDashboardService_GetDashboardSummary:
  ✅ Returns unit summary only
  ✅ Returns payment summary only
  ✅ Returns error if unit summary fails
  ✅ Returns error if payment summary fails
```

---

## 🧪 Phase 3: Handler Tests (HTTP Layer)

### 3.1 RentalHandler Tests

#### File: `handlers/rental_handler_test.go`

**Setup:** Mock all services, use `httptest` package

```go
TestRentalHandler_Dashboard:
  ✅ Returns 200 with dashboard template
  ✅ Returns 405 for non-GET
  ✅ Handles service errors
  ✅ Passes correct data to template

TestRentalHandler_CreateTenant:
  ✅ Creates tenant successfully
  ✅ Returns temp password
  ✅ Validates JSON input
  ✅ Validates date format
  ✅ Returns 400 for invalid input
  ✅ Handles existing tenant flag
  ✅ Handles credential creation failure gracefully

TestRentalHandler_MarkPaymentAsPaid:
  ✅ Verifies transaction (new flow)
  ✅ Marks payment as paid (legacy flow)
  ✅ Requires owner authentication
  ✅ Validates transaction amount > 0
  ✅ Returns 400 for invalid input
  ✅ Returns 401 for unauthorized

TestRentalHandler_SyncPaymentHistory:
  ✅ Syncs payment history
  ✅ Validates tenant_id required
  ✅ Validates payments array required
  ✅ Returns 400 for invalid input
  ✅ Returns 401 for unauthorized

TestRentalHandler_RejectTransaction:
  ✅ Rejects pending transaction
  ✅ Returns 400 if transaction_id missing
  ✅ Returns 401 for unauthorized
  ✅ Handles rejection failure
```

### 3.2 TenantHandler Tests

#### File: `handlers/tenant_handler_test.go`

```go
TestTenantHandler_Me:
  ✅ Returns tenant dashboard
  ✅ Requires tenant authentication
  ✅ Returns 404 if tenant not found
  ✅ Returns 404 if user not linked
  ✅ Redirects to login if no session
  ✅ Loads payments for tenant

TestTenantHandler_SubmitPayment:
  ✅ Creates payment transaction
  ✅ Requires tenant authentication
  ✅ Validates txn_id required
  ✅ Returns 204 on success
  ✅ Returns 400 for invalid input
  ✅ Returns 401 for unauthorized

TestTenantHandler_ChangePassword:
  ✅ Updates password
  ✅ Validates old password
  ✅ Validates new password length >= 6
  ✅ Returns 400 for invalid input
  ✅ Returns 401 for unauthorized

TestTenantHandler_AddFamilyMember:
  ✅ Creates family member
  ✅ Validates family member data
  ✅ Returns 400 for invalid input
  ✅ Returns 401 for unauthorized
```

### 3.3 AuthHandler Tests

#### File: `handlers/auth_handler_test.go`

```go
TestAuthHandler_Login:
  ✅ Creates session on valid credentials
  ✅ Sets cookie
  ✅ Redirects owner to /dashboard
  ✅ Redirects tenant to /me
  ✅ Returns 401 for invalid credentials
  ✅ Validates role if provided
  ✅ Returns 400 for invalid JSON

TestAuthHandler_Logout:
  ✅ Deletes session
  ✅ Clears cookie
  ✅ Redirects to /login
```

---

## 🧪 Phase 4: Edge Cases & Integration Tests

### Critical Edge Cases

```go
// Date/Time Edge Cases
TestPaymentService_MonthBoundaries:
  ✅ Payment due on last day of month
  ✅ Next payment on first day of next month
  ✅ Year rollover (Dec 31 → Jan 1)
  ✅ Leap year handling (Feb 29)

TestPaymentService_Timezones:
  ✅ Handles different timezones correctly
  ✅ Move-in date timezone vs due date timezone

// Amount Edge Cases
TestPaymentService_PartialPayments:
  ✅ Multiple partial payments sum correctly
  ✅ Partial payment + full payment = fully paid
  ✅ Partial payment cannot exceed amount

// Transaction Edge Cases
TestPaymentTransactionService_DuplicateTransactions:
  ✅ Same transaction ID submitted twice (idempotent)
  ✅ Different amounts for same transaction ID
  ✅ Transaction ID case sensitivity

// Historical Payment Edge Cases
TestPaymentHistoryService_MoveInDateEdgeCases:
  ✅ Move-in on payment due day
  ✅ Move-in day after payment due day
  ✅ Move-in day before payment due day
  ✅ Move-in in different month

// Tenant Edge Cases
TestTenantService_ConcurrentOccupancy:
  ✅ Two tenants cannot occupy same unit
  ✅ Unit occupied check happens atomically

TestTenantService_PaymentCreationOnMoveIn:
  ✅ First payment created after move-in
  ✅ Payment due date calculated correctly
```

---

## 🧪 Phase 5: Repository Tests (Optional - Integration Tests)

### Only test critical database operations

#### File: `repository/postgres/payment_repository_test.go`

```go
// Requires real database connection (use test database)
TestPostgresPaymentRepository_CreatePayment:
  ✅ Creates payment with correct fields
  ✅ Handles NULL dates correctly
  ✅ Returns error on duplicate

TestPostgresPaymentRepository_VerifyTransaction:
  ✅ Updates transaction and payment atomically
  ✅ Handles concurrent updates
  ✅ Maintains data consistency
```

---

## 📊 Test Implementation Priority

### Phase 1: Foundation (Week 1)
1. ✅ Domain validation tests
2. ✅ Critical service tests (PaymentService core)
3. ✅ AuthService tests (security critical)

### Phase 2: Core Services (Week 2)
1. ✅ PaymentQueryService tests
2. ✅ PaymentTransactionService tests
3. ✅ TenantService tests

### Phase 3: Extended Services (Week 3)
1. ✅ PaymentHistoryService tests
2. ✅ DashboardService tests

### Phase 4: Handlers (Week 4)
1. ✅ Handler tests with mocked services
2. ✅ Authentication/authorization tests

### Phase 5: Edge Cases (Week 5)
1. ✅ Edge case tests
2. ✅ Integration tests (critical paths only)

---

## 🛠️ Test Utilities & Helpers

### Create: `internal/test/mocks/`

```go
// Mock repositories
type MockPaymentRepository struct { ... }
type MockTenantRepository struct { ... }
type MockUnitRepository struct { ... }
type MockUserRepository struct { ... }
type MockSessionRepository struct { ... }

// Test helpers
func NewTestPayment() *domain.Payment { ... }
func NewTestTenant() *domain.Tenant { ... }
func NewTestUnit() *domain.Unit { ... }

// Test fixtures
var ValidTenantData = ...
var ValidPaymentData = ...
```

---

## 📈 Coverage Goals

```
Target Coverage:
├── Services: 80%+
├── Critical Paths: 100%
│   ├── Payment creation
│   ├── Transaction verification
│   ├── Tenant creation
│   └── Authentication
├── Handlers: 70%+ (focus on happy paths + auth)
└── Domain: 100% (validation logic)
```

---

## 🚀 Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/service

# Run with verbose output
go test -v ./internal/service

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## ✅ Test Checklist Template

For each test function:
- [ ] Test name clearly describes what's being tested
- [ ] Tests one specific behavior
- [ ] Uses table-driven tests where appropriate
- [ ] Tests both success and failure cases
- [ ] Tests edge cases
- [ ] Uses meaningful assertions
- [ ] Cleans up test data (if integration test)
- [ ] Is independent (no shared state)

---

**Ready to start?** Let's begin with Phase 1 (Domain Validation Tests) - the foundation for everything else!


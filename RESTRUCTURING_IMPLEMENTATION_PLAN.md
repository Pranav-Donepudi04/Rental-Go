# Restructuring Implementation Plan

## 🔍 Complete Code Flow Analysis

### Current Initialization Flow (main.go)

```
1. Database Connection
   └─> db (sql.DB)

2. Repositories (Created in order)
   ├─> unitRepo (UnitRepository)
   ├─> tenantRepo (TenantRepository)
   ├─> paymentRepo (PaymentRepository)
   ├─> userRepo (UserRepository)
   └─> sessionRepo (SessionRepository)

3. Services (Created in specific order due to dependencies)
   ├─> unitService (UnitService)
   │   └─> Depends: unitRepo
   │
   ├─> paymentService (PaymentService) ⚠️ Created BEFORE tenantService
   │   └─> Depends: paymentRepo, tenantRepo, unitRepo
   │
   ├─> tenantService (TenantService)
   │   └─> Depends: tenantRepo, unitRepo, paymentService
   │
   └─> authService (AuthService)
       └─> Depends: userRepo, sessionRepo

4. Handlers (Created after services)
   ├─> rentalHandler (RentalHandler)
   │   └─> Depends: unitService, tenantService, paymentService, authService
   │
   ├─> authHandler (AuthHandler)
   │   └─> Depends: authService
   │
   └─> tenantHandler (TenantHandler)
       └─> Depends: tenantService, paymentService, userRepo, authService

5. Router
   ├─> Depends: authHandler, rentalHandler, tenantHandler, userRepo
   ├─> Calls: router.SetUserRepository(userRepo) on rentalHandler
   └─> Calls: router.SetupRoutes()
```

### Current Route Mapping (router.go)

**Owner Routes (requireOwner middleware):**
- `/dashboard` → `rentalHandler.Dashboard`
- `/unit/` → `rentalHandler.UnitDetails`
- `/api/units` → `rentalHandler.GetUnits`
- `/api/tenants` (GET/POST) → `rentalHandler.GetTenants` / `rentalHandler.CreateTenant`
- `/api/payments` → `rentalHandler.GetPayments`
- `/api/payments/mark-paid` → `rentalHandler.MarkPaymentAsPaid`
- `/api/payments/pending-verifications` → `rentalHandler.GetPendingVerifications`
- `/api/payments/reject-transaction` → `rentalHandler.RejectTransaction`
- `/api/payments/sync-history` → `rentalHandler.SyncPaymentHistory`
- `/api/payments/adjust-due-date` → `rentalHandler.AdjustPaymentDueDate`
- `/api/tenants/vacate` → `rentalHandler.VacateTenant`
- `/api/summary` → `rentalHandler.GetSummary`

**Tenant Routes (requireTenant middleware):**
- `/me` → `tenantHandler.Me`
- `/api/payments/submit` → `tenantHandler.SubmitPayment`
- `/api/me/change-password` → `tenantHandler.ChangePassword`
- `/api/me/family-members` → `tenantHandler.AddFamilyMember`

---

## 📋 Detailed Implementation Plan

### Phase 1: Extract PaymentQueryService (SAFEST - Start Here)

**Goal:** Separate read-only payment queries into independent service

**Why First:**
- Zero risk - only moving read operations
- No dependencies on tenantRepo/unitRepo needed
- Can coexist with existing PaymentService
- Immediate benefit: cleaner separation

**Steps:**

1. **Create `internal/service/payment_query_service.go`**
   ```go
   type PaymentQueryService struct {
       paymentRepo interfaces.PaymentRepository
   }
   
   // Methods to move (READ-ONLY, no tenantRepo/unitRepo needed):
   - GetAllPayments() ✅ (direct repo call)
   - GetOverduePayments() ✅ (filters payments, no deps)
   - GetPendingPayments() ✅ (filters payments, no deps)
   - GetPaymentSummary() ✅ (aggregates payments, no deps)
   - GetUnpaidPaymentsByTenantID() ✅ (direct repo call)
   
   // Methods to KEEP in PaymentService:
   - GetPaymentByID() ❌ (needs tenantRepo/unitRepo to load tenant/unit for display)
   - GetPaymentsByTenantID() ❌ (loads transactions - part of payment lifecycle)
   ```

2. **Update main.go** - Add PaymentQueryService creation
   ```go
   paymentQueryService := service.NewPaymentQueryService(paymentRepo)
   ```

3. **Update handlers** - Use PaymentQueryService for queries
   ```go
   // In RentalHandler:
   payments, err := h.paymentQueryService.GetAllPayments()
   summary, err := h.paymentQueryService.GetPaymentSummary()
   ```

4. **Keep PaymentService methods** (for backward compatibility during migration)
   ```go
   // Mark as deprecated, call queryService internally
   func (s *PaymentService) GetAllPayments() ([]*domain.Payment, error) {
       return s.queryService.GetAllPayments() // Forward to new service
   }
   ```

**Files to Create/Modify:**
- ✅ CREATE: `internal/service/payment_query_service.go`
- ✏️ MODIFY: `cmd/server/main.go` (add service creation)
- ✏️ MODIFY: `internal/handlers/rental_handler.go` (use queryService)
- ✏️ MODIFY: `internal/service/payment_service.go` (deprecate query methods)

**Testing:**
- All existing tests should pass
- Dashboard should work identically
- Summary endpoints should work

---

### Phase 2: Extract PaymentTransactionService

**Goal:** Separate transaction operations from payment CRUD

**Dependencies:**
- Needs `paymentService` for `getOrCreateCurrentPayment()`
- Needs `paymentRepo` for transaction operations

**Steps:**

1. **Create `internal/service/payment_transaction_service.go`**
   ```go
   type PaymentTransactionService struct {
       paymentRepo    interfaces.PaymentRepository
       paymentService *PaymentService // For getOrCreateCurrentPayment
   }
   
   // Methods to move:
   - SubmitPaymentIntent()
   - VerifyTransaction()
   - RejectTransaction()
   - GetPendingVerifications()
   ```

2. **Update main.go**
   ```go
   // Create PaymentTransactionService AFTER PaymentService
   paymentTransactionService := service.NewPaymentTransactionService(
       paymentRepo, 
       paymentService,
   )
   ```

3. **Update handlers**
   ```go
   // TenantHandler.SubmitPayment
   err := h.paymentTransactionService.SubmitPaymentIntent(...)
   
   // RentalHandler methods
   h.paymentTransactionService.VerifyTransaction(...)
   h.paymentTransactionService.RejectTransaction(...)
   ```

**Files to Create/Modify:**
- ✅ CREATE: `internal/service/payment_transaction_service.go`
- ✏️ MODIFY: `cmd/server/main.go`
- ✏️ MODIFY: `internal/handlers/tenant_handler.go`
- ✏️ MODIFY: `internal/handlers/rental_handler.go`
- ✏️ MODIFY: `internal/service/payment_service.go` (remove transaction methods)

**Testing:**
- Transaction submission should work
- Transaction verification should work
- Transaction rejection should work

---

### Phase 3: Extract PaymentHistoryService

**Goal:** Separate historical payment management

**Dependencies:**
- Needs `paymentRepo`, `tenantRepo`, `unitRepo`
- Needs `paymentService` for creating next payment

**Steps:**

1. **Create `internal/service/payment_history_service.go`**
   ```go
   type PaymentHistoryService struct {
       paymentRepo    interfaces.PaymentRepository
       tenantRepo     interfaces.TenantRepository
       unitRepo       interfaces.UnitRepository
       paymentService *PaymentService
   }
   
   // Methods to move:
   - CreateHistoricalPaidPayment()
   - SyncPaymentHistory()
   - AdjustFirstPaymentDueDate()
   - autoCreateNextPaymentAfterSync()
   ```

2. **Update main.go**
   ```go
   paymentHistoryService := service.NewPaymentHistoryService(
       paymentRepo,
       tenantRepo,
       unitRepo,
       paymentService,
   )
   ```

3. **Update handlers**
   ```go
   // RentalHandler
   h.paymentHistoryService.SyncPaymentHistory(...)
   h.paymentHistoryService.AdjustFirstPaymentDueDate(...)
   ```

**Files to Create/Modify:**
- ✅ CREATE: `internal/service/payment_history_service.go`
- ✏️ MODIFY: `cmd/server/main.go`
- ✏️ MODIFY: `internal/handlers/rental_handler.go`
- ✏️ MODIFY: `internal/service/payment_service.go` (remove history methods)

---

### Phase 4: Extract DashboardService

**Goal:** Separate dashboard aggregation logic

**Dependencies:**
- Needs `unitService`, `tenantService`, `paymentQueryService`

**Steps:**

1. **Create `internal/service/dashboard_service.go`**
   ```go
   type DashboardService struct {
       unitService        *UnitService
       tenantService      *TenantService
       paymentQueryService *PaymentQueryService
   }
   
   // Methods:
   - GetDashboardData() // Returns all data for dashboard
   - GetDashboardSummary() // Returns summary only
   ```

2. **Update main.go**
   ```go
   dashboardService := service.NewDashboardService(
       unitService,
       tenantService,
       paymentQueryService,
   )
   ```

3. **Update handlers**
   ```go
   // RentalHandler.Dashboard
   data, err := h.dashboardService.GetDashboardData()
   ```

**Files to Create/Modify:**
- ✅ CREATE: `internal/service/dashboard_service.go`
- ✏️ MODIFY: `cmd/server/main.go`
- ✏️ MODIFY: `internal/handlers/rental_handler.go`

---

### Phase 5: Split Handlers (After Services are Split)

**Strategy:** Split handlers after services are stable

#### 5A. Create TransactionHandler

**Purpose:** Handle transaction verification/rejection

**Methods:**
```go
- GetPendingVerifications()
- VerifyTransaction()
- RejectTransaction()
```

**Dependencies:**
- `paymentTransactionService`
- `authService` (for session)

**Route Updates:**
```go
// In router.go
transactionHandler := handlers.NewTransactionHandler(
    paymentTransactionService,
    authService,
)

// Routes
"/api/payments/pending-verifications" → transactionHandler.GetPendingVerifications
"/api/payments/mark-paid" (transaction part) → transactionHandler.VerifyTransaction
"/api/payments/reject-transaction" → transactionHandler.RejectTransaction
```

---

#### 5B. Create PaymentHandler

**Purpose:** Handle payment CRUD operations

**Methods:**
```go
- GetPayments()
- MarkPaymentAsPaid() // Legacy full payment marking
```

**Dependencies:**
- `paymentService`
- `authService`

---

#### 5C. Create OwnerTenantHandler

**Purpose:** Handle tenant management from owner perspective

**Methods:**
```go
- GetTenants()
- CreateTenant()
- VacateTenant()
```

**Dependencies:**
- `tenantService`
- `authService`
- `paymentHistoryService` (for sync)

---

#### 5D. Create DashboardHandler

**Purpose:** Handle dashboard view

**Methods:**
```go
- Dashboard() // Render dashboard page
- GetSummary() // JSON summary
```

**Dependencies:**
- `dashboardService`
- `templates`

---

#### 5E. Create UnitHandler

**Purpose:** Handle unit operations

**Methods:**
```go
- GetUnits()
- UnitDetails()
```

**Dependencies:**
- `unitService`
- `tenantService` (for tenant in unit)
- `paymentService` (for payments in unit)
- `paymentTransactionService` (for pending verifications)
- `templates`

---

## 📊 Final Dependency Graph (After All Phases)

### Repositories
```
unitRepo, tenantRepo, paymentRepo, userRepo, sessionRepo
```

### Services (Creation Order)
```
1. unitService (unitRepo)
2. paymentService (paymentRepo, tenantRepo, unitRepo)
3. paymentQueryService (paymentRepo) ← NEW, no deps
4. paymentTransactionService (paymentRepo, paymentService)
5. paymentHistoryService (paymentRepo, tenantRepo, unitRepo, paymentService)
6. tenantService (tenantRepo, unitRepo, paymentService)
7. dashboardService (unitService, tenantService, paymentQueryService)
8. authService (userRepo, sessionRepo)
```

### Handlers (Creation Order)
```
1. unitHandler (unitService, tenantService, paymentService, paymentTransactionService)
2. ownerTenantHandler (tenantService, authService, paymentHistoryService)
3. paymentHandler (paymentService, authService)
4. transactionHandler (paymentTransactionService, authService)
5. dashboardHandler (dashboardService, templates)
6. tenantHandler (tenantService, paymentTransactionService, userRepo, authService)
7. authHandler (authService)
```

### Router
```
router (all handlers, userRepo)
```

---

## 🚨 Critical Dependencies to Preserve

### 1. Service Creation Order
```go
// CORRECT ORDER:
paymentService := service.NewPaymentService(...)      // Must be first
paymentTransactionService := service.NewPaymentTransactionService(..., paymentService)
paymentHistoryService := service.NewPaymentHistoryService(..., paymentService)
tenantService := service.NewTenantService(..., paymentService)  // After paymentService
```

### 2. PaymentService Core Methods (Keep in PaymentService)
```go
// These MUST stay in PaymentService because:
- CreatePaymentForTenant() // Used by TenantService and PaymentTransactionService
- getOrCreateCurrentPayment() // Used by PaymentTransactionService
- CreateNextPayment() // Used by PaymentHistoryService
- autoCreateNextPayment() // Used internally
```

### 3. Methods That Load Related Data
```go
// Keep in PaymentService (needs tenant/unit repos):
- GetPaymentByID() // Loads tenant/unit for display
- GetPaymentsByTenantID() // Loads transactions
```

---

## ✅ Step-by-Step Migration Checklist

### Phase 1: PaymentQueryService ✅ (Start Here)

- [ ] Create `internal/service/payment_query_service.go`
- [ ] Move query methods: `GetAllPayments`, `GetOverduePayments`, `GetPendingPayments`, `GetPaymentSummary`
- [ ] Update `main.go` to create `PaymentQueryService`
- [ ] Update `RentalHandler` to use `PaymentQueryService`
- [ ] Update `Dashboard` method to use query service
- [ ] Update `GetSummary` method to use query service
- [ ] Keep deprecated methods in `PaymentService` (forward calls)
- [ ] Test: Dashboard works
- [ ] Test: Summary endpoint works
- [ ] Test: All payment queries work

### Phase 2: PaymentTransactionService

- [ ] Create `internal/service/payment_transaction_service.go`
- [ ] Move: `SubmitPaymentIntent`, `VerifyTransaction`, `RejectTransaction`, `GetPendingVerifications`
- [ ] Keep `getOrCreateCurrentPayment` in PaymentService (called by transaction service)
- [ ] Update `main.go` to create `PaymentTransactionService`
- [ ] Update `TenantHandler.SubmitPayment` to use transaction service
- [ ] Update `RentalHandler` transaction methods to use transaction service
- [ ] Test: Tenant can submit transaction
- [ ] Test: Owner can verify transaction
- [ ] Test: Owner can reject transaction

### Phase 3: PaymentHistoryService

- [ ] Create `internal/service/payment_history_service.go`
- [ ] Move: `CreateHistoricalPaidPayment`, `SyncPaymentHistory`, `AdjustFirstPaymentDueDate`, `autoCreateNextPaymentAfterSync`
- [ ] Update `main.go` to create `PaymentHistoryService`
- [ ] Update `RentalHandler` history methods
- [ ] Test: Sync payment history works
- [ ] Test: Adjust due date works

### Phase 4: DashboardService

- [ ] Create `internal/service/dashboard_service.go`
- [ ] Move aggregation logic from `RentalHandler.Dashboard`
- [ ] Update `main.go` to create `DashboardService`
- [ ] Update `RentalHandler.Dashboard` to use service
- [ ] Test: Dashboard renders correctly

### Phase 5: Split Handlers

- [ ] Create `TransactionHandler`
- [ ] Move transaction methods from `RentalHandler`
- [ ] Update router to use `TransactionHandler`
- [ ] Test: All transaction routes work
- [ ] Create `PaymentHandler`
- [ ] Create `OwnerTenantHandler`
- [ ] Create `DashboardHandler`
- [ ] Create `UnitHandler`
- [ ] Update router for all new handlers
- [ ] Deprecate `RentalHandler`

---

## 🔧 Implementation Details

### Shared Methods Strategy

**Problem:** Some methods are used by multiple services
- `getOrCreateCurrentPayment()` - used by PaymentTransactionService
- `CreatePaymentForTenant()` - used by TenantService and PaymentTransactionService

**Solution:** Keep shared methods in PaymentService
```go
// PaymentService (core methods - shared)
func (s *PaymentService) getOrCreateCurrentPayment(tenantID int) (*domain.Payment, error)
func (s *PaymentService) CreatePaymentForTenant(...) (*domain.Payment, error)

// PaymentTransactionService (uses PaymentService)
func (s *PaymentTransactionService) SubmitPaymentIntent(tenantID int, txnID string) error {
    payment, err := s.paymentService.getOrCreateCurrentPayment(tenantID)
    // ... transaction creation
}
```

---

## 📁 Final File Structure

```
internal/
├── handlers/
│   ├── dashboard_handler.go          (NEW - Phase 5)
│   ├── unit_handler.go                 (NEW - Phase 5)
│   ├── owner_tenant_handler.go         (NEW - Phase 5)
│   ├── payment_handler.go              (NEW - Phase 5)
│   ├── transaction_handler.go         (NEW - Phase 5)
│   ├── rental_handler.go              (DEPRECATED - remove after Phase 5)
│   ├── tenant_handler.go              (KEEP - update in Phase 2)
│   └── auth_handler.go                (KEEP)
│
├── service/
│   ├── payment_service.go              (REFACTORED - core CRUD only)
│   ├── payment_query_service.go        (NEW - Phase 1)
│   ├── payment_transaction_service.go  (NEW - Phase 2)
│   ├── payment_history_service.go      (NEW - Phase 3)
│   ├── dashboard_service.go           (NEW - Phase 4)
│   ├── tenant_service.go              (KEEP - no changes needed)
│   ├── unit_service.go                 (KEEP - no changes needed)
│   └── auth_service.go                 (KEEP - no changes needed)
```

---

## 🎯 Ready to Start?

**Phase 1 (PaymentQueryService) is ready to implement:**
- ✅ No breaking changes
- ✅ Can coexist with existing code
- ✅ Immediate benefit
- ✅ Low risk

**Should I proceed with Phase 1?**


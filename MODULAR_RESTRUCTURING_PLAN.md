# Modular Restructuring Plan

## 📊 Current State Analysis

### RentalHandler (593 lines, 14 methods)
**Responsibilities:**
- ✅ Dashboard aggregation
- ✅ Unit CRUD operations
- ✅ Tenant CRUD operations  
- ✅ Payment CRUD operations
- ✅ Transaction verification/rejection
- ✅ Payment history sync
- ✅ Payment due date adjustment
- ✅ Summary/aggregation endpoints

**Dependencies:**
- `unitService`, `tenantService`, `paymentService`, `authService`, `userRepo`, `templates`

**Issues:**
- Too many responsibilities (violates Single Responsibility Principle)
- Mixing CRUD operations with complex business logic
- Hard to test individual features
- Difficult to locate specific functionality

---

### PaymentService (665 lines, 24 methods)
**Responsibilities grouped by domain:**
1. **Payment CRUD** (5 methods): Create, Read, Update, Delete operations
2. **Payment Queries** (4 methods): Overdue, Pending, Summary, Status queries
3. **Transaction Management** (4 methods): Submit, Verify, Reject, GetPending
4. **Payment Lifecycle** (4 methods): Create, Auto-create, Due date calculation
5. **Historical Payments** (3 methods): Sync history, Create historical, Adjust dates
6. **Helpers** (4 methods): Internal utility methods

**Dependencies:**
- `paymentRepo`, `tenantRepo`, `unitRepo` ✅ (All valid - needs rent/tenant info)

**Issues:**
- Multiple responsibilities (CRUD + Queries + Transactions + History)
- Large file harder to navigate
- Can be split for better organization

---

## 🎯 Proposed Modular Structure

### Strategy: Split by Domain & Responsibility

```
Handlers/
├── dashboard_handler.go      (Dashboard aggregation)
├── unit_handler.go           (Unit CRUD)
├── owner_tenant_handler.go   (Owner's tenant management)
├── payment_handler.go        (Payment operations for owners)
├── transaction_handler.go    (Transaction verification/rejection)
└── rental_handler.go         (DEPRECATED - merge into above)

Services/
├── payment_service.go           (Core payment CRUD & lifecycle)
├── payment_query_service.go     (Queries: overdue, pending, summaries)
├── payment_transaction_service.go (Transaction operations)
├── payment_history_service.go   (Historical payment management)
└── dashboard_service.go        (Dashboard aggregation logic)
```

---

## 📋 Detailed Restructuring Plan

### 1. Handler Layer Split

#### **A. DashboardHandler** (~100 lines)
**Purpose:** Aggregates data for dashboard view

**Methods:**
```go
- Dashboard() // Renders dashboard page
- GetSummary() // Returns JSON summary
```

**Dependencies:**
- `unitService`, `tenantService`, `paymentService` (or `paymentQueryService`)
- `dashboardService` (new - handles aggregation logic)
- `templates`

**Benefits:**
- Separates dashboard logic from CRUD operations
- Easier to optimize dashboard queries
- Can cache dashboard data separately

---

#### **B. UnitHandler** (~150 lines)
**Purpose:** Unit management operations

**Methods:**
```go
- GetUnits() // Get all units (JSON)
- GetUnitDetails() // Get unit detail page
```

**Dependencies:**
- `unitService`
- `tenantService` (for getting tenant in unit)
- `paymentService` (for getting payments in unit)
- `templates`

**Benefits:**
- Focused on unit operations
- Clear separation of concerns

---

#### **C. OwnerTenantHandler** (~200 lines)
**Purpose:** Tenant management from owner's perspective

**Methods:**
```go
- GetTenants() // Get all tenants (JSON)
- CreateTenant() // Create new tenant
- VacateTenant() // Move out tenant
- SyncPaymentHistory() // Sync historical payments
- AdjustPaymentDueDate() // Adjust payment due date
```

**Dependencies:**
- `tenantService`
- `authService` (for creating credentials)
- `paymentHistoryService` (for syncing)

**Benefits:**
- Separates owner tenant operations from general tenant operations
- Clear distinction between owner and tenant views

---

#### **D. PaymentHandler** (~150 lines)
**Purpose:** Payment management operations

**Methods:**
```go
- GetPayments() // Get all payments (JSON)
- MarkPaymentAsPaid() // Legacy payment marking
```

**Dependencies:**
- `paymentService`
- `authService` (for user session)

**Benefits:**
- Focused on payment CRUD
- Separates from transaction operations

---

#### **E. TransactionHandler** (~150 lines)
**Purpose:** Transaction verification and rejection

**Methods:**
```go
- GetPendingVerifications() // Get pending transactions
- VerifyTransaction() // Verify a transaction
- RejectTransaction() // Reject a transaction
```

**Dependencies:**
- `paymentTransactionService`
- `authService` (for user session)

**Benefits:**
- Clear separation of transaction operations
- Easier to add transaction-specific features (notifications, audit logs)

---

### 2. Service Layer Split

#### **A. PaymentService** (Core - ~250 lines)
**Purpose:** Core payment CRUD and lifecycle management

**Methods:**
```go
// CRUD Operations
- CreatePaymentForTenant()
- GetPaymentByID()
- GetPaymentsByTenantID()
- GetUnpaidPaymentsByTenantID()
- UpdatePayment() (via repo)
- DeletePayment()
- DeletePaymentsByTenantID()

// Lifecycle
- CreateNextPayment()
- autoCreateNextPayment()
- getOrCreateCurrentPayment()

// Shared by other services
- CreateMonthlyPayment() // Legacy support
```

**Dependencies:**
- `paymentRepo`, `tenantRepo`, `unitRepo`

**Why keep tenantRepo/unitRepo?**
- Needs to fetch rent amount when creating payments
- Needs tenant/unit info for payment creation
- This dependency is VALID and necessary

---

#### **B. PaymentQueryService** (~150 lines)
**Purpose:** Read-only payment queries and aggregations

**Methods:**
```go
// Queries
- GetAllPayments()
- GetOverduePayments()
- GetPendingPayments()
- GetPaymentSummary()
```

**Dependencies:**
- `paymentRepo` only ✅ (No need for tenant/unit repos)

**Benefits:**
- Pure read operations
- Can be cached independently
- No business logic, just queries
- Much simpler dependencies

---

#### **C. PaymentTransactionService** (~200 lines)
**Purpose:** Transaction submission, verification, and rejection

**Methods:**
```go
- SubmitPaymentIntent() // Tenant submits transaction
- VerifyTransaction() // Owner verifies
- RejectTransaction() // Owner rejects
- GetPendingVerifications() // Get pending transactions
```

**Dependencies:**
- `paymentRepo`, `paymentService` (for creating payment if needed)

**Why paymentService?**
- Needs to get or create current payment when tenant submits
- Uses `paymentService.getOrCreateCurrentPayment()`

**Benefits:**
- Focused on transaction operations
- Clear transaction workflow
- Can add transaction-specific features (email notifications, audit logs)

---

#### **D. PaymentHistoryService** (~150 lines)
**Purpose:** Historical payment management (migration/backfill)

**Methods:**
```go
- CreateHistoricalPaidPayment()
- SyncPaymentHistory()
- AdjustFirstPaymentDueDate()
- autoCreateNextPaymentAfterSync()
```

**Dependencies:**
- `paymentRepo`, `tenantRepo`, `unitRepo`
- `paymentService` (for creating next payment)

**Benefits:**
- Isolated historical payment logic
- Doesn't pollute core payment service
- Easy to test migration scenarios

---

#### **E. DashboardService** (~100 lines)
**Purpose:** Aggregates data for dashboard

**Methods:**
```go
- GetDashboardData() // Returns all dashboard data
- GetDashboardSummary() // Returns summary only
```

**Dependencies:**
- `unitService`, `tenantService`, `paymentQueryService`

**Benefits:**
- Encapsulates dashboard aggregation logic
- Can optimize queries (parallel fetching, caching)
- Handlers just call one service method
- Easy to add dashboard-specific caching

---

## 🔄 Dependency Graph (After Restructuring)

```
Handlers (Thin Layer - HTTP handling only)
│
├─ DashboardHandler
│  └─ DashboardService
│     ├─ UnitService
│     ├─ TenantService
│     └─ PaymentQueryService ✅ (no tenant/unit repo!)
│
├─ UnitHandler
│  └─ UnitService
│
├─ OwnerTenantHandler
│  ├─ TenantService
│  ├─ AuthService
│  └─ PaymentHistoryService
│
├─ PaymentHandler
│  └─ PaymentService
│
└─ TransactionHandler
   └─ PaymentTransactionService
      └─ PaymentService (for getOrCreateCurrentPayment)

Services (Business Logic)
│
├─ PaymentService (Core)
│  ├─ paymentRepo
│  ├─ tenantRepo ✅ (needs rent info)
│  └─ unitRepo ✅ (needs rent info)
│
├─ PaymentQueryService (Read-only)
│  └─ paymentRepo ✅ (no other deps!)
│
├─ PaymentTransactionService
│  ├─ paymentRepo
│  └─ PaymentService (for payment creation)
│
├─ PaymentHistoryService
│  ├─ paymentRepo
│  ├─ tenantRepo
│  ├─ unitRepo
│  └─ PaymentService (for next payment)
│
└─ DashboardService (Aggregation)
   ├─ UnitService
   ├─ TenantService
   └─ PaymentQueryService
```

---

## 📝 Migration Steps

### Phase 1: Extract Query Service (Lowest Risk)
1. Create `PaymentQueryService` with read-only methods
2. Update handlers to use `PaymentQueryService` for queries
3. Keep `PaymentService` methods for backward compatibility
4. Gradually migrate all query calls

**Impact:** Minimal - just moving read operations

---

### Phase 2: Extract Transaction Service (Medium Risk)
1. Create `PaymentTransactionService`
2. Move transaction methods from `PaymentService`
3. Update `TenantHandler` to use `PaymentTransactionService`
4. Update `RentalHandler` transaction methods
5. Remove transaction methods from `PaymentService`

**Impact:** Medium - transaction flow changes

---

### Phase 3: Extract History Service (Low Risk)
1. Create `PaymentHistoryService`
2. Move historical payment methods
3. Update `RentalHandler.SyncPaymentHistory`
4. Update `RentalHandler.AdjustPaymentDueDate`

**Impact:** Low - only affects migration features

---

### Phase 4: Split Handlers (Medium Risk)
1. Create `DashboardHandler` - move dashboard methods
2. Create `UnitHandler` - move unit methods
3. Create `OwnerTenantHandler` - move tenant creation/vacate
4. Create `PaymentHandler` - move payment CRUD
5. Create `TransactionHandler` - move transaction methods
6. Update router to use new handlers
7. Deprecate `RentalHandler` (remove after migration)

**Impact:** Medium - router changes needed

---

### Phase 5: Create Dashboard Service (Low Risk)
1. Create `DashboardService` for aggregation logic
2. Move aggregation from `DashboardHandler` to service
3. Handler becomes thin wrapper

**Impact:** Low - internal refactoring

---

## ✅ Benefits After Restructuring

### 1. **Single Responsibility**
- Each handler/service has one clear purpose
- Easier to understand and maintain

### 2. **Reduced Coupling**
- `PaymentQueryService` has minimal dependencies (only paymentRepo)
- Services can be tested independently
- Clear dependency boundaries

### 3. **Better Testability**
- Mock individual services instead of large ones
- Test transaction logic separately from payment CRUD
- Test queries without transaction complexity

### 4. **Easier Navigation**
- Know where to find functionality:
  - Payment queries? → `PaymentQueryService`
  - Transaction verification? → `TransactionHandler` + `PaymentTransactionService`
  - Dashboard? → `DashboardHandler` + `DashboardService`

### 5. **Scalability**
- Can add caching to `PaymentQueryService` easily
- Can add transaction-specific features without touching payment CRUD
- Can optimize dashboard queries independently

### 6. **Team Collaboration**
- Multiple developers can work on different handlers/services
- Less merge conflicts
- Clearer code ownership

---

## ⚠️ Considerations

### Backward Compatibility
- Keep old methods during migration (mark as deprecated)
- Gradual migration reduces risk
- Can roll back easily if issues

### Circular Dependencies to Avoid
- ✅ `PaymentTransactionService` → `PaymentService` (OK - one-way)
- ✅ `PaymentHistoryService` → `PaymentService` (OK - one-way)
- ❌ Don't create: `PaymentService` → `PaymentTransactionService` (avoid)

### Shared Logic
- Payment creation logic shared between:
  - `PaymentService.CreatePaymentForTenant()`
  - `PaymentTransactionService.getOrCreateCurrentPayment()`
  
**Solution:** Keep shared method in `PaymentService`, `PaymentTransactionService` calls it

---

## 📊 Size Comparison

### Before:
- `RentalHandler`: 593 lines
- `PaymentService`: 665 lines

### After (Estimated):
- `DashboardHandler`: ~100 lines
- `UnitHandler`: ~150 lines
- `OwnerTenantHandler`: ~200 lines
- `PaymentHandler`: ~150 lines
- `TransactionHandler`: ~150 lines
- `PaymentService`: ~250 lines
- `PaymentQueryService`: ~150 lines
- `PaymentTransactionService`: ~200 lines
- `PaymentHistoryService`: ~150 lines
- `DashboardService`: ~100 lines

**Total:** ~1,600 lines (vs 1,258 before)
- Increase is due to separation overhead (struct definitions, constructors)
- BUT: Each file is focused and manageable
- Trade-off: More files, but easier to navigate and maintain

---

## 🚀 Implementation Priority

### High Priority (Do First)
1. ✅ Extract `PaymentQueryService` - Simplest, highest benefit
2. ✅ Extract `TransactionHandler` - Clear separation, frequently used

### Medium Priority
3. Extract `PaymentTransactionService`
4. Extract `PaymentHistoryService`
5. Split handlers (starting with `TransactionHandler`)

### Low Priority
6. Create `DashboardService`
7. Complete handler split
8. Remove deprecated `RentalHandler`

---

## 📁 Proposed File Structure

```
internal/
├── handlers/
│   ├── dashboard_handler.go        (NEW)
│   ├── unit_handler.go              (NEW)
│   ├── owner_tenant_handler.go      (NEW)
│   ├── payment_handler.go           (NEW)
│   ├── transaction_handler.go       (NEW)
│   ├── rental_handler.go            (DEPRECATED - remove after migration)
│   ├── tenant_handler.go            (Keep - tenant self-service)
│   └── auth_handler.go              (Keep)
│
├── service/
│   ├── payment_service.go           (Refactored - core CRUD)
│   ├── payment_query_service.go     (NEW)
│   ├── payment_transaction_service.go (NEW)
│   ├── payment_history_service.go  (NEW)
│   ├── dashboard_service.go         (NEW)
│   ├── tenant_service.go            (Keep)
│   ├── unit_service.go              (Keep)
│   └── auth_service.go              (Keep)
```

---

## 🎯 Success Metrics

After restructuring, we should achieve:
1. ✅ No handler > 250 lines
2. ✅ No service > 300 lines (except core PaymentService if needed)
3. ✅ Each service has clear, single responsibility
4. ✅ Dependencies flow in one direction (no cycles)
5. ✅ Easy to locate functionality (within 2 files)
6. ✅ All tests passing
7. ✅ No breaking API changes

---

## 🔧 Example: Migration Snippet

### Before (PaymentService):
```go
// payment_service.go (665 lines)
func (s *PaymentService) GetPaymentSummary() (*PaymentSummary, error) {
    payments, err := s.paymentRepo.GetAllPayments()
    // ... aggregation logic
}

func (s *PaymentService) VerifyTransaction(...) {
    // ... verification logic
}
```

### After:
```go
// payment_query_service.go (150 lines)
func (s *PaymentQueryService) GetPaymentSummary() (*PaymentSummary, error) {
    payments, err := s.paymentRepo.GetAllPayments()
    // ... aggregation logic
}

// payment_transaction_service.go (200 lines)
func (s *PaymentTransactionService) VerifyTransaction(...) {
    // ... verification logic
}
```

---

**Ready to proceed?** Start with Phase 1 (PaymentQueryService) - it's the safest and provides immediate benefit.


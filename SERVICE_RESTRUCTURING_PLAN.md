# Service Layer Restructuring Plan

> **⚠️ REVIEW COMPLETE:** After thorough analysis of dependencies and code flow, this plan reflects actual architecture patterns and real issues found.

## 🔍 Analysis Summary

### Current Architecture Flow Analyzed:
- ✅ Handlers → Services → Repositories pattern confirmed
- ✅ Service dependencies traced through actual code
- ✅ Code flow patterns identified
- ✅ Real issues vs. theoretical issues separated

---

## ⚠️ REAL Issues Found (Not Theoretical)

### 1. **TenantService Bypasses PaymentService** (CRITICAL)
**Location:** `tenant_service.go:96`, `tenant_service.go:160-167`

**Problem:**
- `createFirstPayment()` uses `paymentRepo.CreatePayment()` directly
- `MoveOutTenant()` uses `paymentRepo` methods directly
- Duplicates payment creation logic from PaymentService

**Impact:**
- Payment business logic exists in 2 places
- Changes must be made in multiple places
- Inconsistent payment creation

**Evidence:**
```go
// TenantService.createFirstPayment() - line 96
return s.paymentRepo.CreatePayment(payment)  // ❌ Direct repo usage

// PaymentService.getOrCreateCurrentPayment() - line 364
s.paymentRepo.CreatePayment(payment)  // ✅ Same logic, different place
```

---

### 2. **VerifyTransaction Inefficient Query** (HIGH PRIORITY)
**Location:** `payment_service.go:275-315`

**Problem:**
- Loops through ALL payments to find one transaction
- Should query transaction first to get payment_id

**Current (Inefficient - O(n)):**
```go
payments, _ := s.paymentRepo.GetAllPayments()  // Gets ALL payments
for _, p := range payments {
    txs, _ := s.paymentRepo.GetPaymentTransactionsByPaymentID(p.ID)
    for _, tx := range txs {
        if tx.TransactionID == transactionID {
            paymentID = p.ID
        }
    }
}
```

**Should be (Efficient - O(1)):**
```go
// Query transaction directly
tx, err := s.paymentRepo.GetTransactionByID(transactionID)
paymentID = tx.PaymentID
```

---

### 3. **Duplicated Payment Creation Logic** (MEDIUM)
**Location:** 
- `TenantService.createFirstPayment()` (lines 66-96)
- `PaymentService.getOrCreateCurrentPayment()` (lines 319-369)

**Problem:** Both methods:
- Calculate due date based on PaymentDueDay
- Create Payment with MonthlyRent
- Set default values (AmountPaid=0, RemainingBalance=Amount)

**Difference:** Only in date calculation (move-in date vs current date)

---

## ✅ What's Actually Good (No Changes Needed)

### PaymentService Dependencies Are Valid
- **tenantRepo/unitRepo needed**: PaymentService must fetch rent info when creating payments
- **Pattern**: `tenantID → tenant → unitID → unit → MonthlyRent`
- **This is correct architecture** - PaymentService needs this info for business logic

### PaymentService Size (414 lines)
- All methods are payment-related (cohesive)
- Dependencies make sense
- Size is acceptable for Go services
- **Not a real issue** - splitting would add complexity without clear benefit

---

---

## 🎯 Recommended Restructuring Plan (Based on Real Issues)

### Phase 1: Critical Fixes (DO FIRST) ✅

**Priority 1: Fix TenantService to Use PaymentService**

**Problem:** TenantService bypasses PaymentService business logic
**Solution:** Inject PaymentService, remove paymentRepo dependency

**Changes Needed:**
```go
// internal/service/tenant_service.go
type TenantService struct {
    tenantRepo     interfaces.TenantRepository
    unitRepo       interfaces.UnitRepository
    paymentService *PaymentService  // ✅ NEW - replace paymentRepo
}

func NewTenantService(
    tenantRepo interfaces.TenantRepository,
    unitRepo interfaces.UnitRepository,
    paymentService *PaymentService,  // ✅ Changed from paymentRepo
) *TenantService {
    return &TenantService{
        tenantRepo:     tenantRepo,
        unitRepo:       unitRepo,
        paymentService: paymentService,
    }
}

// createFirstPayment() - Use PaymentService method
func (s *TenantService) createFirstPayment(tenant *domain.Tenant) error {
    // Option 1: Call PaymentService.CreateFirstPaymentForTenant() (needs new method)
    // Option 2: Use existing PaymentService.getOrCreateCurrentPayment() with move-in date logic
    // Option 3: Create shared helper in PaymentService
}
```

**Also Update:** `main.go:64` - Change dependency injection

---

**Priority 2: Fix VerifyTransaction Efficiency**

**Add to PaymentRepository Interface:**
```go
GetTransactionByID(transactionID string) (*domain.PaymentTransaction, error)
```

**Implement in Postgres Repository:**
```go
// Query transaction by transaction_id directly
SELECT id, payment_id, transaction_id, ...
FROM payment_transactions
WHERE transaction_id = $1
```

**Refactor VerifyTransaction:**
```go
// Before (inefficient)
payments, _ := s.paymentRepo.GetAllPayments()
for _, p := range payments { ... }

// After (efficient)
tx, err := s.paymentRepo.GetTransactionByID(transactionID)
if err != nil || tx == nil {
    return fmt.Errorf("transaction not found")
}
paymentID = tx.PaymentID
```

---

**Priority 3: Extract Shared Payment Creation Helper**

**Create in PaymentService:**
```go
// CreatePaymentForTenant creates a payment for a tenant with given due date
func (s *PaymentService) CreatePaymentForTenant(
    tenantID int,
    dueDate time.Time,
    amount int,
) (*domain.Payment, error) {
    // Shared logic for creating payments
    // Used by both createFirstPayment and getOrCreateCurrentPayment
}
```

---

### Phase 2: Optional Improvements

### Option A: Split PaymentService (NOT RECOMMENDED - YET) ⚠️

```
PaymentService (Core Payments)
├── Payment CRUD operations
├── Payment queries
└── Payment lifecycle (auto-create next)

PaymentTransactionService (NEW)
├── SubmitPaymentIntent
├── VerifyTransaction
└── Transaction queries

PaymentStatusService (NEW - Optional)
├── GetOverduePayments
├── GetPendingPayments
├── GetPaymentSummary
└── Status calculation logic
```

**Benefits:**
- Clear separation of concerns
- Easier to test
- Each service has single responsibility
- Maintains backward compatibility (PaymentService delegates to new services)

### Option B: Extract Transaction Service Only

Keep PaymentService as-is, but extract transaction logic:

```
PaymentService (stays as main service)
├── Payment operations
└── Delegates to PaymentTransactionService

PaymentTransactionService (NEW)
├── Transaction submission
├── Transaction verification
└── Transaction queries
```

**Benefits:**
- Minimal changes
- Transaction logic isolated
- Easier to extend transaction features

---

## ✅ Recommended Implementation (Updated Plan)

### Phase 1: Fix TenantService (CRITICAL - DO FIRST)

**File:** `internal/service/tenant_service.go`

**Step 1: Update Constructor**
```go
func NewTenantService(
    tenantRepo interfaces.TenantRepository,
    unitRepo interfaces.UnitRepository,
    paymentService *PaymentService,  // ✅ Changed from paymentRepo
) *TenantService {
    return &TenantService{
        tenantRepo:     tenantRepo,
        unitRepo:       unitRepo,
        paymentService: paymentService,  // ✅ Changed
    }
}
```

**Step 2: Update createFirstPayment()**
```go
func (s *TenantService) createFirstPayment(tenant *domain.Tenant) error {
    // Get unit for rent info
    unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
    if err != nil {
        return fmt.Errorf("unit not found: %w", err)
    }

    // Calculate first due date
    moveInDate := tenant.MoveInDate
    firstDueDate := time.Date(moveInDate.Year(), moveInDate.Month(), 
        unit.PaymentDueDay, 0, 0, 0, 0, moveInDate.Location())
    
    if moveInDate.Day() > unit.PaymentDueDay {
        firstDueDate = firstDueDate.AddDate(0, 1, 0)
    }

    // ✅ Use PaymentService method (need to add this method)
    // Option: Add CreatePaymentWithDueDate() to PaymentService
    // OR: Use PaymentService.getOrCreateCurrentPayment() if it accepts dueDate param
    return s.paymentService.CreatePaymentForTenant(
        tenant.ID,
        tenant.UnitID,
        firstDueDate,
        unit.MonthlyRent,
    )
}
```

**Step 3: Update MoveOutTenant()**
```go
func (s *TenantService) MoveOutTenant(tenantID int) error {
    tenant, err := s.tenantRepo.GetTenantByID(tenantID)
    if err != nil {
        return fmt.Errorf("tenant not found: %w", err)
    }

    // ✅ Use PaymentService to delete unpaid payments
    // Option: Add DeleteUnpaidPaymentsByTenantID() to PaymentService
    unpaid, err := s.paymentService.GetUnpaidPaymentsByTenantID(tenantID)
    if err == nil {
        for _, payment := range unpaid {
            if err := s.paymentService.DeletePayment(payment.ID); err != nil {
                fmt.Printf("Warning: Failed to delete payment %d: %v\n", payment.ID, err)
            }
        }
    }

    // Rest of the method...
}
```

**Step 4: Update main.go**
```go
// Before
tenantService := service.NewTenantService(tenantRepo, unitRepo, paymentRepo)

// After
paymentService := service.NewPaymentService(paymentRepo, tenantRepo, unitRepo)
tenantService := service.NewTenantService(tenantRepo, unitRepo, paymentService)  // ✅
```

---

### Phase 2: Fix VerifyTransaction Efficiency (HIGH PRIORITY)

**Step 1: Add Method to PaymentRepository Interface**
```go
// internal/repository/interfaces/payment_repository.go
GetTransactionByID(transactionID string) (*domain.PaymentTransaction, error)
```

**Step 2: Implement in Postgres Repository**
```go
// internal/repository/postgres/postgres_payment_repository_transactions.go
func (r *PostgresPaymentRepository) GetTransactionByID(transactionID string) (*domain.PaymentTransaction, error) {
    query := `
        SELECT id, payment_id, transaction_id, amount, submitted_at, 
               verified_at, verified_by_user_id, notes, created_at
        FROM payment_transactions
        WHERE transaction_id = $1`
    
    tx := &domain.PaymentTransaction{}
    // ... scan logic
    return tx, nil
}
```

**Step 3: Refactor VerifyTransaction**
```go
// internal/service/payment_service.go
func (s *PaymentService) VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error {
    // ✅ Query transaction directly (efficient)
    tx, err := s.paymentRepo.GetTransactionByID(transactionID)
    if err != nil || tx == nil {
        return fmt.Errorf("transaction not found")
    }
    
    paymentID := tx.PaymentID
    
    // Rest of verification logic...
}
```

---

### Phase 3: Extract Shared Payment Helper (OPTIONAL - REDUCE DUPLICATION)

**Add to PaymentService:**
```go
// CreatePaymentForTenant creates a payment with explicit parameters
// Used by both TenantService (first payment) and PaymentService (auto-create)
func (s *PaymentService) CreatePaymentForTenant(
    tenantID int,
    unitID int,
    dueDate time.Time,
    amount int,
) (*domain.Payment, error) {
    payment := &domain.Payment{
        TenantID:         tenantID,
        UnitID:           unitID,
        Amount:           amount,
        AmountPaid:       0,
        RemainingBalance: amount,
        DueDate:          dueDate,
        IsPaid:           false,
        IsFullyPaid:      false,
        PaymentMethod:    "UPI",
        UPIID:            "9848790200@ybl",
    }
    
    return payment, s.paymentRepo.CreatePayment(payment)
}
```

**Benefits:**
- Eliminates duplication between TenantService and PaymentService
- Single place for payment creation logic
- Easier to maintain

---

## Implementation Details

### 1. PaymentTransactionService Structure

```go
type PaymentTransactionService struct {
    paymentRepo       interfaces.PaymentRepository
    paymentService    *PaymentService  // For payment operations
    tenantRepo        interfaces.TenantRepository
    unitRepo          interfaces.UnitRepository
}

// SubmitPaymentIntent creates transaction record
func (s *PaymentTransactionService) SubmitPaymentIntent(...)

// VerifyTransaction verifies and updates payment
func (s *PaymentTransactionService) VerifyTransaction(...)

// GetPendingVerifications gets pending transactions
func (s *PaymentTransactionService) GetPendingVerifications(...)
```

### 2. PaymentService Simplified

```go
type PaymentService struct {
    paymentRepo  interfaces.PaymentRepository
    tenantRepo   interfaces.TenantRepository
    unitRepo     interfaces.UnitRepository
    // Transaction service injected for complex operations
    txService    *PaymentTransactionService  // Optional
}

// Core payment operations remain
// Transaction methods delegate to PaymentTransactionService
func (s *PaymentService) SubmitPaymentIntent(...) {
    return s.txService.SubmitPaymentIntent(...)
}
```

### 3. Service Dependencies

```
main.go:
  ├── PaymentService (core)
  ├── PaymentTransactionService (uses PaymentService)
  └── TenantService (uses PaymentService for first payment)

Handler layer:
  ├── RentalHandler (PaymentService, PaymentTransactionService)
  └── TenantHandler (PaymentService)
```

---

## Migration Strategy

### Step 1: Create PaymentTransactionService
- Move transaction methods from PaymentService
- Keep PaymentService methods as wrappers (backward compatibility)
- Update handlers to use new service

### Step 2: Update Dependency Injection
- Update `main.go` to create PaymentTransactionService
- Pass to handlers that need it
- Keep PaymentService for existing code

### Step 3: Update TenantService
- Remove paymentRepo dependency
- Inject PaymentService
- Use PaymentService for first payment creation

### Step 4: Test & Verify
- All existing functionality works
- No breaking changes
- Cleaner code structure

---

## File Structure After Restructuring

```
internal/service/
├── auth_service.go                 (unchanged)
├── unit_service.go                 (unchanged)
├── tenant_service.go               (refactored - uses PaymentService instead of paymentRepo)
├── payment_service.go              (same file, efficiency improvements)
└── (No new files needed - keep PaymentService as-is for now)
```

**Changes Summary:**
- ✅ TenantService: Remove paymentRepo, add PaymentService dependency
- ✅ PaymentService: Add GetTransactionByID() method for efficiency
- ✅ PaymentService: Add CreatePaymentForTenant() helper method
- ❌ **NO splitting** - PaymentService dependencies are valid

---

## Benefits of Recommended Changes

✅ **Consistent Payment Logic**: TenantService uses PaymentService (no bypass)  
✅ **Performance**: VerifyTransaction optimized (O(1) instead of O(n))  
✅ **DRY Principle**: Shared payment creation helper eliminates duplication  
✅ **Maintainability**: Payment business logic in one place  
✅ **Backward Compatibility**: Existing handlers continue to work  
✅ **Valid Dependencies**: PaymentService's tenantRepo/unitRepo dependencies are correct  

---

## ⚠️ Why NOT to Split PaymentService (Yet)

### Analysis Conclusion:
1. **Size (414 lines)**: Acceptable for Go services - methods are cohesive
2. **Dependencies (tenantRepo/unitRepo)**: Valid - needed for rent info
3. **Transaction Logic**: Simple enough (submit, verify, query)
4. **Status Logic**: Simple calculations (overdue, pending, summary)
5. **Splitting Would**: Add complexity without clear benefit
6. **Better Approach**: Fix real issues first, then reassess

### When to Consider Splitting:
- Transaction logic grows complex (payment methods, refunds, etc.)
- Status calculations become complex (reporting, analytics)
- PaymentService exceeds 800+ lines
- Team requests it for organizational reasons

---

## Optional: Further Improvements

1. **Payment Calculator Helper**
   - Extract due date calculation logic
   - Extract balance calculation logic
   - Reusable across services

2. **Payment Validator**
   - Validate payment amounts
   - Validate payment dates
   - Business rule validation

3. **Service Interfaces**
   - Define interfaces for services
   - Enable easier testing and swapping implementations


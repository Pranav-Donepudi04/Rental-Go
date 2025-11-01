# Architecture Dependency & Flow Analysis

## 🔍 Current Dependency Graph

```
main.go
│
├─ Repositories (Data Access Layer)
│  ├─ unitRepo (UnitRepository)
│  ├─ tenantRepo (TenantRepository)
│  ├─ paymentRepo (PaymentRepository)
│  ├─ userRepo (UserRepository)
│  └─ sessionRepo (SessionRepository)
│
├─ Services (Business Logic Layer)
│  ├─ unitService (UnitService)
│  │  └─ Depends on: unitRepo
│  │
│  ├─ tenantService (TenantService) ⚠️
│  │  └─ Depends on: tenantRepo, unitRepo, paymentRepo (DIRECT - ISSUE)
│  │
│  ├─ paymentService (PaymentService) ⚠️
│  │  └─ Depends on: paymentRepo, tenantRepo, unitRepo
│  │     └─ Uses tenantRepo/unitRepo to fetch rent info when creating payments
│  │
│  └─ authService (AuthService)
│     └─ Depends on: userRepo, sessionRepo
│
└─ Handlers (Presentation Layer)
   ├─ rentalHandler (RentalHandler)
   │  └─ Depends on: unitService, tenantService, paymentService, authService
   │
   ├─ tenantHandler (TenantHandler)
   │  └─ Depends on: tenantService, paymentService, userRepo, authService
   │
   └─ authHandler (AuthHandler)
      └─ Depends on: authService
```

---

## 📊 Service Method Usage Analysis

### PaymentService (414 lines) - Current Responsibilities

| Method | Uses tenantRepo? | Uses unitRepo? | Purpose |
|--------|-----------------|----------------|---------|
| `CreateMonthlyPayment` | ✅ Yes (line 29) | ✅ Yes (line 34) | Gets tenant → unit → MonthlyRent for payment amount |
| `MarkPaymentAsPaid` | ❌ No | ❌ No | Updates payment status |
| `GetPaymentByID` | ✅ Yes (line 104) | ✅ Yes (line 111) | Loads tenant/unit for display |
| `GetPaymentsByTenantID` | ❌ No | ❌ No | Direct repo call |
| `GetOverduePayments` | ❌ No | ❌ No | Filters payments by status |
| `GetPendingPayments` | ❌ No | ❌ No | Filters payments by status |
| `GetPaymentSummary` | ❌ No | ❌ No | Aggregates payment data |
| `GetAllPayments` | ❌ No | ❌ No | Direct repo call |
| `GetPendingVerifications` | ❌ No | ❌ No | Direct repo call |
| `SubmitPaymentIntent` | ✅ Yes (indirect via getOrCreateCurrentPayment) | ✅ Yes (indirect) | Gets/creates payment, needs unit info |
| `VerifyTransaction` | ❌ No | ❌ No | ⚠️ INEFFICIENT: Loops all payments to find transaction |
| `getOrCreateCurrentPayment` | ✅ Yes (line 332) | ✅ Yes (line 337) | Gets tenant → unit → MonthlyRent |
| `CreateNextPayment` | ❌ No | ❌ No | Uses existing payment data |
| `autoCreateNextPayment` | ❌ No | ❌ No | Checks/creates next payment |

**PaymentService Dependencies Analysis:**
- **tenantRepo/unitRepo usage**: Necessary - PaymentService needs to know rent amount when creating payments
- **Reason**: Payment amount comes from `unit.MonthlyRent`, not from tenant
- **Pattern**: `tenantID → tenant → unitID → unit → MonthlyRent`

---

### TenantService (237 lines) - Current Responsibilities

| Method | Uses paymentRepo? | Should Use PaymentService? |
|--------|-------------------|----------------------------|
| `CreateTenant` | ✅ Yes (via createFirstPayment) | ✅ YES - Should use PaymentService |
| `createFirstPayment` | ✅ Yes (line 96) | ✅ YES - Duplicates payment creation logic |
| `MoveOutTenant` | ✅ Yes (lines 160, 167) | ✅ YES - Should use PaymentService |
| `GetTenantByID` | ❌ No | ❌ No |
| `GetAllTenants` | ❌ No | ❌ No |
| Others | ❌ No | ❌ No |

**TenantService Issues:**
1. **Direct paymentRepo usage** - Bypasses PaymentService business logic
2. **Duplicated payment creation logic** - `createFirstPayment` duplicates `getOrCreateCurrentPayment`
3. **Inconsistent** - Should use PaymentService like handlers do

---

### Handler Usage Analysis

#### RentalHandler Methods:

| Handler Method | Services Called | Payment Methods Used |
|----------------|----------------|----------------------|
| `Dashboard` | unitService, tenantService, paymentService | GetAllPayments, GetPaymentSummary |
| `GetUnits` | unitService | - |
| `GetTenants` | tenantService | - |
| `CreateTenant` | tenantService | (indirect - creates payment) |
| `GetPayments` | paymentService | GetAllPayments |
| `MarkPaymentAsPaid` | paymentService | VerifyTransaction, MarkPaymentAsPaid |
| `VacateTenant` | tenantService | (indirect - deletes payments) |
| `GetSummary` | unitService, paymentService | GetPaymentSummary |
| `UnitDetails` | unitService, tenantService, paymentService | GetPaymentsByTenantID, GetPendingVerifications |
| `GetPendingVerifications` | paymentService | GetPendingVerifications |

#### TenantHandler Methods:

| Handler Method | Services Called | Payment Methods Used |
|----------------|----------------|----------------------|
| `Me` | tenantService, paymentService | GetPaymentsByTenantID |
| `SubmitPayment` | paymentService | SubmitPaymentIntent |
| `ChangePassword` | authService, userRepo | - |

---

## 🔄 Code Flow Examples

### Flow 1: Tenant Submits Transaction ID

```
TenantHandler.SubmitPayment()
  ↓
paymentService.SubmitPaymentIntent(tenantID, txnID)
  ↓
paymentService.getOrCreateCurrentPayment(tenantID)
  │
  ├─> paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
  │   └─> If found: return unpaid[0]
  │
  └─> If not found:
      ├─> tenantRepo.GetTenantByID(tenantID)  ← NEEDS tenantRepo
      ├─> unitRepo.GetUnitByID(tenant.UnitID)  ← NEEDS unitRepo
      └─> paymentRepo.CreatePayment()  ← Creates with MonthlyRent
  ↓
paymentRepo.GetTransactionByPaymentAndID(paymentID, txnID)
  ↓
paymentRepo.CreatePaymentTransaction(tx)
```

**Why tenantRepo/unitRepo needed?**
- PaymentService doesn't know the rent amount
- Must fetch: tenant → unit → MonthlyRent

---

### Flow 2: Owner Verifies Transaction

```
RentalHandler.MarkPaymentAsPaid()
  ↓
paymentService.VerifyTransaction(transactionID, amount, userID)
  ↓
⚠️ INEFFICIENT LOOP (lines 281-293):
  ├─> paymentRepo.GetAllPayments()  ← Gets ALL payments
  ├─> For each payment:
  │   └─> paymentRepo.GetPaymentTransactionsByPaymentID(payment.ID)
  │       └─> Check if transactionID matches
  └─> If found: paymentID = p.ID
  ↓
paymentRepo.VerifyTransaction(transactionID, amount, userID)
  │   └─> Updates transaction + payment in DB transaction
  ↓
paymentRepo.GetPaymentByID(paymentID)
  ↓
If IsFullyPaid:
  └─> paymentService.autoCreateNextPayment(payment)
      └─> paymentService.CreateNextPayment(payment)
```

**Issue:** VerifyTransaction loops through ALL payments to find one transaction!

---

### Flow 3: Create Tenant (with first payment)

```
RentalHandler.CreateTenant()
  ↓
tenantService.CreateTenant(tenant)
  │
  ├─> tenantRepo.CreateTenant()
  ├─> unitRepo.UpdateUnitOccupancy()
  └─> tenantService.createFirstPayment(tenant)  ← DIRECT paymentRepo usage
      │
      ├─> unitRepo.GetUnitByID()  ← Gets unit info
      ├─> Calculate due date
      └─> paymentRepo.CreatePayment()  ⚠️ BYPASSES PaymentService
```

**Issue:** TenantService directly uses paymentRepo instead of PaymentService

---

## ⚠️ Identified Issues

### Issue 1: TenantService Bypasses PaymentService
**Location:** `tenant_service.go:96`, `tenant_service.go:160, 167`
- `createFirstPayment()` uses `paymentRepo.CreatePayment()` directly
- `MoveOutTenant()` uses `paymentRepo` methods directly
- Should use PaymentService to ensure business logic consistency

**Impact:** 
- Payment creation logic duplicated
- Changes to payment creation must be made in 2 places
- Potential inconsistency

---

### Issue 2: VerifyTransaction Inefficient Query
**Location:** `payment_service.go:275-315`
- Loops through ALL payments to find transaction
- Should query transaction first to get payment_id directly

**Current (Inefficient):**
```go
payments, _ := s.paymentRepo.GetAllPayments()
for _, p := range payments {
    txs, _ := s.paymentRepo.GetPaymentTransactionsByPaymentID(p.ID)
    for _, tx := range txs {
        if tx.TransactionID == transactionID {
            paymentID = p.ID
        }
    }
}
```

**Should be:**
```go
// Query transaction first to get payment_id
tx, err := s.paymentRepo.GetTransactionByID(transactionID)
if err != nil || tx == nil {
    return fmt.Errorf("transaction not found")
}
paymentID = tx.PaymentID
```

**Impact:** Performance issue - O(n) where n = total payments

---

### Issue 3: PaymentService Size (414 lines)
**Responsibilities:**
1. Payment CRUD operations ✅
2. Payment queries ✅
3. Status calculations (overdue, pending, summary) ✅
4. Transaction submission/verification ✅
5. Payment lifecycle (auto-create next) ✅
6. Helper methods (getOrCreateCurrentPayment) ✅

**Analysis:**
- Methods are logically related (all payment-related)
- Dependencies make sense (needs tenant/unit info for rent)
- BUT: Could be split for better separation

---

### Issue 4: Duplicated Payment Creation Logic
**Location:** 
- `TenantService.createFirstPayment()` (lines 66-96)
- `PaymentService.getOrCreateCurrentPayment()` (lines 319-369)

**Similarities:**
- Both calculate due date based on PaymentDueDay
- Both create Payment with MonthlyRent
- Both set default values (AmountPaid=0, RemainingBalance=Amount)

**Difference:**
- `createFirstPayment`: Uses move-in date
- `getOrCreateCurrentPayment`: Uses current date

**Impact:** Logic duplication, maintenance burden

---

## ✅ What's Actually Good

### 1. PaymentService Dependencies Are Valid
- **tenantRepo/unitRepo needed**: PaymentService must fetch rent info when creating payments
- **Pattern**: Payment amount = `unit.MonthlyRent`, not stored in tenant
- **Solution**: This dependency is correct and necessary

### 2. Handler Separation is Clean
- RentalHandler: Owner operations
- TenantHandler: Tenant operations
- Clear separation of concerns

### 3. Repository Layer is Clean
- Each repository handles one entity
- Interfaces enable testing
- No cross-repository dependencies

---

## 🎯 Restructuring Recommendations

Based on **actual code flow analysis**, here's what should be done:

### Priority 1: Fix TenantService (Critical)
**Problem:** Direct paymentRepo usage bypasses PaymentService
**Solution:** Inject PaymentService into TenantService

**Changes:**
```go
// Before
type TenantService struct {
    tenantRepo  interfaces.TenantRepository
    unitRepo    interfaces.UnitRepository
    paymentRepo interfaces.PaymentRepository  // ❌ Remove
}

// After
type TenantService struct {
    tenantRepo     interfaces.TenantRepository
    unitRepo       interfaces.UnitRepository
    paymentService *PaymentService  // ✅ Add
}
```

**Impact:** 
- `createFirstPayment()` → Use PaymentService method
- `MoveOutTenant()` → Use PaymentService method
- Eliminates duplication
- Ensures consistency

---

### Priority 2: Fix VerifyTransaction Efficiency (High)
**Problem:** Loops all payments to find transaction
**Solution:** Query transaction first

**Add to PaymentRepository:**
```go
GetTransactionByID(transactionID string) (*PaymentTransaction, error)
```

**Refactor VerifyTransaction:**
```go
// Before: Loop all payments
payments, _ := s.paymentRepo.GetAllPayments()
for _, p := range payments { ... }

// After: Direct query
tx, err := s.paymentRepo.GetTransactionByID(transactionID)
paymentID = tx.PaymentID
```

---

### Priority 3: Extract Transaction Service (Optional)
**Problem:** PaymentService handles transactions + payments
**Benefit:** Clear separation, but adds complexity

**Decision:** 
- If splitting: Create PaymentTransactionService
- If keeping: Document that PaymentService handles both

**Recommendation:** Keep together for now (transactions tightly coupled to payments)

---

### Priority 4: Extract Status Service (Optional)
**Problem:** Status calculation mixed with data operations
**Benefit:** Reusable status logic

**Decision:** 
- Low priority - status methods are simple
- If splitting: Create PaymentStatusService
- Better alternative: Keep in PaymentService but group methods

---

## 📋 Recommended Restructuring Plan

### Phase 1: Critical Fixes (DO FIRST) ✅

1. **Fix TenantService Dependencies**
   - Remove paymentRepo from TenantService
   - Inject PaymentService
   - Refactor `createFirstPayment()` to use PaymentService
   - Refactor `MoveOutTenant()` to use PaymentService

2. **Fix VerifyTransaction Efficiency**
   - Add `GetTransactionByID()` to PaymentRepository
   - Refactor `VerifyTransaction()` to use direct query

3. **Extract Payment Creation Helper**
   - Create shared method for payment creation
   - Used by both TenantService and PaymentService
   - Eliminates duplication

### Phase 2: Optional Improvements

4. **Extract PaymentTransactionService** (Optional)
   - Only if transaction logic grows complex
   - For now, transactions are simple enough to keep together

5. **Group PaymentService Methods** (Documentation)
   - Document method groupings
   - Add comments for each responsibility section

---

## 🔗 Dependency Flow After Restructuring

```
TenantService
  ├─> tenantRepo (direct)
  ├─> unitRepo (direct - for occupancy)
  └─> paymentService (NEW) ← Uses PaymentService instead of paymentRepo
      │
      └─> PaymentService
          ├─> paymentRepo (direct)
          ├─> tenantRepo (for rent info) ← Valid dependency
          └─> unitRepo (for rent info) ← Valid dependency
```

---

## ✅ Final Recommendation

**DO:**
1. ✅ Fix TenantService to use PaymentService
2. ✅ Fix VerifyTransaction inefficiency
3. ✅ Extract shared payment creation helper

**DON'T Split Yet:**
- ❌ Don't split PaymentService yet (dependencies are valid)
- ❌ Don't create PaymentTransactionService yet (transactions are simple)
- ❌ Don't create PaymentStatusService yet (status logic is simple)

**REASONING:**
- PaymentService size (414 lines) is acceptable for Go services
- Dependencies (tenantRepo/unitRepo) are necessary and valid
- Transaction logic is simple (submit, verify, query)
- Splitting adds complexity without clear benefit
- Better to fix actual issues first (TenantService, efficiency)

---

## 🎯 Clear Action Plan

**Step 1:** Fix TenantService (highest priority)
**Step 2:** Fix VerifyTransaction efficiency
**Step 3:** Extract shared payment helper
**Step 4:** Review if splitting is still needed after fixes


# Partial Payment System - Complete Implementation Plan

## 📋 Overview

This plan implements partial payment tracking, transaction management, and auto-creation of next payments for the rental management system.

**Key Features:**
- ✅ Partial payment tracking (amount_paid, remaining_balance)
- ✅ Separate transaction ID management
- ✅ Owner verification with amount entry
- ✅ Auto-create next payment when fully paid
- ✅ Due date stays fixed until fully paid, then resets to `due_date + 1 month`

---

## 🎯 Confirmed Requirements

### Payment Logic:
1. **Move-in**: Account created → Immediately create first payment (due date = next 10th >= move-in date)
2. **Due Date**: Stays fixed (e.g., Aug 10) until fully paid, then resets to `due_date + 1 month` (Sept 10)
3. **Partial Payments**: Track cumulative amount_paid, calculate remaining_balance
4. **Transactions**: Separate table linking transaction IDs to amounts
5. **Owner Verification**: Owner enters amount when verifying transaction
6. **Auto-Create**: Next payment created immediately when current payment fully paid
7. **Vacate**: Delete all future payments when tenant vacates

### Payment Cycle:
- **1 month** payment cycle (Aug 10 → Sept 10)
- Due date calculation: Always `current_payment.due_date + 1 month` (not fully_paid_date)

---

## 📊 Database Schema Changes

### Step 1: Run SQL Migration

**File:** `migrations/001_add_partial_payment_support.sql`

**Changes:**
1. Add 4 columns to `payments` table:
   - `amount_paid INT DEFAULT 0`
   - `remaining_balance INT DEFAULT 0`
   - `is_fully_paid BOOLEAN DEFAULT FALSE`
   - `fully_paid_date TIMESTAMP NULL`

2. Create `payment_transactions` table:
   - `id`, `payment_id`, `transaction_id`, `amount` (NULL until verified)
   - `submitted_at`, `verified_at`, `verified_by_user_id`

3. Backfill existing data:
   - Paid payments: `amount_paid = amount`, `is_fully_paid = TRUE`
   - Unpaid: `amount_paid = 0`, `remaining_balance = amount`

**Action:** Run migration in Neon console

---

## 🔧 Code Implementation

### Phase 1: Domain Models

#### 1.1 Update `Payment` Domain Model

**File:** `internal/domain/payment.go`

**Add fields:**
```go
type Payment struct {
    // ... existing fields ...
    AmountPaid       int        `json:"amount_paid" db:"amount_paid"`
    RemainingBalance int        `json:"remaining_balance" db:"remaining_balance"`
    IsFullyPaid      bool       `json:"is_fully_paid" db:"is_fully_paid"`
    FullyPaidDate    *time.Time `json:"fully_paid_date" db:"fully_paid_date"`
}

// Add helper methods
func (p *Payment) RecalculateBalance() {
    p.RemainingBalance = p.Amount - p.AmountPaid
    p.IsFullyPaid = (p.RemainingBalance <= 0)
}

func (p *Payment) GetPaymentStatus() string {
    if p.IsFullyPaid {
        return "Fully Paid"
    }
    if p.AmountPaid > 0 {
        return "Partially Paid"
    }
    // Check for pending verification...
    return "Pending"
}
```

#### 1.2 Create `PaymentTransaction` Domain Model

**File:** `internal/domain/payment_transaction.go` (NEW)

```go
package domain

import "time"

type PaymentTransaction struct {
    ID              int        `json:"id" db:"id"`
    PaymentID       int        `json:"payment_id" db:"payment_id"`
    TransactionID   string     `json:"transaction_id" db:"transaction_id"`
    Amount          *int       `json:"amount" db:"amount"` // NULL until verified
    SubmittedAt     time.Time  `json:"submitted_at" db:"submitted_at"`
    VerifiedAt      *time.Time `json:"verified_at" db:"verified_at"`
    VerifiedByUserID *int       `json:"verified_by_user_id" db:"verified_by_user_id"`
    Notes           string     `json:"notes" db:"notes"`
    CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

func (pt *PaymentTransaction) IsVerified() bool {
    return pt.VerifiedAt != nil && pt.Amount != nil
}
```

---

### Phase 2: Repository Layer

#### 2.1 Update `PaymentRepository` Interface

**File:** `internal/repository/interfaces/payment_repository.go`

**Add methods:**
```go
type PaymentRepository interface {
    // ... existing methods ...
    
    // NEW: Transaction methods
    CreatePaymentTransaction(tx *domain.PaymentTransaction) error
    GetPaymentTransactionsByPaymentID(paymentID int) ([]*domain.PaymentTransaction, error)
    GetPendingVerifications(tenantID int) ([]*domain.PaymentTransaction, error)
    VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error
    
    // NEW: Auto-create helpers
    GetLatestPaymentByTenantID(tenantID int) (*domain.Payment, error)
    GetUnpaidPaymentsByTenantID(tenantID int) ([]*domain.Payment, error)
}
```

#### 2.2 Implement Postgres Methods

**File:** `internal/repository/postgres/postgres_payment_repository.go`

**Update `CreatePayment`:**
- Include new fields in INSERT

**Update `GetPaymentByID`:**
- SELECT new fields

**Update `UpdatePayment`:**
- UPDATE new fields

**Add new methods:**
```go
func (r *PostgresPaymentRepository) CreatePaymentTransaction(tx *domain.PaymentTransaction) error
func (r *PostgresPaymentRepository) GetPaymentTransactionsByPaymentID(paymentID int) ([]*domain.PaymentTransaction, error)
func (r *PostgresPaymentRepository) VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error
func (r *PostgresPaymentRepository) GetLatestPaymentByTenantID(tenantID int) (*domain.Payment, error)
```

**Implementation details:**
- Use transactions for atomicity
- Handle NULL amounts properly
- Link transactions to payments

---

### Phase 3: Service Layer

#### 3.1 Update `PaymentService`

**File:** `internal/service/payment_service.go`

**Replace `SubmitPaymentIntent`:**
```go
func (s *PaymentService) SubmitPaymentIntent(tenantID int, txnID string) error {
    // Get or create current unpaid payment
    payment, err := s.getOrCreateCurrentPayment(tenantID)
    if err != nil {
        return err
    }
    
    // Check if transaction already exists
    existing, _ := s.paymentRepo.GetTransactionByID(payment.ID, txnID)
    if existing != nil {
        return nil // Already exists
    }
    
    // Create payment transaction (amount NULL until verified)
    tx := &domain.PaymentTransaction{
        PaymentID:     payment.ID,
        TransactionID: txnID,
        Amount:       nil, // NULL until owner verifies
        SubmittedAt:  time.Now(),
    }
    
    return s.paymentRepo.CreatePaymentTransaction(tx)
}
```

**Add `VerifyTransaction`:**
```go
func (s *PaymentService) VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error {
    // 1. Get transaction
    // 2. Verify it (update transaction.amount, verified_at, verified_by)
    // 3. Update payment: amount_paid += amount, recalculate balance
    // 4. If fully paid: Set fully_paid_date, auto-create next payment
}
```

**Add `CreateNextPayment`:**
```go
func (s *PaymentService) CreateNextPayment(currentPayment *domain.Payment) (*domain.Payment, error) {
    // Calculate next due date: currentPayment.DueDate + 1 month
    nextDueDate := currentPayment.DueDate.AddDate(0, 1, 0)
    
    // Create next payment
    nextPayment := &domain.Payment{
        TenantID:         currentPayment.TenantID,
        UnitID:           currentPayment.UnitID,
        Amount:           currentPayment.Amount, // Same rent amount
        DueDate:          nextDueDate,
        AmountPaid:       0,
        RemainingBalance: currentPayment.Amount,
        IsFullyPaid:      false,
        PaymentMethod:    currentPayment.PaymentMethod,
        UPIID:            currentPayment.UPIID,
    }
    
    return nextPayment, s.paymentRepo.CreatePayment(nextPayment)
}
```

**Add `getOrCreateCurrentPayment`:**
```go
func (s *PaymentService) getOrCreateCurrentPayment(tenantID int) (*domain.Payment, error) {
    // Get latest unpaid payment
    unpaid, err := s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
    if err != nil {
        return nil, err
    }
    
    // If exists and not fully paid, return it
    if len(unpaid) > 0 {
        return unpaid[0], nil
    }
    
    // Otherwise, get tenant/unit and create new payment
    // Calculate due date: Next 10th >= today
}
```

#### 3.2 Update `TenantService`

**File:** `internal/service/tenant_service.go`

**Update `CreateTenant`:**
```go
func (s *TenantService) CreateTenant(tenant *domain.Tenant) error {
    // ... existing validation ...
    
    // Create tenant
    if err := s.tenantRepo.CreateTenant(tenant); err != nil {
        return err
    }
    
    // Update unit occupancy
    if err := s.unitRepo.UpdateUnitOccupancy(tenant.UnitID, true); err != nil {
        s.tenantRepo.DeleteTenant(tenant.ID)
        return err
    }
    
    // NEW: Create first payment immediately
    if err := s.createFirstPayment(tenant); err != nil {
        // Log error but don't fail tenant creation
        // Payment can be created manually if needed
    }
    
    return nil
}

func (s *TenantService) createFirstPayment(tenant *domain.Tenant) error {
    // Get unit
    unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
    if err != nil {
        return err
    }
    
    // Calculate first due date: Next 10th >= move_in_date
    moveInDate := tenant.MoveInDate
    firstDueDate := time.Date(moveInDate.Year(), moveInDate.Month(), unit.PaymentDueDay, 0, 0, 0, 0, time.UTC)
    
    // If move-in date is after due date in same month, use next month
    if moveInDate.Day() > unit.PaymentDueDay {
        firstDueDate = firstDueDate.AddDate(0, 1, 0) // Next month
    }
    
    // Create payment
    payment := &domain.Payment{
        TenantID:         tenant.ID,
        UnitID:           tenant.UnitID,
        Amount:           unit.MonthlyRent,
        DueDate:          firstDueDate,
        AmountPaid:       0,
        RemainingBalance: unit.MonthlyRent,
        IsFullyPaid:      false,
        PaymentMethod:    "UPI",
        UPIID:            "9848790200@ybl",
    }
    
    return s.paymentRepo.CreatePayment(payment)
}
```

**Update `MoveOutTenant`:**
```go
func (s *TenantService) MoveOutTenant(tenantID int) error {
    // ... existing code ...
    
    // NEW: Delete all unpaid/future payments for this tenant
    unpaid, _ := s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
    for _, p := range unpaid {
        s.paymentRepo.DeletePayment(p.ID)
    }
    
    // ... rest of existing code ...
}
```

---

### Phase 4: Handler Layer

#### 4.1 Update `TenantHandler`

**File:** `internal/handlers/tenant_handler.go`

**Update `SubmitPayment`:** (Keep same, calls updated service)

**Update `Me`:** 
- Load payment transactions
- Show amount_paid, remaining_balance

#### 4.2 Update `RentalHandler`

**File:** `internal/handlers/rental_handler.go`

**Replace `MarkPaymentAsPaid`:**
```go
func (h *RentalHandler) MarkPaymentAsPaid(w http.ResponseWriter, r *http.Request) {
    var req struct {
        PaymentID   int    `json:"payment_id"`
        PaymentDate string `json:"payment_date"`
        Amount      int    `json:"amount"` // NEW: Amount being verified
        Notes       string `json:"notes"`
        TransactionID string `json:"transaction_id"` // NEW: Which transaction
    }
    
    // Parse request
    // Get user from context (owner)
    // Call: paymentService.VerifyTransaction(transactionID, amount, userID)
    // If fully paid after verification, auto-create next payment
}
```

**Add new handler:**
```go
func (h *RentalHandler) GetPendingVerifications(w http.ResponseWriter, r *http.Request) {
    // Get all pending verifications
    // Return JSON with transactions needing verification
}
```

---

### Phase 5: Templates

#### 5.1 Update `tenant-dashboard.html`

**File:** `templates/tenant-dashboard.html`

**Changes:**
- Show `amount_paid` and `remaining_balance`
- Show transaction status (verified/pending)
- Show "Partially Paid" status
- Display pending verification transactions

#### 5.2 Update `unit-detail.html`

**File:** `templates/unit-detail.html`

**Changes:**
- Show payment details with partial payment info
- Show pending verifications
- Update "Mark Payment" form to include:
  - Transaction ID selection
  - Amount entry field
  - Verify button

#### 5.3 Update `dashboard.html`

**File:** `templates/dashboard.html`

**Changes:**
- Show pending verifications count
- Update payment status display
- Show partial payment indicators

---

## 📝 Implementation Checklist

### Database:
- [ ] Run `migrations/001_add_partial_payment_support.sql`
- [ ] Verify migration successful
- [ ] Test backfill of existing data

### Domain Models:
- [ ] Update `Payment` struct with new fields
- [ ] Add helper methods to `Payment`
- [ ] Create `PaymentTransaction` struct

### Repository:
- [ ] Update `PaymentRepository` interface
- [ ] Implement transaction methods in postgres repository
- [ ] Update existing methods to include new fields
- [ ] Test repository methods

### Service:
- [ ] Update `SubmitPaymentIntent` to use transactions
- [ ] Implement `VerifyTransaction`
- [ ] Implement `CreateNextPayment`
- [ ] Update `CreateTenant` to create first payment
- [ ] Update `MoveOutTenant` to delete future payments
- [ ] Update payment status methods

### Handlers:
- [ ] Update `MarkPaymentAsPaid` to verify transactions
- [ ] Add `GetPendingVerifications` handler
- [ ] Update `TenantHandler.Me` to show transactions
- [ ] Update `UnitDetails` to show payment details

### Templates:
- [ ] Update tenant dashboard
- [ ] Update unit detail page
- [ ] Update owner dashboard

### Testing:
- [ ] Test partial payment flow
- [ ] Test transaction verification
- [ ] Test auto-create next payment
- [ ] Test move-in payment creation
- [ ] Test tenant vacate (delete future payments)

---

## 🔄 Key Flows

### Flow 1: Tenant Submits Transaction
```
1. Tenant submits transaction ID
2. System creates PaymentTransaction (amount = NULL)
3. Status: "Pending Verification"
```

### Flow 2: Owner Verifies Transaction
```
1. Owner sees pending transaction
2. Owner enters amount (can be partial or full)
3. System updates:
   - PaymentTransaction.amount = entered amount
   - PaymentTransaction.verified_at = now
   - Payment.amount_paid += amount
   - Payment.remaining_balance = amount - amount_paid
4. If remaining_balance = 0:
   - Set Payment.is_fully_paid = true
   - Set Payment.fully_paid_date = now
   - Auto-create next payment (due_date + 1 month)
```

### Flow 3: Tenant Moves In
```
1. Tenant created
2. System calculates: next 10th >= move_in_date
3. Create first payment immediately
4. Due date shows on tenant dashboard
```

### Flow 4: Payment Fully Paid
```
1. Last transaction verified
2. remaining_balance = 0
3. Set is_fully_paid = true
4. Calculate: next_due_date = current_payment.due_date + 1 month
5. Create next payment with next_due_date
```

---

## 🚨 Important Notes

1. **Backward Compatibility**: Keep `is_paid` field working during transition
2. **Due Date Logic**: Always use `payment.due_date + 1 month`, NOT `fully_paid_date + 1 month`
3. **Transaction Amounts**: NULL until owner verifies
4. **Auto-Create**: Only when fully paid, delete if tenant vacates
5. **Move-in Payment**: Create immediately on tenant creation

---

## 📚 Files to Modify

### New Files:
- `internal/domain/payment_transaction.go`

### Modified Files:
- `internal/domain/payment.go`
- `internal/repository/interfaces/payment_repository.go`
- `internal/repository/postgres/postgres_payment_repository.go`
- `internal/service/payment_service.go`
- `internal/service/tenant_service.go`
- `internal/handlers/rental_handler.go`
- `internal/handlers/tenant_handler.go`
- `templates/tenant-dashboard.html`
- `templates/unit-detail.html`
- `templates/dashboard.html`

---

## ✅ Success Criteria

1. ✅ Tenant can submit multiple transaction IDs
2. ✅ Owner can verify transactions with amounts
3. ✅ Partial payments tracked correctly
4. ✅ Due date stays fixed until fully paid
5. ✅ Next payment auto-created when fully paid
6. ✅ First payment created on tenant move-in
7. ✅ Future payments deleted on tenant vacate

---

**This plan provides complete implementation guidance for the partial payment system!** 🚀


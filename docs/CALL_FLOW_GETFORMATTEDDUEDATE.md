
# Call Flow: Tenant Submits Payment by Transaction ID

## 🔄 Complete Call Chain

This section traces the flow when a tenant enters a UPI transaction ID and clicks "Submit Payment".

---

## 📍 Flow Overview

```
tenant-dashboard.html (Template)
    ↓
User enters txn_id and clicks "Submit Payment"
    ↓
JavaScript fetch() to /api/payments/submit
    ↓
tenant_handler.go (Handler Layer)
    ↓ SubmitPayment() - Line 90
    ↓
payment_transaction_service.go (Service Layer)
    ↓ SubmitPaymentIntent() - Line 32
    ↓
payment_service.go (Service Layer)
    ↓ getOrCreateCurrentPayment() - Line 363
    ↓ [May query or create payment]
    ↓
postgres_payment_repository.go (Repository Layer)
    ↓ Multiple DB operations:
    ↓ 1. GetUnpaidPaymentsByTenantID() - SELECT unpaid payments
    ↓ 2. GetTenantByID() - SELECT tenant (if creating payment)
    ↓ 3. GetUnitByID() - SELECT unit (if creating payment)
    ↓ 4. CreatePayment() - INSERT payment (if needed)
    ↓ 5. GetTransactionByPaymentAndID() - SELECT check existing
    ↓ 6. CreatePaymentTransaction() - INSERT transaction
    ↓
PostgreSQL Database
    ↓ Multiple INSERT/SELECT queries
```

---

## 📋 Detailed Step-by-Step Flow: Submit Payment

### Step 1: User Enters Transaction ID
**File:** `templates/tenant-dashboard.html`  
**Lines:** 79-83, 153-178

```html
<form id="txnForm" method="post" action="/api/payments/submit">
    <label>Enter UPI Transaction ID</label>
    <input name="txn_id" placeholder="e.g., 2349ABCD..." />
    <button class="btn" type="submit">Submit Payment</button>
</form>
```

**JavaScript Handler:**
```javascript
// Intercept payment submit to avoid navigating to raw API endpoint
(function(){
    const f = document.getElementById('txnForm');
    if (!f) return;
    f.addEventListener('submit', function(e){
        e.preventDefault();  // ← Prevents default form submission
        const btn = f.querySelector('button[type=submit]');
        if (btn) btn.disabled = true;  // Disable button during request
        
        // Convert form data to URL-encoded string
        const data = new URLSearchParams(new FormData(f));
        
        // Make HTTP POST request
        fetch('/api/payments/submit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: data.toString()  // "txn_id=2349ABCD1234"
        }).then(r=>{
            if (r.status === 204) {  // HTTP 204 No Content = success
                alert('Transaction ID submitted successfully! Owner will verify...');
                f.reset();
                location.reload();  // Refresh page to show updated payments
                return; 
            }
            return r.text().then(text => { throw new Error(text || 'Failed'); });
        }).catch(err=>{
            alert('Failed to submit: ' + err.message);
            if (btn) btn.disabled = false;  // Re-enable button on error
        });
    });
})();
```

**What happens:**
1. User enters transaction ID (e.g., "2349ABCD...") in the input field
2. User clicks "Submit Payment" button
3. JavaScript **intercepts** the form submit (`e.preventDefault()`)
4. JavaScript converts form data to URL-encoded string: `"txn_id=2349ABCD1234"`
5. JavaScript sends **HTTP POST request** to `/api/payments/submit` endpoint
6. Request includes:
   - **Method:** POST
   - **URL:** `/api/payments/submit`
   - **Headers:** `Content-Type: application/x-www-form-urlencoded`
   - **Body:** `txn_id=2349ABCD1234`

---

### Step 1.5: HTTP Request Routing
**File:** `internal/http/router.go`  
**Line:** 70

```go
http.HandleFunc("/api/payments/submit", r.requireTenant(r.tenantHandler.SubmitPayment))
```

**What happens:**
1. **HTTP Server** receives POST request to `/api/payments/submit`
2. **Router** matches the route pattern `/api/payments/submit`
3. **Middleware** `requireTenant()` wraps the handler:
   - Validates session cookie ("sid")
   - Checks user role is "tenant"
   - Adds user to request context
4. **Handler** `tenantHandler.SubmitPayment()` is called with authenticated request

**Routing Chain:**
```
HTTP POST /api/payments/submit
    ↓
Router.SetupRoutes() matches "/api/payments/submit"
    ↓
requireTenant() middleware:
    - Validates session cookie
    - Checks user role = "tenant"
    - Adds user to context
    ↓
tenantHandler.SubmitPayment() called
```

---

### Step 2: Handler Validates and Processes Request
**File:** `internal/handlers/tenant_handler.go`  
**Function:** `SubmitPayment()`  
**Lines:** 90-129

```go
func (h *TenantHandler) SubmitPayment(w http.ResponseWriter, r *http.Request) {
    // Step 2.1: Validate HTTP method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Step 2.2: Validate session cookie
    c, err := r.Cookie(h.cookieName)  // Get "sid" cookie
    if err != nil {
        http.Error(w, "no cookie", http.StatusUnauthorized)
        return
    }
    
    // Step 2.3: Validate session
    sess, err := h.auth.ValidateSession(c.Value)
    if err != nil {
        http.Error(w, "invalid session", http.StatusUnauthorized)
        return
    }
    if sess == nil {
        http.Error(w, "session expired", http.StatusUnauthorized)
        return
    }
    
    // Step 2.4: Parse form data from request body
    if err := r.ParseForm(); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    
    // Step 2.5: Extract transaction ID from form
    txn := r.FormValue("txn_id")  // ← Gets "2349ABCD1234" from form
    if txn == "" {
        http.Error(w, "txn_id required", http.StatusBadRequest)
        return
    }
    
    // Step 2.6: Get user to extract tenantID
    // ⚠️  REDUNDANT: Middleware already got user and put it in context!
    // Should use: GetUserFromContext(r.Context()) instead
    user, err := h.users.GetByID(sess.UserID)
    if err != nil || user == nil || user.TenantID == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Step 2.7: Call service to submit payment intent
    if err := h.paymentTransactionService.SubmitPaymentIntent(*user.TenantID, txn); err != nil {
        http.Error(w, "failed", http.StatusBadRequest)
        return
    }
    
    // Step 2.8: Return success response
    w.WriteHeader(http.StatusNoContent)  // HTTP 204 No Content = success
}
```

**What happens:**
1. **Validates HTTP method** - Must be POST
2. **Gets session cookie** - Extracts "sid" cookie from request
3. **Validates session** - Checks if session is valid and not expired
4. **Parses form data** - Parses URL-encoded body (`txn_id=2349ABCD1234`)
5. **Extracts transaction ID** - Gets `txn_id` value from parsed form
6. **Gets user** - Retrieves user from session to get `tenantID`
7. **Calls service** - `paymentTransactionService.SubmitPaymentIntent(tenantID, txnID)`
8. **Returns HTTP 204** - Success response (No Content)

**Note:** The `requireTenant()` middleware already validated the user role, so the handler can trust the session is valid.

**⚠️  REDUNDANCY ISSUE:** The handler is doing **ALL** of this redundant work:

| Step | Handler Does | Middleware Already Did |
|------|--------------|----------------------|
| **2.2: Get Cookie** | `r.Cookie("sid")` (line 95) | ✅ `req.Cookie("sid")` (router.go:129) |
| **2.3: Validate Session** | `h.auth.ValidateSession()` (line 100) | ✅ `r.authHandler.ValidateSession()` (router.go:134) |
| **2.6: Get User from DB** | `h.users.GetByID()` (line 119) | ✅ `r.userRepo.GetByID()` (router.go:139) |

**All three steps are completely redundant!** The middleware already:
1. ✅ Got the cookie
2. ✅ Validated the session
3. ✅ Got the user from database
4. ✅ Validated the user role
5. ✅ Put the user in the request context

**Better Approach:** The handler should remove ALL redundant validation and use the user from context:

```go
func (h *TenantHandler) SubmitPayment(w http.ResponseWriter, r *http.Request) {
    // Step 2.1: Validate HTTP method (only thing handler needs to check)
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // ✅ REMOVE steps 2.2, 2.3, 2.6 - middleware already did this!
    // ❌ DELETE: r.Cookie() - middleware already got it
    // ❌ DELETE: ValidateSession() - middleware already validated it
    // ❌ DELETE: GetByID() - middleware already got user from DB
    
    // ✅ Use user from context (middleware already put it there)
    user, ok := http.GetUserFromContext(r.Context())
    if !ok || user == nil || user.TenantID == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Parse form and continue...
    if err := r.ParseForm(); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    
    txn := r.FormValue("txn_id")
    if txn == "" {
        http.Error(w, "txn_id required", http.StatusBadRequest)
        return
    }
    
    // Call service...
    if err := h.paymentTransactionService.SubmitPaymentIntent(*user.TenantID, txn); err != nil {
        http.Error(w, "failed", http.StatusBadRequest)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}
```

This would:
- ✅ Avoid duplicate database query (performance improvement)
- ✅ Remove redundant session validation
- ✅ Use data already fetched by middleware
- ✅ Follow the context pattern intended by the router

**Current Code Issue - Triple Redundancy:**

1. **Cookie Retrieval (Redundant):**
   - Middleware: `req.Cookie("sid")` (router.go:129)
   - Handler: `r.Cookie("sid")` (tenant_handler.go:95) ❌ **Duplicate!**

2. **Session Validation (Redundant):**
   - Middleware: `ValidateSession(cookie.Value)` (router.go:134)
   - Handler: `ValidateSession(c.Value)` (tenant_handler.go:100) ❌ **Duplicate!**

3. **Database Query (Redundant):**
   - Middleware: `r.userRepo.GetByID(session.UserID)` (router.go:139)
   - Handler: `h.users.GetByID(sess.UserID)` (tenant_handler.go:119) ❌ **Duplicate!**

**Performance Impact:**
- ⚠️ **3 redundant operations** per request
- ⚠️ **1 unnecessary database query** (doubles DB load)
- ⚠️ **2 unnecessary session validations** (doubles validation overhead)

**Solution:** Trust the middleware! Use `GetUserFromContext()` - it's already validated and in context.

---

### Step 3: Transaction Service Orchestrates Payment Creation
**File:** `internal/service/payment_transaction_service.go`  
**Function:** `SubmitPaymentIntent()`  
**Lines:** 30-61

```go
func (s *PaymentTransactionService) SubmitPaymentIntent(tenantID int, txnID string) error {
    // Step 3.1: Get or create current unpaid payment
    payment, err := s.paymentService.getOrCreateCurrentPayment(tenantID)
    if err != nil {
        return fmt.Errorf("get or create payment: %w", err)
    }

    // Step 3.2: Check if transaction already exists
    existing, err := s.paymentRepo.GetTransactionByPaymentAndID(payment.ID, txnID)
    if err != nil {
        return fmt.Errorf("check existing transaction: %w", err)
    }
    if existing != nil {
        return nil // Already exists, no error
    }

    // Step 3.3: Create payment transaction record
    tx := &domain.PaymentTransaction{
        PaymentID:     payment.ID,
        TransactionID: txnID,
        Amount:        nil, // NULL until owner verifies
        SubmittedAt:   time.Now(),
    }

    if err := s.paymentRepo.CreatePaymentTransaction(tx); err != nil {
        return fmt.Errorf("create payment transaction: %w", err)
    }

    return nil
}
```

**What happens:**
1. **Gets or creates payment** (see Step 4 below)
2. **Checks for duplicates** - queries if transaction ID already exists for this payment
3. **Creates transaction record** - inserts new record with `amount = NULL` (pending verification)

---

### Step 4: Get or Create Current Payment
**File:** `internal/service/payment_service.go`  
**Function:** `getOrCreateCurrentPayment()`  
**Lines:** 362-397

```go
func (s *PaymentService) getOrCreateCurrentPayment(tenantID int) (*domain.Payment, error) {
    // Step 4.1: Get unpaid payments
    unpaid, err := s.paymentRepo.GetUnpaidPaymentsByTenantID(tenantID)
    if err != nil {
        return nil, fmt.Errorf("get unpaid payments: %w", err)
    }

    // Step 4.2: If exists and not fully paid, return it
    if len(unpaid) > 0 {
        return unpaid[0], nil
    }

    // Step 4.3: Otherwise, create new payment
    tenant, err := s.tenantRepo.GetTenantByID(tenantID)
    unit, err := s.unitRepo.GetUnitByID(tenant.UnitID)
    
    // Calculate due date: Next PaymentDueDay >= today
    now := time.Now()
    dueDate := time.Date(now.Year(), now.Month(), unit.PaymentDueDay, 0, 0, 0, 0, time.UTC)
    if now.Day() > unit.PaymentDueDay {
        dueDate = dueDate.AddDate(0, 1, 0) // Next month
    }

    // Create new payment
    return s.CreatePaymentForTenant(tenantID, tenant.UnitID, dueDate, unit.MonthlyRent)
}
```

**What happens:**
1. **Query unpaid payments** - Checks if tenant has any unpaid payments
2. **Return existing** - If unpaid payment exists, return it
3. **Create new payment** - If no unpaid payment:
   - Fetches tenant from DB
   - Fetches unit from DB (to get `MonthlyRent` and `PaymentDueDay`)
   - Calculates next due date based on unit's `PaymentDueDay`
   - Creates new payment record in DB

---

### Step 5: Database Operations (Repository Layer)

#### 5.1: Query Unpaid Payments
**File:** `internal/repository/postgres/postgres_payment_repository.go`  
**Function:** `GetUnpaidPaymentsByTenantID()`  
**Lines:** 402-459

```sql
SELECT id, tenant_id, unit_id, amount, amount_paid, remaining_balance, 
       payment_date, due_date, is_paid, is_fully_paid, fully_paid_date, 
       payment_method, upi_id, notes, created_at
FROM payments
WHERE tenant_id = $1 AND is_fully_paid = FALSE
ORDER BY due_date ASC
```

**Database hit:** ✅ **SELECT** query on `payments` table

---

#### 5.2: Get Tenant (if creating new payment)
**File:** `internal/repository/postgres/postgres_tenant_repository.go`  
**Function:** `GetTenantByID()`

```sql
SELECT id, name, phone, aadhar_number, move_in_date, number_of_people, unit_id
FROM tenants
WHERE id = $1
```

**Database hit:** ✅ **SELECT** query on `tenants` table (only if creating new payment)

---

#### 5.3: Get Unit (if creating new payment)
**File:** `internal/repository/postgres/postgres_unit_repository.go`  
**Function:** `GetUnitByID()`

```sql
SELECT id, unit_code, unit_type, monthly_rent, payment_due_day, is_occupied
FROM units
WHERE id = $1
```

**Database hit:** ✅ **SELECT** query on `units` table (only if creating new payment)

---

#### 5.4: Create Payment (if needed)
**File:** `internal/repository/postgres/postgres_payment_repository.go`  
**Function:** `CreatePayment()`

```sql
INSERT INTO payments 
    (tenant_id, unit_id, amount, amount_paid, remaining_balance, 
     due_date, is_paid, is_fully_paid, payment_method, upi_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id
```

**Database hit:** ✅ **INSERT** query on `payments` table (only if no unpaid payment exists)

---

#### 5.5: Check Existing Transaction
**File:** `internal/repository/postgres/postgres_payment_repository_transactions.go`  
**Function:** `GetTransactionByPaymentAndID()`  
**Lines:** 96-140

```sql
SELECT id, payment_id, transaction_id, amount, submitted_at, 
       verified_at, verified_by_user_id, notes, created_at
FROM payment_transactions
WHERE payment_id = $1 AND transaction_id = $2
```

**Database hit:** ✅ **SELECT** query on `payment_transactions` table (prevents duplicates)

---

#### 5.6: Create Payment Transaction
**File:** `internal/repository/postgres/postgres_payment_repository_transactions.go`  
**Function:** `CreatePaymentTransaction()`  
**Lines:** 14-36

```sql
INSERT INTO payment_transactions 
    (payment_id, transaction_id, amount, submitted_at, verified_at, 
     verified_by_user_id, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at
```

**Database hit:** ✅ **INSERT** query on `payment_transactions` table

**Important:** `amount` is **NULL** at this point - owner will verify and set amount later.

---

## 🎯 Key Points: Submit Payment Flow

### 1. **Multiple Database Queries**
   - **Best case (unpaid payment exists):** 2 queries
     - SELECT unpaid payments
     - SELECT check existing transaction
     - INSERT payment transaction
   - **Worst case (no unpaid payment):** 6 queries
     - SELECT unpaid payments (empty)
     - SELECT tenant
     - SELECT unit
     - INSERT payment
     - SELECT check existing transaction
     - INSERT payment transaction

### 2. **Transaction Record Status**
   - `amount = NULL` (pending owner verification)
   - `verified_at = NULL` (not verified yet)
   - `verified_by_user_id = NULL` (no verifier yet)

### 3. **Duplicate Prevention**
   - Checks if transaction ID already exists for the payment
   - If exists, returns success (idempotent operation)

### 4. **Payment Creation Logic**
   - Only creates new payment if no unpaid payment exists
   - Calculates due date based on unit's `PaymentDueDay`
   - Uses unit's `MonthlyRent` for payment amount

---

## 🔍 Code Locations Summary: Submit Payment

| Layer | File | Function | Line | Database Operation |
|-------|------|----------|------|-------------------|
| **Template** | `templates/tenant-dashboard.html` | Form submit | 79-83, 153-178 | - |
| **Handler** | `internal/handlers/tenant_handler.go` | `SubmitPayment()` | 90-129 | - |
| **Service** | `internal/service/payment_transaction_service.go` | `SubmitPaymentIntent()` | 32-61 | - |
| **Service** | `internal/service/payment_service.go` | `getOrCreateCurrentPayment()` | 363-397 | - |
| **Repository** | `postgres_payment_repository.go` | `GetUnpaidPaymentsByTenantID()` | 402-459 | ✅ **SELECT** payments |
| **Repository** | `postgres_tenant_repository.go` | `GetTenantByID()` | - | ✅ **SELECT** tenant (if needed) |
| **Repository** | `postgres_unit_repository.go` | `GetUnitByID()` | - | ✅ **SELECT** unit (if needed) |
| **Repository** | `postgres_payment_repository.go` | `CreatePayment()` | - | ✅ **INSERT** payment (if needed) |
| **Repository** | `postgres_payment_repository_transactions.go` | `GetTransactionByPaymentAndID()` | 96-140 | ✅ **SELECT** check duplicate |
| **Repository** | `postgres_payment_repository_transactions.go` | `CreatePaymentTransaction()` | 14-36 | ✅ **INSERT** transaction |
| **Database** | PostgreSQL | Various tables | - | Stores transaction with `amount = NULL` |

---

## 🎬 Complete Example: Submit Payment

**User Action:** Tenant enters transaction ID "2349ABCD1234" and clicks "Submit Payment"

**Execution Flow:**
1. JavaScript: User submits form → `fetch('/api/payments/submit', ...)`
2. Router: Routes to `TenantHandler.SubmitPayment()`
3. Handler: Validates session, extracts `tenantID = 5`, `txnID = "2349ABCD1234"`
4. Handler: Calls `paymentTransactionService.SubmitPaymentIntent(5, "2349ABCD1234")`
5. Transaction Service: Calls `paymentService.getOrCreateCurrentPayment(5)`
6. Payment Service: 
   - Queries DB: `SELECT ... FROM payments WHERE tenant_id = 5 AND is_fully_paid = FALSE`
   - If no unpaid payment:
     - Queries DB: `SELECT ... FROM tenants WHERE id = 5`
     - Queries DB: `SELECT ... FROM units WHERE id = [tenant.unit_id]`
     - Inserts DB: `INSERT INTO payments ...` (creates new payment)
7. Transaction Service: Queries DB: `SELECT ... FROM payment_transactions WHERE payment_id = X AND transaction_id = '2349ABCD1234'` (check duplicate)
8. Transaction Service: Inserts DB: `INSERT INTO payment_transactions (payment_id, transaction_id, amount=NULL, submitted_at=NOW(), ...)` 
9. Handler: Returns HTTP 204 (No Content)
10. JavaScript: Shows success alert and reloads page

**Database Changes:**
- **New row created** in `payment_transactions` table:
  - `payment_id`: [payment ID]
  - `transaction_id`: "2349ABCD1234"
  - `amount`: NULL (pending verification)
  - `verified_at`: NULL (not verified yet)
  - `submitted_at`: [current timestamp]

---

## 📊 Visual Flow Diagram: Submit Payment

```
┌─────────────────────────────────────────────────────────────┐
│  Template Layer (Frontend)                                 │
│  tenant-dashboard.html                                     │
│  User enters: "2349ABCD1234"                              │
│  Clicks: "Submit Payment"                                  │
│  JavaScript: fetch('/api/payments/submit', {...})          │
│  Body: "txn_id=2349ABCD1234"                               │
│  Headers: Content-Type: application/x-www-form-urlencoded │
└────────────────────────────┬────────────────────────────────┘
                              │ HTTP POST Request
                              │ URL: /api/payments/submit
                              │ Cookie: sid=[session_token]
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  HTTP Server (Go net/http)                                 │
│  Receives HTTP request                                      │
│  Matches route pattern: "/api/payments/submit"              │
└────────────────────────────┬────────────────────────────────┘
                              │ Route match
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Router Layer                                               │
│  router.go: SetupRoutes()                                   │
│  Line 70: http.HandleFunc("/api/payments/submit", ...)      │
│                                                              │
│  requireTenant() Middleware:                                │
│  - Extracts "sid" cookie                                    │
│  - Validates session                                        │
│  - Checks user role = "tenant"                              │
│  - Adds user to request context                             │
└────────────────────────────┬────────────────────────────────┘
                              │ Authenticated request
                              │ (user validated)
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Handler Layer                                               │
│  tenant_handler.go                                          │
│  SubmitPayment()                                             │
│  - Validates HTTP method = POST                             │
│  - Gets session cookie                                      │
│  - Validates session                                        │
│  - Parses form: r.ParseForm()                               │
│  - Extracts: r.FormValue("txn_id")                          │
│  - Gets user: h.users.GetByID()                             │
│  - Calls: paymentTransactionService.SubmitPaymentIntent()    │
└────────────────────────────┬────────────────────────────────┘
                              │ Service call
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Transaction Service Layer                                   │
│  payment_transaction_service.go                             │
│  SubmitPaymentIntent()                                      │
│  1. Calls getOrCreateCurrentPayment()                        │
│  2. Checks for duplicate transaction                        │
│  3. Creates payment transaction                             │
└────────────────────────────┬────────────────────────────────┘
                              │ Service call
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Payment Service Layer                                      │
│  payment_service.go                                         │
│  getOrCreateCurrentPayment()                                │
│  - Queries unpaid payments                                  │
│  - OR creates new payment (if needed)                        │
└────────────────────────────┬────────────────────────────────┘
                              │ Repository calls
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Repository Layer                                           │
│  Multiple DB Operations:                                    │
│                                                              │
│  1. GetUnpaidPaymentsByTenantID()                           │
│     SELECT ... FROM payments                                │
│     WHERE tenant_id = $1 AND is_fully_paid = FALSE          │
│                                                              │
│  2. GetTenantByID() [if creating payment]                  │
│     SELECT ... FROM tenants WHERE id = $1                   │
│                                                              │
│  3. GetUnitByID() [if creating payment]                     │
│     SELECT ... FROM units WHERE id = $1                      │
│                                                              │
│  4. CreatePayment() [if creating payment]                   │
│     INSERT INTO payments (...) VALUES (...)                 │
│                                                              │
│  5. GetTransactionByPaymentAndID()                          │
│     SELECT ... FROM payment_transactions                    │
│     WHERE payment_id = $1 AND transaction_id = $2           │
│                                                              │
│  6. CreatePaymentTransaction()                               │
│     INSERT INTO payment_transactions                        │
│     (payment_id, transaction_id, amount=NULL, ...)          │
│     VALUES (...)                                            │
└────────────────────────────┬────────────────────────────────┘
                              │ SQL Queries
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  PostgreSQL Database                                         │
│                                                              │
│  Tables:                                                    │
│  - payments (SELECT/INSERT)                                 │
│  - tenants (SELECT - if needed)                            │
│  - units (SELECT - if needed)                                │
│  - payment_transactions (SELECT/INSERT)                      │
│                                                              │
│  Result: New row in payment_transactions with:              │
│  - transaction_id = "2349ABCD1234"                         │
│  - amount = NULL (pending verification)                     │
│  - verified_at = NULL                                        │
└─────────────────────────────────────────────────────────────┘
```

---

## ✅ Summary: Submit Payment Flow

**Q:** "When tenant clicks on submit payment by entering transaction ID, what happens?"

**A:** 

1. **Frontend:** User enters transaction ID, JavaScript sends POST to `/api/payments/submit`

2. **Handler:** Validates session, extracts `tenantID` and `txnID`, calls service

3. **Service Flow:**
   - Gets or creates current unpaid payment (may query/create payment in DB)
   - Checks if transaction already exists (queries `payment_transactions`)
   - Creates new transaction record (inserts into `payment_transactions`)

4. **Database Operations:**
   - **Always:** SELECT unpaid payments, SELECT check duplicate, INSERT transaction
   - **If no unpaid payment:** SELECT tenant, SELECT unit, INSERT payment

5. **Transaction Record:**
   - Created with `amount = NULL` (pending owner verification)
   - Owner must verify later to set amount and update payment

**Key Database Tables:**
- ✅ `payment_transactions` - INSERT new transaction record
- ✅ `payments` - SELECT (or INSERT if creating new payment)
- ✅ `tenants` - SELECT (only if creating payment)
- ✅ `units` - SELECT (only if creating payment)

---

# Call Flow: Owner Verifies Transaction

## 🔄 Complete Call Chain

This section traces the flow when an owner verifies a transaction submitted by a tenant.

---

## 📍 Flow Overview

```
unit-detail.html (Template)
    ↓
Owner views pending transactions
    ↓
Owner enters amount & clicks "Verify"
    ↓
JavaScript fetch() to /api/payments/mark-paid
    ↓
rental_handler.go (Handler Layer)
    ↓ MarkPaymentAsPaid() - Line 188
    ↓
payment_transaction_service.go (Service Layer)
    ↓ VerifyTransaction() - Line 65
    ↓
postgres_payment_repository_transactions.go (Repository Layer)
    ↓ VerifyTransaction() - Line 250
    ↓
PostgreSQL Database
    ↓ BEGIN TRANSACTION
    ↓ SELECT transaction (get payment_id)
    ↓ UPDATE payment_transactions (set amount, verified_at)
    ↓ UPDATE payments (update amount_paid, remaining_balance)
    ↓ COMMIT TRANSACTION
    ↓
payment_service.go (Service Layer)
    ↓ AutoCreateNextPayment() - Line 447 (if fully paid)
    ↓
PostgreSQL Database
    ↓ INSERT next payment (if needed)
```

---

## 📋 Detailed Step-by-Step Flow: Owner Verifies Transaction

### Step 1: Owner Views Pending Transactions
**File:** `templates/unit-detail.html`  
**Lines:** 398-424

```html
<!-- Pending Verifications -->
{{if .PendingVerifications}}
<div class="card">
    <h2>Pending Transaction Verifications</h2>
    {{range $idx, $txn := .PendingVerifications}}
    <div class="payment-item">
        <h4>Transaction ID: {{$txn.TransactionID}}</h4>
        <p>Submitted: {{$txn.SubmittedAt.Format "Jan 2, 2006 3:04 PM"}}</p>
        <p>⚠️ Waiting for owner verification</p>
        
        <input type="number" id="verify_amount_{{$txn.ID}}" placeholder="Enter amount" />
        <button onclick="verifyTransaction('{{$txn.TransactionID}}', 'verify_amount_{{$txn.ID}}')">
            Verify
        </button>
    </div>
    {{end}}
</div>
{{end}}
```

**What happens:**
- Template displays all pending transactions (transactions with `verified_at = NULL`)
- Shows transaction ID and submission time
- Provides input field for owner to enter amount
- Shows "Verify" button for each transaction

---

### Step 2: Owner Enters Amount and Clicks Verify
**File:** `templates/unit-detail.html`  
**Lines:** 546-579

```javascript
function verifyTransaction(transactionID, amountInputId) {
    const amountInput = document.getElementById(amountInputId);
    const amount = parseInt(amountInput.value);
    
    if (!amount || amount <= 0) {
        alert('Please enter a valid amount greater than 0');
        return;
    }

    const verifyData = {
        transaction_id: transactionID,  // e.g., "2349ABCD1234"
        amount: amount                  // e.g., 5000
    };

    fetch('/api/payments/mark-paid', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(verifyData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            alert('Transaction verified successfully! Payment updated.');
            location.reload();
        }
    });
}
```

**What happens:**
1. JavaScript gets amount from input field
2. Validates amount > 0
3. Sends HTTP POST request to `/api/payments/mark-paid`
4. Request body: `{"transaction_id": "2349ABCD1234", "amount": 5000}`

---

### Step 3: Handler Validates and Processes Request
**File:** `internal/handlers/rental_handler.go`  
**Function:** `MarkPaymentAsPaid()`  
**Lines:** 188-281

```go
func (h *RentalHandler) MarkPaymentAsPaid(w http.ResponseWriter, r *http.Request) {
    // Step 3.1: Validate HTTP method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Step 3.2: Parse request body
    var req struct {
        PaymentID     int    `json:"payment_id"`     // Legacy
        TransactionID string `json:"transaction_id"` // New flow
        Amount        int    `json:"amount"`         // Verification amount
        PaymentDate   string `json:"payment_date"`    // Legacy
        Notes         string `json:"notes"`          // Legacy
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Step 3.3: Validate session (owner only)
    cookie, err := r.Cookie(h.cookieName)
    sess, err := h.authService.ValidateSession(cookie.Value)
    user, err := h.userRepo.GetByID(sess.UserID)

    // Step 3.4: NEW - Transaction verification flow
    if req.TransactionID != "" {
        if req.Amount <= 0 {
            http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
            return
        }

        // Step 3.5: Call service to verify transaction
        if err := h.paymentTransactionService.VerifyTransaction(
            req.TransactionID, 
            req.Amount, 
            user.ID,  // Owner who verified
        ); err != nil {
            // Return error response
            return
        }

        // Step 3.6: Return success
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": true,
            "message": "Transaction verified and payment updated",
        })
        return
    }

    // LEGACY: Old flow (if payment_id provided instead)
    // ...
}
```

**What happens:**
1. **Validates HTTP method** - Must be POST
2. **Parses JSON request** - Extracts `transaction_id` and `amount`
3. **Validates session** - Gets authenticated owner user
4. **Checks for new flow** - If `transaction_id` provided, uses new verification flow
5. **Validates amount** - Must be > 0
6. **Calls service** - `paymentTransactionService.VerifyTransaction(transactionID, amount, ownerID)`
7. **Returns success** - JSON response with success message

---

### Step 4: Transaction Service Orchestrates Verification
**File:** `internal/service/payment_transaction_service.go`  
**Function:** `VerifyTransaction()`  
**Lines:** 63-94

```go
func (s *PaymentTransactionService) VerifyTransaction(
    transactionID string, 
    amount int, 
    verifiedByUserID int,
) error {
    // Step 4.1: Get transaction by transaction_id (UPI transaction ID string)
    // transactionID = "2349ABCD1234" (the UPI transaction ID, NOT the database primary key)
    tx, err := s.paymentRepo.GetTransactionByID(transactionID)
    if err != nil {
        return fmt.Errorf("get transaction: %w", err)
    }
    if tx == nil {
        return fmt.Errorf("transaction not found")
    }

    paymentID := tx.PaymentID  // Get payment ID from transaction record

    // Step 4.2: Verify transaction (updates transaction + payment atomically)
    if err := s.paymentRepo.VerifyTransaction(transactionID, amount, verifiedByUserID); err != nil {
        return fmt.Errorf("verify transaction: %w", err)
    }

    // Step 4.3: Get updated payment to check if fully paid
    payment, err := s.paymentRepo.GetPaymentByID(paymentID)
    if err != nil {
        return fmt.Errorf("get updated payment: %w", err)
    }

    // Step 4.4: Auto-create next payment if fully paid
    if payment.IsFullyPaid {
        return s.paymentService.AutoCreateNextPayment(payment)
    }

    return nil
}
```

**What happens:**
1. **Gets transaction** - Queries `payment_transactions` table by `transaction_id` (the UPI transaction ID string like "2349ABCD1234") to get `paymentID`
2. **Verifies transaction** - Calls repository to update transaction and payment (atomic)
3. **Checks payment status** - Gets updated payment to see if fully paid
4. **Auto-creates next payment** - If payment is fully paid, creates next month's payment

**Important:** `transactionID` is the **UPI transaction ID string** (e.g., "2349ABCD1234"), NOT the database primary key. The query searches by `transaction_id` column, not the `id` column.

---

### Step 5: Repository Performs Atomic Database Operations
**File:** `internal/repository/postgres/postgres_payment_repository_transactions.go`  
**Function:** `VerifyTransaction()`  
**Lines:** 248-311

```go
func (r *PostgresPaymentRepository) VerifyTransaction(
    transactionID string, 
    amount int, 
    verifiedByUserID int,
) error {
    // Step 5.1: Begin database transaction (atomicity)
    dbTx, err := r.db.Begin()
    defer dbTx.Rollback()  // Rollback on error

    // Step 5.2: Get transaction to find payment_id
    var paymentID int
    var currentAmount sql.NullInt64
    err = dbTx.QueryRow(`
        SELECT payment_id, amount 
        FROM payment_transactions 
        WHERE transaction_id = $1`,
        transactionID,
    ).Scan(&paymentID, &currentAmount)

    if err != nil {
        return fmt.Errorf("transaction not found: %w", err)
    }

    // Step 5.3: Check if already verified
    if currentAmount.Valid {
        return fmt.Errorf("transaction already verified")
    }

    // Step 5.4: Update transaction record
    now := time.Now()
    _, err = dbTx.Exec(`
        UPDATE payment_transactions 
        SET amount = $1, verified_at = $2, verified_by_user_id = $3
        WHERE transaction_id = $4`,
        amount, now, verifiedByUserID, transactionID,
    )

    // Step 5.5: Update payment (add amount to amount_paid, recalculate balance)
    _, err = dbTx.Exec(`
        UPDATE payments 
        SET amount_paid = amount_paid + $1,
            remaining_balance = amount - (amount_paid + $1),
            is_fully_paid = (amount - (amount_paid + $1) <= 0),
            fully_paid_date = CASE 
                WHEN (amount - (amount_paid + $1) <= 0) AND fully_paid_date IS NULL THEN $2
                ELSE fully_paid_date
            END
        WHERE id = $3`,
        amount, now, paymentID,
    )

    // Step 5.6: Commit transaction (all or nothing)
    if err = dbTx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

**What happens:**
1. **BEGIN TRANSACTION** - Starts database transaction for atomicity
2. **SELECT transaction** - Gets `payment_id` and checks if already verified
3. **UPDATE payment_transactions** - Sets `amount`, `verified_at`, `verified_by_user_id`
4. **UPDATE payments** - Adds `amount` to `amount_paid`, recalculates `remaining_balance`, updates `is_fully_paid`
5. **COMMIT** - Commits all changes atomically (or rolls back on error)

**Database Operations:**

**Step 4.1: Get Transaction by transaction_id**
```sql
-- Table: payment_transactions
-- Column: transaction_id (VARCHAR) - This is the UPI transaction ID, NOT the primary key
-- Example: transaction_id = '2349ABCD1234'

SELECT id, payment_id, transaction_id, amount, submitted_at, 
       verified_at, verified_by_user_id, notes, created_at
FROM payment_transactions 
WHERE transaction_id = '2349ABCD1234';
```
**Returns:** Transaction record with `payment_id` needed for next step

**Step 5.2-5.5: Verify Transaction (Atomic)**
```sql
-- Operation 1: Get transaction (inside DB transaction)
SELECT payment_id, amount 
FROM payment_transactions 
WHERE transaction_id = '2349ABCD1234';

-- Operation 2: Update transaction record
UPDATE payment_transactions 
SET amount = 5000, verified_at = NOW(), verified_by_user_id = 1
WHERE transaction_id = '2349ABCD1234';

-- Operation 3: Update payment record
UPDATE payments 
SET amount_paid = amount_paid + 5000,
    remaining_balance = amount - (amount_paid + 5000),
    is_fully_paid = (amount - (amount_paid + 5000) <= 0),
    fully_paid_date = CASE 
        WHEN (amount - (amount_paid + 5000) <= 0) THEN NOW()
        ELSE fully_paid_date
    END
WHERE id = 123;  -- payment_id from transaction record
```

**Table Structure:**
```sql
CREATE TABLE payment_transactions (
    id SERIAL PRIMARY KEY,                    -- Auto-increment integer (1, 2, 3...)
    payment_id INTEGER NOT NULL,              -- Foreign key to payments table
    transaction_id VARCHAR(255) NOT NULL,      -- UPI transaction ID (e.g., "2349ABCD1234")
    amount INTEGER,                           -- NULL until owner verifies
    submitted_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,                    -- NULL until owner verifies
    verified_by_user_id INTEGER,              -- NULL until owner verifies
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Key Points:**
- **Table:** `payment_transactions`
- **Query Column:** `transaction_id` (VARCHAR) - the UPI transaction ID string (e.g., "2349ABCD1234")
- **NOT the primary key:** We query by `transaction_id` (string), NOT by `id` (integer primary key)
- **Example:** `WHERE transaction_id = '2349ABCD1234'` (string match)
- **Purpose:** Find which payment this transaction belongs to (`payment_id`)

**Why use `transaction_id` instead of `id`?**
- Owner receives the UPI transaction ID from tenant (e.g., "2349ABCD1234")
- Owner doesn't know the database `id` (could be 42, 100, etc.)
- So we search by the business identifier (`transaction_id`), not the technical ID (`id`)

**Important:** All operations are in a single database transaction - either all succeed or all fail.

---

### Step 6: Auto-Create Next Payment (If Fully Paid)
**File:** `internal/service/payment_service.go`  
**Function:** `AutoCreateNextPayment()`  
**Lines:** 445-462

```go
func (s *PaymentService) AutoCreateNextPayment(payment *domain.Payment) error {
    if !payment.IsFullyPaid {
        return nil // Not fully paid, no need to create next
    }

    // Check if next payment already exists
    nextDueDate := payment.DueDate.AddDate(0, 1, 0)  // Next month
    existing, err := s.paymentRepo.GetPaymentByTenantAndMonth(
        payment.TenantID, 
        nextDueDate.Month(), 
        nextDueDate.Year(),
    )
    if err == nil && existing != nil {
        return nil // Next payment already exists
    }

    // Create next payment
    _, err = s.CreateNextPayment(payment)
    return err
}
```

**What happens:**
1. **Checks if fully paid** - Only creates next if payment is fully paid
2. **Checks if next exists** - Prevents duplicate payments
3. **Creates next payment** - Creates payment for next month with same amount

---

## 🎯 Key Points: Owner Verification Flow

### 1. **Atomic Database Transaction**
   - All database operations happen in one transaction
   - Either all succeed (commit) or all fail (rollback)
   - Ensures data consistency

### 2. **Payment Update Logic**
   - Adds verified amount to `amount_paid`
   - Recalculates `remaining_balance = amount - amount_paid`
   - Sets `is_fully_paid = true` if `remaining_balance <= 0`
   - Sets `fully_paid_date` when payment becomes fully paid

### 3. **Transaction Record Update**
   - Sets `amount` (was NULL, now has value)
   - Sets `verified_at` (timestamp of verification)
   - Sets `verified_by_user_id` (owner who verified)

### 4. **Automatic Next Payment**
   - If payment becomes fully paid, automatically creates next month's payment
   - Prevents duplicate payment creation

### 5. **Partial Payment Support**
   - Owner can verify any amount (not just full payment)
   - Multiple transactions can partially pay a single payment
   - Payment is marked fully paid when `remaining_balance <= 0`

---

## 🔍 Code Locations Summary: Owner Verification

| Layer | File | Function | Line | Database Operation |
|-------|------|----------|------|-------------------|
| **Template** | `templates/unit-detail.html` | `verifyTransaction()` | 546-579 | - |
| **Handler** | `internal/handlers/rental_handler.go` | `MarkPaymentAsPaid()` | 188-281 | - |
| **Service** | `internal/service/payment_transaction_service.go` | `VerifyTransaction()` | 65-94 | - |
| **Service** | `internal/service/payment_service.go` | `AutoCreateNextPayment()` | 447-462 | - |
| **Repository** | `postgres_payment_repository_transactions.go` | `GetTransactionByID()` | 142-186 | ✅ **SELECT** transaction |
| **Repository** | `postgres_payment_repository_transactions.go` | `VerifyTransaction()` | 248-311 | ✅ **BEGIN**, **SELECT**, **UPDATE** transaction, **UPDATE** payment, **COMMIT** |
| **Repository** | `postgres_payment_repository.go` | `GetPaymentByID()` | - | ✅ **SELECT** payment (if needed) |
| **Repository** | `postgres_payment_repository.go` | `CreateNextPayment()` | - | ✅ **INSERT** payment (if fully paid) |
| **Database** | PostgreSQL | Transaction + Updates | - | Atomic transaction updates |

---

## 🎬 Complete Example: Owner Verifies Transaction

**User Action:** Owner views unit details, sees pending transaction "2349ABCD1234", enters amount ₹5000, clicks "Verify"

**Execution Flow:**
1. **Template:** Displays pending transactions with verify buttons
2. **JavaScript:** Owner enters amount 5000, clicks "Verify"
3. **HTTP Request:** POST `/api/payments/mark-paid` with `{"transaction_id": "2349ABCD1234", "amount": 5000}`
4. **Router:** Routes to `RentalHandler.MarkPaymentAsPaid()` (requires owner role)
5. **Handler:** Validates session, extracts `transactionID` and `amount`, calls service
6. **Transaction Service:** 
   - Queries DB: `SELECT payment_id, amount FROM payment_transactions WHERE transaction_id = '2349ABCD1234'`
   - Gets `paymentID = 123`
7. **Repository:** Executes atomic transaction:
   - `UPDATE payment_transactions SET amount = 5000, verified_at = NOW(), verified_by_user_id = 1 WHERE transaction_id = '2349ABCD1234'`
   - `UPDATE payments SET amount_paid = amount_paid + 5000, remaining_balance = amount - (amount_paid + 5000), is_fully_paid = ... WHERE id = 123`
   - `COMMIT` transaction
8. **Transaction Service:** Gets updated payment, checks if `IsFullyPaid = true`
9. **Payment Service:** If fully paid, creates next month's payment
10. **Handler:** Returns success JSON response
11. **JavaScript:** Shows success alert, reloads page

**Database Changes:**
- **payment_transactions table:**
  - `amount`: NULL → 5000
  - `verified_at`: NULL → [current timestamp]
  - `verified_by_user_id`: NULL → [owner user ID]
  
- **payments table:**
  - `amount_paid`: [old value] → [old value + 5000]
  - `remaining_balance`: [old value] → [amount - new amount_paid]
  - `is_fully_paid`: false → true (if balance <= 0)
  - `fully_paid_date`: NULL → [current timestamp] (if fully paid)

---

## 📊 Visual Flow Diagram: Owner Verification

```
┌─────────────────────────────────────────────────────────────┐
│  Template Layer (Frontend)                                 │
│  unit-detail.html                                          │
│  Owner views: Transaction "2349ABCD1234"                   │
│  Owner enters: Amount = 5000                               │
│  Owner clicks: "Verify"                                    │
│  JavaScript: fetch('/api/payments/mark-paid', {...})       │
└────────────────────────────┬────────────────────────────────┘
                              │ HTTP POST
                              │ Body: {"transaction_id": "...", "amount": 5000}
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Handler Layer                                               │
│  rental_handler.go                                          │
│  MarkPaymentAsPaid()                                         │
│  - Validates session (owner)                                │
│  - Extracts transaction_id & amount                          │
│  - Calls paymentTransactionService.VerifyTransaction()       │
└────────────────────────────┬────────────────────────────────┘
                              │ Service call
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Transaction Service Layer                                   │
│  payment_transaction_service.go                             │
│  VerifyTransaction()                                        │
│  1. Gets transaction by ID                                   │
│  2. Calls repository to verify (atomic)                       │
│  3. Checks if payment fully paid                             │
│  4. Auto-creates next payment (if needed)                    │
└────────────────────────────┬────────────────────────────────┘
                              │ Repository calls
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Repository Layer                                           │
│  postgres_payment_repository_transactions.go                │
│  VerifyTransaction()                                        │
│                                                              │
│  BEGIN TRANSACTION                                          │
│  1. SELECT transaction (get payment_id)                      │
│  2. UPDATE payment_transactions                              │
│     SET amount = 5000, verified_at = NOW(), ...             │
│  3. UPDATE payments                                          │
│     SET amount_paid = amount_paid + 5000,                    │
│         remaining_balance = amount - (amount_paid + 5000),  │
│         is_fully_paid = ...                                  │
│  COMMIT TRANSACTION                                         │
└────────────────────────────┬────────────────────────────────┘
                              │ SQL Queries (Atomic)
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  PostgreSQL Database                                         │
│                                                              │
│  payment_transactions:                                      │
│  - amount: NULL → 5000                                       │
│  - verified_at: NULL → [timestamp]                          │
│  - verified_by_user_id: NULL → [owner ID]                   │
│                                                              │
│  payments:                                                  │
│  - amount_paid: [old] → [old + 5000]                        │
│  - remaining_balance: [old] → [recalculated]                │
│  - is_fully_paid: false → true (if balance <= 0)            │
│  - fully_paid_date: NULL → [timestamp] (if fully paid)      │
└─────────────────────────────────────────────────────────────┘
```

---

## ✅ Summary: Owner Verification Flow

**Q:** "When owner verifies a transaction, what happens?"

**A:** 

1. **Frontend:** Owner enters amount, JavaScript sends POST to `/api/payments/mark-paid`

2. **Handler:** Validates owner session, extracts `transactionID` and `amount`, calls service

3. **Service Flow:**
   - Gets transaction by ID to find `paymentID`
   - Calls repository to verify transaction atomically
   - Checks if payment is fully paid
   - Auto-creates next payment if fully paid

4. **Database Operations (Atomic):**
   - **SELECT** transaction (get payment_id, check if already verified)
   - **UPDATE** payment_transactions (set amount, verified_at, verified_by_user_id)
   - **UPDATE** payments (add amount to amount_paid, recalculate balance, update is_fully_paid)
   - **COMMIT** (all changes succeed together or all fail)

5. **Result:**
   - Transaction marked as verified with amount
   - Payment's `amount_paid` increased
   - Payment's `remaining_balance` decreased
   - Payment marked `is_fully_paid = true` if balance reaches 0
   - Next month's payment created automatically if fully paid

**Key Database Tables:**
- ✅ `payment_transactions` - UPDATE (set amount, verified_at, verified_by_user_id)
- ✅ `payments` - UPDATE (update amount_paid, remaining_balance, is_fully_paid)
- ✅ `payments` - INSERT (create next payment if fully paid)

---

## 🔗 How JavaScript fetch() Connects to Handler

### Connection Flow:

```
JavaScript fetch() → HTTP Request → Router → Middleware → Handler
```

### Detailed Connection:

1. **JavaScript:**
   ```javascript
   fetch('/api/payments/submit', {
       method: 'POST',
       headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
       body: data.toString()  // "txn_id=2349ABCD1234"
   })
   ```
   - Creates HTTP POST request
   - URL: `/api/payments/submit`
   - Body: URL-encoded form data
   - Includes cookie: `sid=[session_token]` (automatically sent by browser)

2. **HTTP Server (Go):**
   - Go's `net/http` package receives the request
   - Matches URL path to registered route handlers

3. **Router (router.go:70):**
   ```go
   http.HandleFunc("/api/payments/submit", r.requireTenant(r.tenantHandler.SubmitPayment))
   ```
   - Registers route `/api/payments/submit`
   - Wraps handler with `requireTenant()` middleware
   - Maps to `tenantHandler.SubmitPayment` function

4. **Middleware (requireTenant):**
   ```go
   func (r *Router) requireTenant(next http.HandlerFunc) http.HandlerFunc {
       return func(w http.ResponseWriter, req *http.Request) {
           user, err := r.loadSessionAndValidateRole(req, "tenant")
           // ... validates session and role ...
           next(w, req.WithContext(ctx))  // Calls next handler
       }
   }
   ```
   - Validates session cookie
   - Checks user role is "tenant"
   - If valid, calls `next()` which is `tenantHandler.SubmitPayment`

5. **Handler (SubmitPayment):**
   ```go
   func (h *TenantHandler) SubmitPayment(w http.ResponseWriter, r *http.Request) {
       // Receives authenticated request
       r.ParseForm()  // Parses "txn_id=2349ABCD1234"
       txn := r.FormValue("txn_id")  // Gets "2349ABCD1234"
       // ... processes request ...
   }
   ```
   - Receives the HTTP request with authenticated user context
   - Parses form data from request body
   - Extracts `txn_id` value
   - Processes the payment submission

### Request/Response Cycle:

```
┌─────────────┐
│  Browser    │
│  JavaScript │
└──────┬──────┘
       │ fetch('/api/payments/submit', {...})
       │ POST /api/payments/submit
       │ Cookie: sid=abc123
       │ Body: txn_id=2349ABCD1234
       ↓
┌──────────────────┐
│  HTTP Server     │
│  (Go net/http)   │
└──────┬───────────┘
       │ Route match: "/api/payments/submit"
       ↓
┌──────────────────┐
│  Router           │
│  SetupRoutes()    │
│  requireTenant()  │ ← Validates session & role
└──────┬───────────┘
       │ Authenticated request
       ↓
┌──────────────────┐
│  Handler         │
│  SubmitPayment() │ ← Processes request
└──────┬───────────┘
       │ HTTP 204 No Content
       │ (Success response)
       ↓
┌─────────────┐
│  Browser    │
│  JavaScript │
│  .then()    │ ← Receives response
└─────────────┘
```

### Key Points:

1. **URL Matching:** Router matches `/api/payments/submit` path exactly
2. **Middleware Chain:** `requireTenant()` → `SubmitPayment()` (executed in order)
3. **Form Data:** JavaScript sends URL-encoded form, Go's `r.ParseForm()` parses it
4. **Session:** Cookie (`sid`) sent automatically by browser, middleware validates it
5. **Response:** Handler returns HTTP 204, JavaScript `.then()` handles it


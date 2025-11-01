# Migration Rollback Guide

## ⚠️ IMPORTANT WARNINGS

**Before Rolling Back:**
1. **Data Loss**: Rolling back will permanently delete:
   - All `payment_transactions` table data
   - All `amount_paid`, `remaining_balance`, `is_fully_paid`, `fully_paid_date` values

2. **Transaction IDs**: Transaction IDs will be migrated back to `notes` field, but:
   - Amount information per transaction will be lost
   - Verification timestamps will be lost
   - Only transaction IDs will be preserved in notes field

3. **Irreversible**: This rollback cannot be fully undone without re-running the forward migration

---

## 📋 Rollback Steps

### Option A: Full Rollback (Recommended)

**File:** `migrations/001_add_partial_payment_support_ROLLBACK.sql`

**What it does:**
1. Updates `is_paid` field based on `is_fully_paid`
2. Migrates transaction IDs back to `notes` field
3. Drops `payment_transactions` table
4. Drops new columns from `payments` table

**Execution:**
```sql
-- Copy and paste the entire rollback file into Neon console
-- Or run via psql:
psql $DATABASE_URL -f migrations/001_add_partial_payment_support_ROLLBACK.sql
```

---

### Option B: Partial Rollback (Keep Some Data)

**If you want to preserve transaction data:**

**Step 1: Backup transaction data**
```sql
BEGIN;

-- Backup transaction table
CREATE TABLE payment_transactions_backup AS 
SELECT * FROM payment_transactions;

-- Backup payment data
CREATE TABLE payments_backup AS 
SELECT * FROM payments;

COMMIT;
```

**Step 2: Review backup**
```sql
SELECT COUNT(*) FROM payment_transactions_backup;
SELECT COUNT(*) FROM payments_backup;
```

**Step 3: Run rollback**
```sql
-- Run the full rollback migration
```

**Step 4: Restore if needed** (if rollback was a mistake)
```sql
-- Restore backup tables if needed
-- (You'd need to restore manually)
```

---

## 🔄 Rollback Process

### Phase 1: Data Migration (Preserve What We Can)

**What Gets Preserved:**
- ✅ `is_paid` status (restored from `is_fully_paid`)
- ✅ `payment_date` (restored from `fully_paid_date`)
- ✅ Transaction IDs (migrated to `notes` field)

**What Gets Lost:**
- ❌ Individual transaction amounts
- ❌ Verification timestamps
- ❌ Who verified each transaction
- ❌ Amount_paid, remaining_balance values

### Phase 2: Schema Rollback

1. Drop indexes
2. Drop `payment_transactions` table
3. Drop new columns from `payments` table

### Phase 3: Verification

Run verification queries to ensure rollback succeeded.

---

## ✅ Rollback Verification

### Check 1: Payments Table Structure
```sql
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'payments' 
ORDER BY ordinal_position;
```

**Expected:** Should NOT include:
- `amount_paid`
- `remaining_balance`
- `is_fully_paid`
- `fully_paid_date`

**Should include:**
- `id`, `tenant_id`, `unit_id`, `amount`, `payment_date`, `due_date`, `is_paid`, `payment_method`, `upi_id`, `notes`, `created_at`

### Check 2: Payment Transactions Table (Should Not Exist)
```sql
SELECT COUNT(*) FROM payment_transactions;
```

**Expected:** Error: `relation "payment_transactions" does not exist`

### Check 3: is_paid Field Restored
```sql
SELECT 
    id, 
    amount, 
    is_paid, 
    payment_date, 
    notes 
FROM payments 
LIMIT 10;
```

**Expected:** `is_paid` should reflect old payment status

### Check 4: Transaction IDs in Notes
```sql
SELECT id, notes 
FROM payments 
WHERE notes LIKE 'TXN:%' 
LIMIT 5;
```

**Expected:** Transaction IDs should appear in notes field

---

## 🚨 Emergency Rollback

**If something goes wrong during forward migration:**

### Immediate Steps:

1. **Stop all application traffic** (if possible)

2. **Check migration status:**
```sql
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'payments' 
AND column_name IN ('amount_paid', 'remaining_balance', 'is_fully_paid', 'fully_paid_date');
```

3. **If columns exist but migration failed:**
   - Run rollback to clean up
   - Fix issues
   - Re-run forward migration

4. **If rollback fails:**
   - Manually drop columns/table
   - Restore from database backup (if available)

---

## 📝 Rollback Scenarios

### Scenario 1: Migration Partially Applied

**Symptoms:**
- Some columns exist, some don't
- Payment_transactions table exists but empty

**Solution:**
- Run full rollback
- Fix any issues
- Re-run forward migration

### Scenario 2: Data Corruption

**Symptoms:**
- Data looks wrong after migration
- Calculations incorrect

**Solution:**
1. Backup current state
2. Run rollback
3. Fix application code
4. Re-run forward migration

### Scenario 3: Need to Test Rollback

**For Testing:**
1. Create test database
2. Run forward migration
3. Run rollback migration
4. Verify everything reverted

---

## 🔐 Safety Checklist

Before rolling back:
- [ ] Backup database (if possible)
- [ ] Document current state
- [ ] Test rollback in staging/test environment first
- [ ] Stop application (if possible)
- [ ] Verify rollback script syntax
- [ ] Plan for data loss

After rolling back:
- [ ] Verify schema reverted correctly
- [ ] Check data integrity
- [ ] Test application functionality
- [ ] Update application code (revert to old logic)

---

## 💡 Best Practices

1. **Always test rollback in non-production first**
2. **Keep database backups** before migrations
3. **Use transactions** (BEGIN/COMMIT) - rollback script uses this
4. **Verify incrementally** - check after each step
5. **Document any manual fixes** needed

---

## 📞 Troubleshooting

### Error: "column does not exist"
- This is OK - column was already dropped or never existed
- Continue with rollback

### Error: "table does not exist"
- This is OK - table was already dropped
- Continue with rollback

### Error: "foreign key constraint"
- Drop foreign key constraints first:
```sql
ALTER TABLE payment_transactions DROP CONSTRAINT IF EXISTS payment_transactions_payment_id_fkey;
```

### Error: "cannot drop column because it is used"
- Check for views or functions using the column
- Drop them first, then drop column

---

**Remember: Rollback is a safety net, but prevention is better than cure!** 🛡️


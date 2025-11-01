-- Quick Verification - Run these after migration
-- Copy and paste into Neon console

-- 1. Check if backfill worked correctly for paid payments
SELECT 
    id, 
    amount, 
    amount_paid, 
    remaining_balance, 
    is_paid, 
    is_fully_paid,
    payment_date,
    fully_paid_date
FROM payments;

-- Expected Results:
-- If payment is_paid = TRUE:
--   - amount_paid should equal amount
--   - remaining_balance should be 0
--   - is_fully_paid should be TRUE
--   - fully_paid_date should equal payment_date
--
-- If payment is_paid = FALSE:
--   - amount_paid should be 0
--   - remaining_balance should equal amount
--   - is_fully_paid should be FALSE
--   - fully_paid_date should be NULL

-- 2. Verify data integrity (balance calculation)
SELECT 
    id,
    amount,
    amount_paid,
    remaining_balance,
    (amount - amount_paid) as calculated_balance,
    CASE 
        WHEN remaining_balance = (amount - amount_paid) THEN '✅ CORRECT'
        ELSE '❌ MISMATCH - RUN FIX BELOW'
    END as status
FROM payments;

-- 3. If you see any ❌ MISMATCH, run this fix:
-- UPDATE payments 
-- SET remaining_balance = amount - amount_paid
-- WHERE remaining_balance != (amount - amount_paid);

-- 4. Count by status
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN is_fully_paid THEN 1 ELSE 0 END) as fully_paid,
    SUM(CASE WHEN amount_paid > 0 AND NOT is_fully_paid THEN 1 ELSE 0 END) as partially_paid,
    SUM(CASE WHEN amount_paid = 0 THEN 1 ELSE 0 END) as unpaid
FROM payments;


-- ROLLBACK Migration: Revert Partial Payment Support
-- Description: Reverts all changes from 001_add_partial_payment_support.sql
-- WARNING: This will DELETE all payment_transactions data!
-- WARNING: This will DELETE amount_paid, remaining_balance, is_fully_paid, fully_paid_date data!
-- Date: 2024
-- Run this ONLY if you need to revert to the old schema

BEGIN;

-- ============================================
-- STEP 1: Backup transaction data (optional)
-- ============================================
-- If you want to preserve transaction data before rolling back,
-- create a backup table first:
-- CREATE TABLE payment_transactions_backup AS SELECT * FROM payment_transactions;
-- Then you can restore it later if needed

-- ============================================
-- STEP 2: Update payments to restore old is_paid logic
-- ============================================
-- Restore is_paid field based on is_fully_paid
UPDATE payments 
SET 
    is_paid = CASE 
        WHEN is_fully_paid = TRUE THEN TRUE 
        ELSE FALSE 
    END,
    payment_date = CASE 
        WHEN is_fully_paid = TRUE THEN fully_paid_date 
        ELSE NULL 
    END
WHERE is_fully_paid IS NOT NULL;

-- For payments where is_fully_paid is NULL (old data), keep current is_paid value
-- This handles edge cases gracefully

-- ============================================
-- STEP 3: Migrate transaction IDs back to notes field
-- ============================================
-- Extract transaction IDs from payment_transactions table back to notes
-- This attempts to preserve transaction ID information
UPDATE payments p
SET notes = COALESCE(
    (SELECT string_agg('TXN:' || pt.transaction_id, '; ' ORDER BY pt.submitted_at)
     FROM payment_transactions pt
     WHERE pt.payment_id = p.id),
    p.notes
)
WHERE EXISTS (
    SELECT 1 FROM payment_transactions pt WHERE pt.payment_id = p.id
);

-- Note: If notes already had data, it will be merged with transaction IDs
-- If you want to preserve original notes, create a backup first:
-- ALTER TABLE payments ADD COLUMN notes_backup TEXT;
-- UPDATE payments SET notes_backup = notes;

-- ============================================
-- STEP 4: Drop indexes (if they exist)
-- ============================================
DROP INDEX IF EXISTS idx_payment_transactions_verified_at;
DROP INDEX IF EXISTS idx_payment_transactions_transaction_id;
DROP INDEX IF EXISTS idx_payment_transactions_payment_id;
DROP INDEX IF EXISTS idx_payments_is_fully_paid;
DROP INDEX IF EXISTS idx_payments_tenant_due_date;

-- ============================================
-- STEP 5: Drop payment_transactions table
-- ============================================
-- WARNING: This permanently deletes all transaction data!
DROP TABLE IF EXISTS payment_transactions CASCADE;

-- ============================================
-- STEP 6: Drop new columns from payments table
-- ============================================
-- WARNING: This permanently deletes amount_paid, remaining_balance, etc.
ALTER TABLE payments 
DROP COLUMN IF EXISTS fully_paid_date,
DROP COLUMN IF EXISTS is_fully_paid,
DROP COLUMN IF EXISTS remaining_balance,
DROP COLUMN IF EXISTS amount_paid;

-- Note: PostgreSQL doesn't support "IF EXISTS" for columns in older versions
-- If you get an error, the column doesn't exist - that's fine, continue

-- ============================================
-- VERIFICATION
-- ============================================
-- Verify rollback:
-- SELECT column_name FROM information_schema.columns WHERE table_name = 'payments';
-- Should NOT include: amount_paid, remaining_balance, is_fully_paid, fully_paid_date
-- SELECT * FROM payment_transactions; -- Should error (table doesn't exist)

COMMIT;

-- ============================================
-- POST-ROLLBACK CHECKS
-- ============================================
-- Run these to verify rollback:
-- 
-- 1. Check payments table structure:
--    SELECT column_name, data_type 
--    FROM information_schema.columns 
--    WHERE table_name = 'payments' 
--    ORDER BY ordinal_position;
--
-- 2. Check payment_transactions table (should not exist):
--    SELECT COUNT(*) FROM payment_transactions; 
--    -- Should error: relation "payment_transactions" does not exist
--
-- 3. Verify is_paid field restored:
--    SELECT id, amount, is_paid, payment_date, notes 
--    FROM payments 
--    LIMIT 5;
--
-- 4. Verify transaction IDs in notes:
--    SELECT id, notes FROM payments WHERE notes LIKE 'TXN:%' LIMIT 5;


-- Migration: Add Partial Payment Support
-- Description: Adds partial payment tracking and transaction management
-- Date: 2024
-- Run this in your Neon database console

BEGIN;

-- ============================================
-- STEP 1: Add new columns to payments table
-- ============================================
ALTER TABLE payments 
ADD COLUMN IF NOT EXISTS amount_paid INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS remaining_balance INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS is_fully_paid BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS fully_paid_date TIMESTAMP NULL;

-- ============================================
-- STEP 2: Backfill existing payment data
-- ============================================
-- For payments already marked as paid, set amount_paid = amount
UPDATE payments 
SET 
    amount_paid = amount,
    remaining_balance = 0,
    is_fully_paid = TRUE,
    fully_paid_date = payment_date
WHERE is_paid = TRUE AND payment_date IS NOT NULL;

-- For unpaid payments, set defaults
UPDATE payments 
SET 
    amount_paid = 0,
    remaining_balance = amount,
    is_fully_paid = FALSE
WHERE is_paid = FALSE OR payment_date IS NULL;

-- ============================================
-- STEP 3: Create payment_transactions table
-- ============================================
CREATE TABLE IF NOT EXISTS payment_transactions (
    id SERIAL PRIMARY KEY,
    payment_id INT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    transaction_id VARCHAR(255) NOT NULL,
    amount INT NULL,                           -- NULL until owner verifies
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP NULL,                -- NULL until owner verifies
    verified_by_user_id INT NULL REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Ensure unique transaction ID per payment
    CONSTRAINT unique_txn_per_payment UNIQUE (payment_id, transaction_id)
);

-- ============================================
-- STEP 4: Add indexes for performance
-- ============================================
CREATE INDEX IF NOT EXISTS idx_payments_tenant_due_date ON payments(tenant_id, due_date);
CREATE INDEX IF NOT EXISTS idx_payments_is_fully_paid ON payments(is_fully_paid);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_payment_id ON payment_transactions(payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_transaction_id ON payment_transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_verified_at ON payment_transactions(verified_at);

-- ============================================
-- STEP 5: Add constraint to ensure data integrity
-- ============================================
-- Ensure remaining_balance = amount - amount_paid (can be done via trigger or in application)
-- For now, we'll enforce in application code

COMMIT;

-- ============================================
-- VERIFICATION QUERIES
-- ============================================
-- Run these to verify migration:
-- SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'payments';
-- SELECT * FROM payment_transactions LIMIT 5;
-- SELECT COUNT(*) FROM payments WHERE is_fully_paid = TRUE;
-- SELECT COUNT(*) FROM payments WHERE remaining_balance > 0;


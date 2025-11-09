-- Migration: Add Payment Label Support
-- Description: Adds label column to payments table for categorizing payments (rent, water_bill, current_bill, maintenance)
-- Date: 2024

BEGIN;

-- ============================================
-- STEP 1: Add label column to payments table
-- ============================================
ALTER TABLE payments 
ADD COLUMN IF NOT EXISTS label VARCHAR(50) DEFAULT 'rent';

-- ============================================
-- STEP 2: Backfill existing payment data
-- ============================================
-- All existing payments are rent payments
UPDATE payments 
SET label = 'rent'
WHERE label IS NULL OR label = '';

-- ============================================
-- STEP 3: Add index for filtering by label
-- ============================================
CREATE INDEX IF NOT EXISTS idx_payments_label ON payments(label);
CREATE INDEX IF NOT EXISTS idx_payments_tenant_label ON payments(tenant_id, label);

-- ============================================
-- STEP 4: Add constraint to ensure valid labels
-- ============================================
-- Note: We'll enforce valid labels in application code
-- Valid labels: 'rent', 'water_bill', 'current_bill', 'maintenance'

COMMIT;

-- ============================================
-- VERIFICATION QUERIES
-- ============================================
-- Run these to verify migration:
-- SELECT column_name, data_type, column_default FROM information_schema.columns WHERE table_name = 'payments' AND column_name = 'label';
-- SELECT label, COUNT(*) FROM payments GROUP BY label;
-- SELECT * FROM payments WHERE label != 'rent' LIMIT 5;


-- Migration: Add notifications table for tracking notification history
-- Created: For notification service feature

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    recipient VARCHAR(20) NOT NULL, -- 'owner' or 'tenant'
    tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
    payment_id INTEGER REFERENCES payments(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    sent_at TIMESTAMP,
    sent_via VARCHAR(50) NOT NULL DEFAULT 'telegram',
    sent_to VARCHAR(255) NOT NULL, -- e.g., telegram chat ID
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient queries
CREATE INDEX IF NOT EXISTS idx_notifications_tenant_id ON notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_payment_id ON notifications(payment_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_type_recipient ON notifications(type, recipient);


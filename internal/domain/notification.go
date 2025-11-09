package domain

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeDueDateReminder NotificationType = "due_date_reminder"
)

// NotificationRecipient represents who should receive the notification
type NotificationRecipient string

const (
	NotificationRecipientOwner  NotificationRecipient = "owner"
	NotificationRecipientTenant NotificationRecipient = "tenant"
)

// Notification represents a notification record
type Notification struct {
	ID        int                   `json:"id" db:"id"`
	Type      NotificationType      `json:"type" db:"type"`
	Recipient NotificationRecipient `json:"recipient" db:"recipient"`
	TenantID  *int                  `json:"tenant_id,omitempty" db:"tenant_id"`
	PaymentID *int                  `json:"payment_id,omitempty" db:"payment_id"`
	Message   string                `json:"message" db:"message"`
	SentAt    *time.Time            `json:"sent_at,omitempty" db:"sent_at"`
	SentVia   string                `json:"sent_via" db:"sent_via"` // e.g., "telegram"
	SentTo    string                `json:"sent_to" db:"sent_to"`   // e.g., telegram chat ID
	Error     string                `json:"error,omitempty" db:"error"`
	CreatedAt time.Time             `json:"created_at" db:"created_at"`
}

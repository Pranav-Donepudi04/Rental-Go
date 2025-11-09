package interfaces

import "backend-form/m/internal/domain"

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	CreateNotification(notification *domain.Notification) error
	GetNotificationByID(id int) (*domain.Notification, error)
	GetNotificationsByTenantID(tenantID int) ([]*domain.Notification, error)
	UpdateNotification(notification *domain.Notification) error
}

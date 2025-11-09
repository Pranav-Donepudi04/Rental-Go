package repository

import (
	domain "backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"database/sql"
	"fmt"
)

// PostgresNotificationRepository implements NotificationRepository interface
type PostgresNotificationRepository struct {
	db *sql.DB
}

// NewPostgresNotificationRepository creates a new PostgresNotificationRepository
func NewPostgresNotificationRepository(db *sql.DB) interfaces.NotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

// CreateNotification creates a new notification
func (r *PostgresNotificationRepository) CreateNotification(notification *domain.Notification) error {
	query := `
		INSERT INTO notifications (type, recipient, tenant_id, payment_id, message, sent_at, sent_via, sent_to, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	var tenantID sql.NullInt64
	var paymentID sql.NullInt64
	var sentAt sql.NullTime

	if notification.TenantID != nil {
		tenantID = sql.NullInt64{Int64: int64(*notification.TenantID), Valid: true}
	}
	if notification.PaymentID != nil {
		paymentID = sql.NullInt64{Int64: int64(*notification.PaymentID), Valid: true}
	}
	if notification.SentAt != nil {
		sentAt = sql.NullTime{Time: *notification.SentAt, Valid: true}
	}

	err := r.db.QueryRow(query,
		notification.Type,
		notification.Recipient,
		tenantID,
		paymentID,
		notification.Message,
		sentAt,
		notification.SentVia,
		notification.SentTo,
		notification.Error,
		notification.CreatedAt,
	).Scan(&notification.ID)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetNotificationByID returns a notification by ID
func (r *PostgresNotificationRepository) GetNotificationByID(id int) (*domain.Notification, error) {
	query := `
		SELECT id, type, recipient, tenant_id, payment_id, message, sent_at, sent_via, sent_to, error, created_at
		FROM notifications
		WHERE id = $1`

	notification := &domain.Notification{}
	var tenantID sql.NullInt64
	var paymentID sql.NullInt64
	var sentAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&notification.ID,
		&notification.Type,
		&notification.Recipient,
		&tenantID,
		&paymentID,
		&notification.Message,
		&sentAt,
		&notification.SentVia,
		&notification.SentTo,
		&notification.Error,
		&notification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if tenantID.Valid {
		tenantIDInt := int(tenantID.Int64)
		notification.TenantID = &tenantIDInt
	}
	if paymentID.Valid {
		paymentIDInt := int(paymentID.Int64)
		notification.PaymentID = &paymentIDInt
	}
	if sentAt.Valid {
		notification.SentAt = &sentAt.Time
	}

	return notification, nil
}

// GetNotificationsByTenantID returns all notifications for a tenant
func (r *PostgresNotificationRepository) GetNotificationsByTenantID(tenantID int) ([]*domain.Notification, error) {
	query := `
		SELECT id, type, recipient, tenant_id, payment_id, message, sent_at, sent_via, sent_to, error, created_at
		FROM notifications
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		notification := &domain.Notification{}
		var tenantID sql.NullInt64
		var paymentID sql.NullInt64
		var sentAt sql.NullTime

		err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.Recipient,
			&tenantID,
			&paymentID,
			&notification.Message,
			&sentAt,
			&notification.SentVia,
			&notification.SentTo,
			&notification.Error,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if tenantID.Valid {
			tenantIDInt := int(tenantID.Int64)
			notification.TenantID = &tenantIDInt
		}
		if paymentID.Valid {
			paymentIDInt := int(paymentID.Int64)
			notification.PaymentID = &paymentIDInt
		}
		if sentAt.Valid {
			notification.SentAt = &sentAt.Time
		}

		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// UpdateNotification updates a notification
func (r *PostgresNotificationRepository) UpdateNotification(notification *domain.Notification) error {
	query := `
		UPDATE notifications
		SET sent_at = $1, sent_via = $2, sent_to = $3, error = $4
		WHERE id = $5`

	var sentAt sql.NullTime
	if notification.SentAt != nil {
		sentAt = sql.NullTime{Time: *notification.SentAt, Valid: true}
	}

	_, err := r.db.Exec(query,
		sentAt,
		notification.SentVia,
		notification.SentTo,
		notification.Error,
		notification.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return nil
}

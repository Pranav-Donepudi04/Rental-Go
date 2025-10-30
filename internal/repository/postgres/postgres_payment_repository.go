package repository

import (
	domain "backend-form/m/internal/domain"
	"backend-form/m/internal/repository/interfaces"
	"database/sql"
	"fmt"
	"time"
)

// PostgresPaymentRepository implements PaymentRepository interface
type PostgresPaymentRepository struct {
	db *sql.DB
}

// NewPostgresPaymentRepository creates a new PostgresPaymentRepository
func NewPostgresPaymentRepository(db *sql.DB) interfaces.PaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// CreatePayment creates a new payment
func (r *PostgresPaymentRepository) CreatePayment(payment *domain.Payment) error {
	query := `
		INSERT INTO payments (tenant_id, unit_id, amount, payment_date, due_date, is_paid, payment_method, upi_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		payment.TenantID,
		payment.UnitID,
		payment.Amount,
		payment.PaymentDate,
		payment.DueDate,
		payment.IsPaid,
		payment.PaymentMethod,
		payment.UPIID,
		payment.Notes,
	).Scan(&payment.ID, &payment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// GetPaymentByID returns a payment by ID
func (r *PostgresPaymentRepository) GetPaymentByID(id int) (*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, unit_id, amount, payment_date, due_date, is_paid, 
		       payment_method, upi_id, notes, created_at
		FROM payments
		WHERE id = $1`

	payment := &domain.Payment{}
	var paymentDate sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.UnitID,
		&payment.Amount,
		&paymentDate,
		&payment.DueDate,
		&payment.IsPaid,
		&payment.PaymentMethod,
		&payment.UPIID,
		&payment.Notes,
		&payment.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if paymentDate.Valid {
		payment.PaymentDate = &paymentDate.Time
	}

	return payment, nil
}

// DeletePaymentsByTenantID deletes all payments for a specific tenant
func (r *PostgresPaymentRepository) DeletePaymentsByTenantID(tenantID int) error {
	query := `DELETE FROM payments WHERE tenant_id = $1`
	_, err := r.db.Exec(query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete payments for tenant %d: %w", tenantID, err)
	}
	return nil
}

// GetAllPayments returns all payments
func (r *PostgresPaymentRepository) GetAllPayments() ([]*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, unit_id, amount, payment_date, due_date, is_paid, 
		       payment_method, upi_id, notes, created_at
		FROM payments
		ORDER BY due_date DESC, created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		var paymentDate sql.NullTime

		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.UnitID,
			&payment.Amount,
			&paymentDate,
			&payment.DueDate,
			&payment.IsPaid,
			&payment.PaymentMethod,
			&payment.UPIID,
			&payment.Notes,
			&payment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		if paymentDate.Valid {
			payment.PaymentDate = &paymentDate.Time
		}

		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	return payments, nil
}

// UpdatePayment updates payment information
func (r *PostgresPaymentRepository) UpdatePayment(payment *domain.Payment) error {
	query := `
		UPDATE payments 
		SET tenant_id = $1, unit_id = $2, amount = $3, payment_date = $4, 
		    due_date = $5, is_paid = $6, payment_method = $7, upi_id = $8, notes = $9
		WHERE id = $10`

	result, err := r.db.Exec(query,
		payment.TenantID,
		payment.UnitID,
		payment.Amount,
		payment.PaymentDate,
		payment.DueDate,
		payment.IsPaid,
		payment.PaymentMethod,
		payment.UPIID,
		payment.Notes,
		payment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment with ID %d not found", payment.ID)
	}

	return nil
}

// DeletePayment deletes a payment
func (r *PostgresPaymentRepository) DeletePayment(id int) error {
	query := `DELETE FROM payments WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment with ID %d not found", id)
	}

	return nil
}

// GetPaymentsByTenantID returns payments for a specific tenant
func (r *PostgresPaymentRepository) GetPaymentsByTenantID(tenantID int) ([]*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, unit_id, amount, payment_date, due_date, is_paid, 
		       payment_method, upi_id, notes, created_at
		FROM payments
		WHERE tenant_id = $1
		ORDER BY due_date DESC, created_at DESC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments by tenant: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		var paymentDate sql.NullTime

		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.UnitID,
			&payment.Amount,
			&paymentDate,
			&payment.DueDate,
			&payment.IsPaid,
			&payment.PaymentMethod,
			&payment.UPIID,
			&payment.Notes,
			&payment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		if paymentDate.Valid {
			payment.PaymentDate = &paymentDate.Time
		}

		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	return payments, nil
}

// GetPaymentByTenantAndMonth returns payment for a specific tenant and month
func (r *PostgresPaymentRepository) GetPaymentByTenantAndMonth(tenantID int, month time.Month, year int) (*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, unit_id, amount, payment_date, due_date, is_paid, 
		       payment_method, upi_id, notes, created_at
		FROM payments
		WHERE tenant_id = $1 AND EXTRACT(MONTH FROM due_date) = $2 AND EXTRACT(YEAR FROM due_date) = $3`

	payment := &domain.Payment{}
	var paymentDate sql.NullTime

	err := r.db.QueryRow(query, tenantID, int(month), year).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.UnitID,
		&payment.Amount,
		&paymentDate,
		&payment.DueDate,
		&payment.IsPaid,
		&payment.PaymentMethod,
		&payment.UPIID,
		&payment.Notes,
		&payment.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No payment found for this month
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if paymentDate.Valid {
		payment.PaymentDate = &paymentDate.Time
	}

	return payment, nil
}

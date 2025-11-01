package repository

import (
	domain "backend-form/m/internal/domain"
	"database/sql"
	"fmt"
	"time"
)

// ============================================
// Payment Transaction Methods
// ============================================

// CreatePaymentTransaction creates a new payment transaction record
func (r *PostgresPaymentRepository) CreatePaymentTransaction(tx *domain.PaymentTransaction) error {
	query := `
		INSERT INTO payment_transactions (payment_id, transaction_id, amount, submitted_at, verified_at, verified_by_user_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		tx.PaymentID,
		tx.TransactionID,
		tx.Amount,
		tx.SubmittedAt,
		tx.VerifiedAt,
		tx.VerifiedByUserID,
		tx.Notes,
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create payment transaction: %w", err)
	}

	return nil
}

// GetPaymentTransactionsByPaymentID returns all transactions for a payment
func (r *PostgresPaymentRepository) GetPaymentTransactionsByPaymentID(paymentID int) ([]*domain.PaymentTransaction, error) {
	query := `
		SELECT id, payment_id, transaction_id, amount, submitted_at, verified_at, verified_by_user_id, notes, created_at
		FROM payment_transactions
		WHERE payment_id = $1
		ORDER BY submitted_at DESC`

	rows, err := r.db.Query(query, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.PaymentTransaction
	for rows.Next() {
		tx := &domain.PaymentTransaction{}
		var amount sql.NullInt64
		var verifiedAt sql.NullTime
		var verifiedByUserID sql.NullInt64

		err := rows.Scan(
			&tx.ID,
			&tx.PaymentID,
			&tx.TransactionID,
			&amount,
			&tx.SubmittedAt,
			&verifiedAt,
			&verifiedByUserID,
			&tx.Notes,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment transaction: %w", err)
		}

		if amount.Valid {
			amt := int(amount.Int64)
			tx.Amount = &amt
		}
		if verifiedAt.Valid {
			tx.VerifiedAt = &verifiedAt.Time
		}
		if verifiedByUserID.Valid {
			uid := int(verifiedByUserID.Int64)
			tx.VerifiedByUserID = &uid
		}

		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment transactions: %w", err)
	}

	return transactions, nil
}

// GetTransactionByPaymentAndID returns a specific transaction by payment ID and transaction ID
func (r *PostgresPaymentRepository) GetTransactionByPaymentAndID(paymentID int, transactionID string) (*domain.PaymentTransaction, error) {
	query := `
		SELECT id, payment_id, transaction_id, amount, submitted_at, verified_at, verified_by_user_id, notes, created_at
		FROM payment_transactions
		WHERE payment_id = $1 AND transaction_id = $2`

	tx := &domain.PaymentTransaction{}
	var amount sql.NullInt64
	var verifiedAt sql.NullTime
	var verifiedByUserID sql.NullInt64

	err := r.db.QueryRow(query, paymentID, transactionID).Scan(
		&tx.ID,
		&tx.PaymentID,
		&tx.TransactionID,
		&amount,
		&tx.SubmittedAt,
		&verifiedAt,
		&verifiedByUserID,
		&tx.Notes,
		&tx.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Transaction not found
		}
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}

	if amount.Valid {
		amt := int(amount.Int64)
		tx.Amount = &amt
	}
	if verifiedAt.Valid {
		tx.VerifiedAt = &verifiedAt.Time
	}
	if verifiedByUserID.Valid {
		uid := int(verifiedByUserID.Int64)
		tx.VerifiedByUserID = &uid
	}

	return tx, nil
}

// GetTransactionByID returns a transaction by transaction ID (efficient lookup)
func (r *PostgresPaymentRepository) GetTransactionByID(transactionID string) (*domain.PaymentTransaction, error) {
	query := `
		SELECT id, payment_id, transaction_id, amount, submitted_at, verified_at, verified_by_user_id, notes, created_at
		FROM payment_transactions
		WHERE transaction_id = $1`

	tx := &domain.PaymentTransaction{}
	var amount sql.NullInt64
	var verifiedAt sql.NullTime
	var verifiedByUserID sql.NullInt64

	err := r.db.QueryRow(query, transactionID).Scan(
		&tx.ID,
		&tx.PaymentID,
		&tx.TransactionID,
		&amount,
		&tx.SubmittedAt,
		&verifiedAt,
		&verifiedByUserID,
		&tx.Notes,
		&tx.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Transaction not found
		}
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	if amount.Valid {
		amt := int(amount.Int64)
		tx.Amount = &amt
	}
	if verifiedAt.Valid {
		tx.VerifiedAt = &verifiedAt.Time
	}
	if verifiedByUserID.Valid {
		uid := int(verifiedByUserID.Int64)
		tx.VerifiedByUserID = &uid
	}

	return tx, nil
}

// GetPendingVerifications returns all pending (unverified) transactions for a tenant
func (r *PostgresPaymentRepository) GetPendingVerifications(tenantID int) ([]*domain.PaymentTransaction, error) {
	query := `
		SELECT pt.id, pt.payment_id, pt.transaction_id, pt.amount, pt.submitted_at, 
		       pt.verified_at, pt.verified_by_user_id, pt.notes, pt.created_at
		FROM payment_transactions pt
		INNER JOIN payments p ON pt.payment_id = p.id
		WHERE p.tenant_id = $1 AND pt.verified_at IS NULL
		ORDER BY pt.submitted_at DESC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending verifications: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.PaymentTransaction
	for rows.Next() {
		tx := &domain.PaymentTransaction{}
		var amount sql.NullInt64
		var verifiedAt sql.NullTime
		var verifiedByUserID sql.NullInt64

		err := rows.Scan(
			&tx.ID,
			&tx.PaymentID,
			&tx.TransactionID,
			&amount,
			&tx.SubmittedAt,
			&verifiedAt,
			&verifiedByUserID,
			&tx.Notes,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment transaction: %w", err)
		}

		if amount.Valid {
			amt := int(amount.Int64)
			tx.Amount = &amt
		}
		if verifiedAt.Valid {
			tx.VerifiedAt = &verifiedAt.Time
		}
		if verifiedByUserID.Valid {
			uid := int(verifiedByUserID.Int64)
			tx.VerifiedByUserID = &uid
		}

		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending verifications: %w", err)
	}

	return transactions, nil
}

// VerifyTransaction verifies a transaction by updating its amount and verification details
// This also updates the payment's amount_paid and remaining_balance
func (r *PostgresPaymentRepository) VerifyTransaction(transactionID string, amount int, verifiedByUserID int) error {
	// Use a transaction to ensure atomicity
	dbTx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	// Get the payment transaction
	var paymentID int
	var currentAmount sql.NullInt64
	err = dbTx.QueryRow(`
		SELECT payment_id, amount 
		FROM payment_transactions 
		WHERE transaction_id = $1`,
		transactionID,
	).Scan(&paymentID, &currentAmount)

	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Check if already verified
	if currentAmount.Valid {
		return fmt.Errorf("transaction already verified")
	}

	// Update transaction
	now := time.Now()
	_, err = dbTx.Exec(`
		UPDATE payment_transactions 
		SET amount = $1, verified_at = $2, verified_by_user_id = $3
		WHERE transaction_id = $4`,
		amount, now, verifiedByUserID, transactionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Update payment: add amount to amount_paid and recalculate balance
	_, err = dbTx.Exec(`
		UPDATE payments 
		SET amount_paid = amount_paid + $1,
		    remaining_balance = amount - (amount_paid + $1),
		    is_fully_paid = (amount - (amount_paid + $1) <= 0),
		    fully_paid_date = CASE 
		        WHEN (amount - (amount_paid + $1) <= 0) AND fully_paid_date IS NULL THEN $2
		        ELSE fully_paid_date
		    END
		WHERE id = $3`,
		amount, now, paymentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if err = dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RejectTransaction deletes a pending transaction (rejects it)
func (r *PostgresPaymentRepository) RejectTransaction(transactionID string) error {
	// Check if transaction exists and is not verified
	var verifiedAt sql.NullTime
	err := r.db.QueryRow(`
		SELECT verified_at 
		FROM payment_transactions 
		WHERE transaction_id = $1`,
		transactionID,
	).Scan(&verifiedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaction not found")
		}
		return fmt.Errorf("failed to check transaction: %w", err)
	}

	// If already verified, cannot reject
	if verifiedAt.Valid {
		return fmt.Errorf("cannot reject already verified transaction")
	}

	// Delete the transaction
	_, err = r.db.Exec(`
		DELETE FROM payment_transactions 
		WHERE transaction_id = $1`,
		transactionID,
	)
	if err != nil {
		return fmt.Errorf("failed to reject transaction: %w", err)
	}

	return nil
}

// ============================================
// Helper Methods for Auto-Create
// ============================================

// GetLatestPaymentByTenantID returns the latest payment for a tenant (by due date)
func (r *PostgresPaymentRepository) GetLatestPaymentByTenantID(tenantID int) (*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, unit_id, amount, amount_paid, remaining_balance, payment_date, 
		       due_date, is_paid, is_fully_paid, fully_paid_date, payment_method, upi_id, notes, created_at
		FROM payments
		WHERE tenant_id = $1
		ORDER BY due_date DESC
		LIMIT 1`

	payment := &domain.Payment{}
	var paymentDate sql.NullTime
	var fullyPaidDate sql.NullTime

	err := r.db.QueryRow(query, tenantID).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.UnitID,
		&payment.Amount,
		&payment.AmountPaid,
		&payment.RemainingBalance,
		&paymentDate,
		&payment.DueDate,
		&payment.IsPaid,
		&payment.IsFullyPaid,
		&fullyPaidDate,
		&payment.PaymentMethod,
		&payment.UPIID,
		&payment.Notes,
		&payment.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No payment found
		}
		return nil, fmt.Errorf("failed to get latest payment: %w", err)
	}

	if paymentDate.Valid {
		payment.PaymentDate = &paymentDate.Time
	}
	if fullyPaidDate.Valid {
		payment.FullyPaidDate = &fullyPaidDate.Time
	}

	return payment, nil
}

// GetUnpaidPaymentsByTenantID returns all unpaid payments for a tenant
func (r *PostgresPaymentRepository) GetUnpaidPaymentsByTenantID(tenantID int) ([]*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, unit_id, amount, amount_paid, remaining_balance, payment_date, 
		       due_date, is_paid, is_fully_paid, fully_paid_date, payment_method, upi_id, notes, created_at
		FROM payments
		WHERE tenant_id = $1 AND is_fully_paid = FALSE
		ORDER BY due_date ASC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unpaid payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		var paymentDate sql.NullTime
		var fullyPaidDate sql.NullTime

		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.UnitID,
			&payment.Amount,
			&payment.AmountPaid,
			&payment.RemainingBalance,
			&paymentDate,
			&payment.DueDate,
			&payment.IsPaid,
			&payment.IsFullyPaid,
			&fullyPaidDate,
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
		if fullyPaidDate.Valid {
			payment.FullyPaidDate = &fullyPaidDate.Time
		}

		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unpaid payments: %w", err)
	}

	return payments, nil
}

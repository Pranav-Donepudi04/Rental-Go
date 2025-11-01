package repository

import (
	"fmt"
)

// CleanupExpiredSessions deletes all expired sessions from the database
// This is safe to run as sessions have no cascading dependencies
func (r *PostgresSessionRepository) CleanupExpiredSessions() (int64, error) {
	const q = `DELETE FROM sessions WHERE expires_at < NOW()`
	result, err := r.db.Exec(q)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired sessions: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}
	return rowsAffected, nil
}

// GetSessionStats returns statistics about sessions
func (r *PostgresSessionRepository) GetSessionStats() (map[string]interface{}, error) {
	const q = `
		SELECT 
			COUNT(*) as total_sessions,
			COUNT(*) FILTER (WHERE expires_at < NOW()) as expired_sessions,
			COUNT(*) FILTER (WHERE expires_at >= NOW()) as active_sessions
		FROM sessions
	`
	var total, expired, active int64
	err := r.db.QueryRow(q).Scan(&total, &expired, &active)
	if err != nil {
		return nil, fmt.Errorf("get session stats: %w", err)
	}
	stats := map[string]interface{}{
		"total":   total,
		"expired": expired,
		"active":  active,
	}
	return stats, nil
}

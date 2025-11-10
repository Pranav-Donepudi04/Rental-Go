package metrics

import (
	"sync"
	"time"
)

// Metrics tracks application metrics
type Metrics struct {
	mu sync.RWMutex

	// HTTP metrics
	HTTPRequestsTotal   map[string]int64           // endpoint -> count
	HTTPRequestDuration map[string][]time.Duration // endpoint -> durations
	HTTPErrorsTotal     map[string]int64           // endpoint -> error count

	// Business metrics
	TenantsCreated       int64
	PaymentsProcessed    int64
	PaymentsVerified     int64
	TransactionsRejected int64
	LoginsTotal          int64
	LoginFailures        int64

	// Database metrics
	DBQueriesTotal  int64
	DBQueryDuration []time.Duration
	DBErrorsTotal   int64
}

var globalMetrics *Metrics
var once sync.Once

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		globalMetrics = &Metrics{
			HTTPRequestsTotal:   make(map[string]int64),
			HTTPRequestDuration: make(map[string][]time.Duration),
			HTTPErrorsTotal:     make(map[string]int64),
		}
	})
	return globalMetrics
}

// IncrementHTTPRequest increments the request count for an endpoint
func (m *Metrics) IncrementHTTPRequest(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HTTPRequestsTotal[endpoint]++
}

// RecordHTTPDuration records the duration of an HTTP request
func (m *Metrics) RecordHTTPDuration(endpoint string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.HTTPRequestDuration[endpoint] == nil {
		m.HTTPRequestDuration[endpoint] = make([]time.Duration, 0, 100)
	}
	// Keep only last 100 durations per endpoint
	if len(m.HTTPRequestDuration[endpoint]) >= 100 {
		m.HTTPRequestDuration[endpoint] = m.HTTPRequestDuration[endpoint][1:]
	}
	m.HTTPRequestDuration[endpoint] = append(m.HTTPRequestDuration[endpoint], duration)
}

// IncrementHTTPError increments the error count for an endpoint
func (m *Metrics) IncrementHTTPError(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HTTPErrorsTotal[endpoint]++
}

// IncrementTenantCreated increments the tenant creation counter
func (m *Metrics) IncrementTenantCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TenantsCreated++
}

// IncrementPaymentProcessed increments the payment processed counter
func (m *Metrics) IncrementPaymentProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PaymentsProcessed++
}

// IncrementPaymentVerified increments the payment verified counter
func (m *Metrics) IncrementPaymentVerified() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PaymentsVerified++
}

// IncrementTransactionRejected increments the transaction rejected counter
func (m *Metrics) IncrementTransactionRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TransactionsRejected++
}

// IncrementLogin increments the login counter
func (m *Metrics) IncrementLogin() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LoginsTotal++
}

// IncrementLoginFailure increments the login failure counter
func (m *Metrics) IncrementLoginFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LoginFailures++
}

// IncrementDBQuery increments the database query counter
func (m *Metrics) IncrementDBQuery() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DBQueriesTotal++
}

// RecordDBDuration records the duration of a database query
func (m *Metrics) RecordDBDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Keep only last 1000 durations
	if len(m.DBQueryDuration) >= 1000 {
		m.DBQueryDuration = m.DBQueryDuration[1:]
	}
	m.DBQueryDuration = append(m.DBQueryDuration, duration)
}

// IncrementDBError increments the database error counter
func (m *Metrics) IncrementDBError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DBErrorsTotal++
}

// GetSnapshot returns a snapshot of current metrics (thread-safe)
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := MetricsSnapshot{
		HTTPRequestsTotal:    make(map[string]int64),
		HTTPErrorsTotal:      make(map[string]int64),
		TenantsCreated:       m.TenantsCreated,
		PaymentsProcessed:    m.PaymentsProcessed,
		PaymentsVerified:     m.PaymentsVerified,
		TransactionsRejected: m.TransactionsRejected,
		LoginsTotal:          m.LoginsTotal,
		LoginFailures:        m.LoginFailures,
		DBQueriesTotal:       m.DBQueriesTotal,
		DBErrorsTotal:        m.DBErrorsTotal,
	}

	for k, v := range m.HTTPRequestsTotal {
		snapshot.HTTPRequestsTotal[k] = v
	}
	for k, v := range m.HTTPErrorsTotal {
		snapshot.HTTPErrorsTotal[k] = v
	}

	return snapshot
}

// MetricsSnapshot is a read-only snapshot of metrics
type MetricsSnapshot struct {
	HTTPRequestsTotal    map[string]int64
	HTTPErrorsTotal      map[string]int64
	TenantsCreated       int64
	PaymentsProcessed    int64
	PaymentsVerified     int64
	TransactionsRejected int64
	LoginsTotal          int64
	LoginFailures        int64
	DBQueriesTotal       int64
	DBErrorsTotal        int64
}


package service

import (
	"backend-form/m/internal/cache"
	"backend-form/m/internal/domain"
	"fmt"
	"time"
)

// DashboardService handles dashboard data aggregation
// This service encapsulates all logic for gathering dashboard information
type DashboardService struct {
	unitService         *UnitService
	tenantService       *TenantService
	paymentQueryService *PaymentQueryService
	cache               *cache.Cache
}

// NewDashboardService creates a new DashboardService
func NewDashboardService(
	unitService *UnitService,
	tenantService *TenantService,
	paymentQueryService *PaymentQueryService,
) *DashboardService {
	// Cache dashboard data for 30 seconds
	return &DashboardService{
		unitService:         unitService,
		tenantService:       tenantService,
		paymentQueryService: paymentQueryService,
		cache:               cache.NewCache(30 * time.Second),
	}
}

// DashboardData represents all data needed for the dashboard view
type DashboardData struct {
	Units          []*domain.Unit    `json:"units"`
	Tenants        []*domain.Tenant  `json:"tenants"`
	Payments       []*domain.Payment `json:"payments"`
	UnitSummary    *RentalSummary    `json:"unit_summary"`
	PaymentSummary *PaymentSummary   `json:"payment_summary"`
}

// GetDashboardData returns all data needed for the dashboard page
// Results are cached for 30 seconds to reduce database load
func (s *DashboardService) GetDashboardData() (*DashboardData, error) {
	// Check cache first
	if cached, found := s.cache.Get("dashboard_data"); found {
		if data, ok := cached.(*DashboardData); ok {
			return data, nil
		}
	}

	// Cache miss - load from database
	units, err := s.unitService.GetAllUnits()
	if err != nil {
		return nil, fmt.Errorf("get units: %w", err)
	}

	tenants, err := s.tenantService.GetAllTenants()
	if err != nil {
		return nil, fmt.Errorf("get tenants: %w", err)
	}

	payments, err := s.paymentQueryService.GetAllPayments()
	if err != nil {
		return nil, fmt.Errorf("get payments: %w", err)
	}

	unitSummary, err := s.unitService.GetRentalSummary()
	if err != nil {
		return nil, fmt.Errorf("get unit summary: %w", err)
	}

	paymentSummary, err := s.paymentQueryService.GetPaymentSummary()
	if err != nil {
		return nil, fmt.Errorf("get payment summary: %w", err)
	}

	data := &DashboardData{
		Units:          units,
		Tenants:        tenants,
		Payments:       payments,
		UnitSummary:    unitSummary,
		PaymentSummary: paymentSummary,
	}

	// Store in cache
	s.cache.Set("dashboard_data", data)

	return data, nil
}

// InvalidateDashboardCache clears the dashboard cache
// Call this when dashboard data changes (e.g., new tenant, payment update)
func (s *DashboardService) InvalidateDashboardCache() {
	s.cache.Delete("dashboard_data")
}

// DashboardSummary represents just the summary data (for JSON API)
type DashboardSummary struct {
	Units    *RentalSummary  `json:"units"`
	Payments *PaymentSummary `json:"payments"`
}

// GetDashboardSummary returns only the summary data (for JSON API endpoint)
func (s *DashboardService) GetDashboardSummary() (*DashboardSummary, error) {
	unitSummary, err := s.unitService.GetRentalSummary()
	if err != nil {
		return nil, fmt.Errorf("get unit summary: %w", err)
	}

	paymentSummary, err := s.paymentQueryService.GetPaymentSummary()
	if err != nil {
		return nil, fmt.Errorf("get payment summary: %w", err)
	}

	return &DashboardSummary{
		Units:    unitSummary,
		Payments: paymentSummary,
	}, nil
}

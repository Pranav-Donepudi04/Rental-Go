package service

import (
	"backend-form/m/internal/domain"
	"fmt"
)

// DashboardService handles dashboard data aggregation
// This service encapsulates all logic for gathering dashboard information
type DashboardService struct {
	unitService         *UnitService
	tenantService       *TenantService
	paymentQueryService *PaymentQueryService
}

// NewDashboardService creates a new DashboardService
func NewDashboardService(
	unitService *UnitService,
	tenantService *TenantService,
	paymentQueryService *PaymentQueryService,
) *DashboardService {
	return &DashboardService{
		unitService:         unitService,
		tenantService:       tenantService,
		paymentQueryService: paymentQueryService,
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
func (s *DashboardService) GetDashboardData() (*DashboardData, error) {
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

	return &DashboardData{
		Units:          units,
		Tenants:        tenants,
		Payments:       payments,
		UnitSummary:    unitSummary,
		PaymentSummary: paymentSummary,
	}, nil
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

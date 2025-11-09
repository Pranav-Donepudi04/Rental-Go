package service

import (
	"backend-form/m/internal/domain"
	interfaces "backend-form/m/internal/repository/interfaces"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
)

// NotificationService handles notification-related business logic
type NotificationService struct {
	notificationRepo interfaces.NotificationRepository
	paymentRepo      interfaces.PaymentRepository
	tenantRepo       interfaces.TenantRepository
	unitRepo         interfaces.UnitRepository
	notifier         *notify.Notify
	telegramBotToken string
	ownerChatID      string
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(
	notificationRepo interfaces.NotificationRepository,
	paymentRepo interfaces.PaymentRepository,
	tenantRepo interfaces.TenantRepository,
	unitRepo interfaces.UnitRepository,
	telegramBotToken string,
	ownerChatID string,
) *NotificationService {
	// Initialize notify library
	notifier := notify.New()

	// Initialize Telegram service if token is provided
	if telegramBotToken != "" && ownerChatID != "" {
		telegramService, err := telegram.New(telegramBotToken)
		if err == nil {
			// Convert owner chat ID to int64
			if ownerChatIDInt, err := strconv.ParseInt(ownerChatID, 10, 64); err == nil {
				telegramService.AddReceivers(ownerChatIDInt)
				notifier.UseServices(telegramService)
			}
		}
	}

	return &NotificationService{
		notificationRepo: notificationRepo,
		paymentRepo:      paymentRepo,
		tenantRepo:       tenantRepo,
		unitRepo:         unitRepo,
		notifier:         notifier,
		telegramBotToken: telegramBotToken,
		ownerChatID:      ownerChatID,
	}
}

// SendTelegramMessage sends a message via Telegram using the notify library
func (s *NotificationService) SendTelegramMessage(chatID string, message string) error {
	if s.telegramBotToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	// Convert chat ID string to int64
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID format: %w", err)
	}

	// Create a temporary notifier with the specific chat ID
	tempNotifier := notify.New()
	tempTelegramService, err := telegram.New(s.telegramBotToken)
	if err != nil {
		return fmt.Errorf("failed to create telegram service: %w", err)
	}
	tempTelegramService.AddReceivers(chatIDInt)
	tempNotifier.UseServices(tempTelegramService)

	// Send the message with context
	ctx := context.Background()
	if err := tempNotifier.Send(ctx, "Rent Reminder", message); err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

// SendDueDateReminderToOwner sends due date reminder to owner
func (s *NotificationService) SendDueDateReminderToOwner(payment *domain.Payment) error {
	// Load tenant and unit data
	tenant, err := s.tenantRepo.GetTenantByID(payment.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	unit, err := s.unitRepo.GetUnitByID(payment.UnitID)
	if err != nil {
		return fmt.Errorf("failed to get unit: %w", err)
	}

	message := fmt.Sprintf(
		"📅 Reminder: On %s, %s (%s) has to pay ₹%d",
		payment.DueDate.Format("Jan 2, 2006"),
		tenant.Name,
		unit.UnitCode,
		payment.Amount,
	)

	// Create notification record
	notification := &domain.Notification{
		Type:      domain.NotificationTypeDueDateReminder,
		Recipient: domain.NotificationRecipientOwner,
		TenantID:  &payment.TenantID,
		PaymentID: &payment.ID,
		Message:   message,
		SentVia:   "telegram",
		SentTo:    s.ownerChatID,
		CreatedAt: time.Now(),
	}

	// Try to send via Telegram
	err = s.SendTelegramMessage(s.ownerChatID, message)
	if err != nil {
		notification.Error = err.Error()
		// Still save the notification record even if sending fails
		if createErr := s.notificationRepo.CreateNotification(notification); createErr != nil {
			return fmt.Errorf("failed to create notification record: %w", createErr)
		}
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	// Mark as sent
	now := time.Now()
	notification.SentAt = &now

	// Save notification record
	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	return nil
}

// SendDueDateReminderToTenant sends due date reminder to tenant (5 days before)
// Note: Tenant Telegram chat IDs need to be configured. For now, this will skip if chat ID is not available.
// In future, you can add telegram_chat_id field to tenants or users table.
func (s *NotificationService) SendDueDateReminderToTenant(payment *domain.Payment, tenantChatID string) error {
	if tenantChatID == "" {
		// Skip if tenant chat ID not configured
		return fmt.Errorf("tenant chat ID not configured")
	}

	// Load unit data
	unit, err := s.unitRepo.GetUnitByID(payment.UnitID)
	if err != nil {
		return fmt.Errorf("failed to get unit: %w", err)
	}

	message := fmt.Sprintf(
		"📅 Reminder: Your rent payment of ₹%d for %s is due on %s. Please make the payment to %s",
		payment.Amount,
		unit.UnitCode,
		payment.DueDate.Format("Jan 2, 2006"),
		payment.UPIID,
	)

	// Create notification record
	notification := &domain.Notification{
		Type:      domain.NotificationTypeDueDateReminder,
		Recipient: domain.NotificationRecipientTenant,
		TenantID:  &payment.TenantID,
		PaymentID: &payment.ID,
		Message:   message,
		SentVia:   "telegram",
		SentTo:    tenantChatID,
		CreatedAt: time.Now(),
	}

	// Try to send via Telegram
	err = s.SendTelegramMessage(tenantChatID, message)
	if err != nil {
		notification.Error = err.Error()
		// Still save the notification record even if sending fails
		if createErr := s.notificationRepo.CreateNotification(notification); createErr != nil {
			return fmt.Errorf("failed to create notification record: %w", createErr)
		}
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	// Mark as sent
	now := time.Now()
	notification.SentAt = &now

	// Save notification record
	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	return nil
}

// CheckAndSendDueDateReminders checks for payments due today (for owner) and due in 5 days (for tenants)
func (s *NotificationService) CheckAndSendDueDateReminders() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 1. Send reminders to owner for payments due today
	paymentsDueToday, err := s.paymentRepo.GetUnpaidPaymentsByDueDate(today)
	if err != nil {
		return fmt.Errorf("failed to get payments due today: %w", err)
	}

	for _, payment := range paymentsDueToday {
		// Double-check payment is not fully paid (race condition protection)
		if payment.IsFullyPaid {
			continue
		}

		// Reload payment to ensure we have latest status
		latestPayment, err := s.paymentRepo.GetPaymentByID(payment.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to reload payment %d: %v\n", payment.ID, err)
			continue
		}

		if latestPayment.IsFullyPaid {
			continue
		}

		if err := s.SendDueDateReminderToOwner(latestPayment); err != nil {
			fmt.Printf("Warning: Failed to send reminder to owner for payment %d: %v\n", latestPayment.ID, err)
			// Continue with other payments
		}
	}

	// 2. Send reminders to tenants for payments due in 5 days
	// Logic: Tenants get reminders exactly 5 days before their due date
	// Example: If today is July 31st, we check for payments due on Aug 5th (today + 5 days)
	//          So payment due Aug 5th → Tenant notified on July 31st (5 days before)
	fiveDaysFromNow := today.AddDate(0, 0, 5)
	paymentsDueIn5Days, err := s.paymentRepo.GetUnpaidPaymentsByDueDate(fiveDaysFromNow)
	if err != nil {
		return fmt.Errorf("failed to get payments due in 5 days: %w", err)
	}

	for _, payment := range paymentsDueIn5Days {
		// Double-check payment is not fully paid
		if payment.IsFullyPaid {
			continue
		}

		// Reload payment to ensure we have latest status
		latestPayment, err := s.paymentRepo.GetPaymentByID(payment.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to reload payment %d: %v\n", payment.ID, err)
			continue
		}

		if latestPayment.IsFullyPaid {
			continue
		}

		// TODO: Get tenant's Telegram chat ID from database (you may want to add this field to tenants/users table)
		// For now, using owner chat ID for testing tenant notifications
		// In production, this should be retrieved from database per tenant
		tenantChatID := s.ownerChatID // TEMPORARY: Using owner chat ID for testing

		if tenantChatID != "" {
			if err := s.SendDueDateReminderToTenant(latestPayment, tenantChatID); err != nil {
				fmt.Printf("Warning: Failed to send reminder to tenant for payment %d: %v\n", latestPayment.ID, err)
				// Continue with other payments
			}
		}
	}

	return nil
}

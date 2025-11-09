package service

import (
	"backend-form/m/internal/logger"
	"time"

	"go.uber.org/zap"
)

// NotificationScheduler handles scheduled notification tasks
type NotificationScheduler struct {
	notificationService *NotificationService
	stopChan            chan bool
}

// NewNotificationScheduler creates a new NotificationScheduler
func NewNotificationScheduler(notificationService *NotificationService) *NotificationScheduler {
	return &NotificationScheduler{
		notificationService: notificationService,
		stopChan:            make(chan bool),
	}
}

// Start starts the scheduler to run daily at a specified time
func (s *NotificationScheduler) Start() {
	go s.run()
}

// Stop stops the scheduler
func (s *NotificationScheduler) Stop() {
	s.stopChan <- true
}

// run executes the scheduler loop
func (s *NotificationScheduler) run() {
	// Run immediately on start (for testing/initial run)
	s.checkAndSendReminders()

	// Calculate time until next 9 AM (or adjust to your preferred time)
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	if now.After(nextRun) {
		// If it's already past 9 AM today, schedule for tomorrow
		nextRun = nextRun.Add(24 * time.Hour)
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Wait until the scheduled time for first run
	time.Sleep(time.Until(nextRun))

	// Run immediately at scheduled time
	s.checkAndSendReminders()

	// Then run daily
	for {
		select {
		case <-ticker.C:
			s.checkAndSendReminders()
		case <-s.stopChan:
			logger.Info("Notification scheduler stopped")
			return
		}
	}
}

// checkAndSendReminders checks and sends due date reminders
func (s *NotificationScheduler) checkAndSendReminders() {
	logger.Info("Running daily notification check...")
	if err := s.notificationService.CheckAndSendDueDateReminders(); err != nil {
		logger.Error("Error checking and sending reminders",
			zap.Error(err),
		)
	} else {
		logger.Info("Notification check completed successfully")
	}
}

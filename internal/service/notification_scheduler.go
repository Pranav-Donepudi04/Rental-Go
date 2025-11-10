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

// Stop stops the scheduler (non-blocking)
func (s *NotificationScheduler) Stop() {
	select {
	case s.stopChan <- true:
		// Stop signal sent
	default:
		// Channel full or already stopping, ignore
	}
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

	// Wait until the scheduled time for first run, but check for stop signal periodically
	waitDuration := time.Until(nextRun)
	if waitDuration > 0 {
		waitTimer := time.NewTimer(waitDuration)
		defer waitTimer.Stop()

		// Check for stop signal every 100ms while waiting
		stopTicker := time.NewTicker(100 * time.Millisecond)
		defer stopTicker.Stop()

		for {
			select {
			case <-waitTimer.C:
				// Time reached, proceed
				goto runScheduled
			case <-s.stopChan:
				logger.Info("Notification scheduler stopped during initial wait")
				return
			case <-stopTicker.C:
				// Check if stop was requested (non-blocking check)
				select {
				case <-s.stopChan:
					logger.Info("Notification scheduler stopped during initial wait")
					return
				default:
					// Continue waiting
				}
			}
		}
	}

runScheduled:

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

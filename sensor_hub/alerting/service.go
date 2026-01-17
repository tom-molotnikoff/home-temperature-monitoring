package alerting

import (
	"fmt"
	"log"
)

type AlertRepository interface {
	GetAlertRuleBySensorID(sensorID int) (*AlertRule, error)
	UpdateLastAlertSent(ruleID int) error
	RecordAlertSent(ruleID, sensorID int, reason string, numericValue float64, statusValue string) error
}

// InAppNotificationCallback is called when threshold alerts should be sent as notifications
type InAppNotificationCallback func(sensorName, sensorType, reason string, numericValue float64)

type AlertService struct {
	repo                AlertRepository
	inAppNotifyCallback InAppNotificationCallback
}

func NewAlertService(repo AlertRepository) *AlertService {
	return &AlertService{
		repo: repo,
	}
}

func (s *AlertService) SetInAppNotificationCallback(cb InAppNotificationCallback) {
	s.inAppNotifyCallback = cb
}

func (s *AlertService) ProcessReadingAlert(sensorID int, sensorName, sensorType string, numericValue float64, statusValue string) error {
	rule, err := s.repo.GetAlertRuleBySensorID(sensorID)
	if err != nil {
		return fmt.Errorf("failed to get alert rule for sensor %d: %w", sensorID, err)
	}

	if rule == nil {
		log.Printf("No alert rule configured for sensor %s (ID: %d), skipping alert check", sensorName, sensorID)
		return nil
	}

	shouldAlert, reason := rule.ShouldAlert(numericValue, statusValue)
	if !shouldAlert {
		log.Printf("Alert conditions not met for sensor %s (ID: %d) with value %.2f", sensorName, sensorID, numericValue)
		return nil
	}

	if rule.IsRateLimited() {
		log.Printf("Alert for sensor %s (ID: %d) is rate limited, skipping", sensorName, sensorID)
		return nil
	}

	log.Printf("Triggering alert for sensor %s (ID: %d): %s (value: %.2f)", sensorName, sensorID, reason, numericValue)

	// Send in-app notification and email via callback (emails handled by NotificationService)
	if s.inAppNotifyCallback != nil {
		go s.inAppNotifyCallback(sensorName, sensorType, reason, numericValue)
	}

	err = s.repo.RecordAlertSent(rule.ID, sensorID, reason, numericValue, statusValue)
	if err != nil {
		log.Printf("Warning: failed to record alert sent for sensor %s: %v", sensorName, err)
	}

	log.Printf("Alert sent for sensor %s (ID: %d): %s", sensorName, sensorID, reason)
	return nil
}

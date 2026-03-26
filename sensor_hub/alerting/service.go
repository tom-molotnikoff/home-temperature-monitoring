package alerting

import (
	"context"
	"fmt"
	"log/slog"
)

type AlertRepository interface {
	GetAlertRuleBySensorID(ctx context.Context, sensorID int) (*AlertRule, error)
	UpdateLastAlertSent(ctx context.Context, ruleID int) error
	RecordAlertSent(ctx context.Context, ruleID, sensorID int, reason string, numericValue float64, statusValue string) error
}

// InAppNotificationCallback is called when threshold alerts should be sent as notifications
type InAppNotificationCallback func(sensorName, sensorType, reason string, numericValue float64)

type AlertService struct {
	repo                AlertRepository
	inAppNotifyCallback InAppNotificationCallback
	logger              *slog.Logger
}

func NewAlertService(repo AlertRepository, logger *slog.Logger) *AlertService {
	return &AlertService{
		repo:   repo,
		logger: logger.With("component", "alert_service"),
	}
}

func (s *AlertService) SetInAppNotificationCallback(cb InAppNotificationCallback) {
	s.inAppNotifyCallback = cb
}

func (s *AlertService) ProcessReadingAlert(ctx context.Context, sensorID int, sensorName, sensorType string, numericValue float64, statusValue string) error {
	rule, err := s.repo.GetAlertRuleBySensorID(ctx, sensorID)
	if err != nil {
		return fmt.Errorf("failed to get alert rule for sensor %d: %w", sensorID, err)
	}

	if rule == nil {
		s.logger.Debug("no alert rule configured, skipping", "sensor", sensorName, "sensor_id", sensorID)
		return nil
	}

	shouldAlert, reason := rule.ShouldAlert(numericValue, statusValue)
	if !shouldAlert {
		s.logger.Debug("alert conditions not met", "sensor", sensorName, "sensor_id", sensorID, "value", numericValue)
		return nil
	}

	if rule.IsRateLimited() {
		s.logger.Debug("alert rate limited, skipping", "sensor", sensorName, "sensor_id", sensorID)
		return nil
	}

	s.logger.Info("triggering alert", "sensor", sensorName, "sensor_id", sensorID, "reason", reason, "value", numericValue)

	// Send in-app notification and email via callback (emails handled by NotificationService)
	if s.inAppNotifyCallback != nil {
		go s.inAppNotifyCallback(sensorName, sensorType, reason, numericValue)
	}

	err = s.repo.RecordAlertSent(ctx, rule.ID, sensorID, reason, numericValue, statusValue)
	if err != nil {
		s.logger.Warn("failed to record alert sent", "sensor", sensorName, "error", err)
	}

	s.logger.Info("alert sent", "sensor", sensorName, "sensor_id", sensorID, "reason", reason)
	return nil
}

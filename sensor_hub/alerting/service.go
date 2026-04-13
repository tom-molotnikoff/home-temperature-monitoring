package alerting

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type AlertRepository interface {
	GetAlertRuleForReading(ctx context.Context, sensorID int, measurementTypeName string) (*AlertRule, error)
	UpdateLastAlertSent(ctx context.Context, ruleID int) error
	RecordAlertSent(ctx context.Context, ruleID, sensorID, measurementTypeId int, reason string, numericValue float64, statusValue string) error
}

// InAppNotificationCallback is called when threshold alerts should be sent as notifications
type InAppNotificationCallback func(sensorName, sensorType, reason string, numericValue float64)

type AlertService struct {
	repo                AlertRepository
	inAppNotifyCallback InAppNotificationCallback
	logger              *slog.Logger
	// In-memory rate limit tracking to prevent TOCTOU races when
	// multiple readings for the same sensor arrive concurrently.
	rateMu    sync.Mutex
	lastFired map[int]time.Time // rule ID → last fire time
}

func NewAlertService(repo AlertRepository, logger *slog.Logger) *AlertService {
	return &AlertService{
		repo:      repo,
		logger:    logger.With("component", "alert_service"),
		lastFired: make(map[int]time.Time),
	}
}

func (s *AlertService) SetInAppNotificationCallback(cb InAppNotificationCallback) {
	s.inAppNotifyCallback = cb
}

func (s *AlertService) ProcessReadingAlert(ctx context.Context, sensorID int, sensorName, measurementType string, numericValue float64, statusValue string) error {
	rule, err := s.repo.GetAlertRuleForReading(ctx, sensorID, measurementType)
	if err != nil {
		return fmt.Errorf("failed to get alert rule for sensor %d measurement %s: %w", sensorID, measurementType, err)
	}

	if rule == nil {
		s.logger.Debug("no alert rule configured, skipping", "sensor", sensorName, "sensor_id", sensorID, "measurement_type", measurementType)
		return nil
	}

	shouldAlert, reason := rule.ShouldAlert(numericValue, statusValue)
	if !shouldAlert {
		s.logger.Debug("alert conditions not met", "sensor", sensorName, "sensor_id", sensorID, "value", numericValue)
		return nil
	}

	// Check both DB-based and in-memory rate limits under a lock to
	// prevent concurrent goroutines from all passing the check.
	s.rateMu.Lock()
	if rule.IsRateLimited() {
		s.rateMu.Unlock()
		s.logger.Debug("alert rate limited (DB)", "sensor", sensorName, "sensor_id", sensorID, "rule_id", rule.ID)
		return nil
	}
	if last, ok := s.lastFired[rule.ID]; ok && rule.RateLimitSeconds > 0 {
		if time.Since(last) < time.Duration(rule.RateLimitSeconds)*time.Second {
			s.rateMu.Unlock()
			s.logger.Debug("alert rate limited (in-memory)", "sensor", sensorName, "sensor_id", sensorID, "rule_id", rule.ID)
			return nil
		}
	}
	s.lastFired[rule.ID] = time.Now()
	s.rateMu.Unlock()

	s.logger.Info("triggering alert", "sensor", sensorName, "sensor_id", sensorID, "reason", reason, "value", numericValue)

	// Send in-app notification and email via callback (emails handled by NotificationService)
	if s.inAppNotifyCallback != nil {
		go s.inAppNotifyCallback(sensorName, measurementType, reason, numericValue)
	}

	err = s.repo.RecordAlertSent(ctx, rule.ID, sensorID, rule.MeasurementTypeId, reason, numericValue, statusValue)
	if err != nil {
		s.logger.Warn("failed to record alert sent", "sensor", sensorName, "error", err)
	}

	s.logger.Info("alert sent", "sensor", sensorName, "sensor_id", sensorID, "reason", reason)
	return nil
}

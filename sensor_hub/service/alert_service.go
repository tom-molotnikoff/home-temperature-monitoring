package service

import (
	"context"

	"example/sensorHub/alerting"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"log/slog"
)

type AlertManagementService struct {
	alertRepo database.AlertRepository
	logger    *slog.Logger
}

func NewAlertManagementService(alertRepo database.AlertRepository, logger *slog.Logger) AlertManagementServiceInterface {
	return &AlertManagementService{
		alertRepo: alertRepo,
		logger:    logger.With("component", "alert_management_service"),
	}
}

func (s *AlertManagementService) ServiceGetAllAlertRules(ctx context.Context) ([]alerting.AlertRule, error) {
	return s.alertRepo.GetAllAlertRules(ctx)
}

func (s *AlertManagementService) ServiceGetAlertRuleBySensorID(ctx context.Context, sensorID int) (*alerting.AlertRule, error) {
	return s.alertRepo.GetAlertRuleBySensorID(ctx, sensorID)
}

func (s *AlertManagementService) ServiceCreateAlertRule(ctx context.Context, rule *alerting.AlertRule) error {
	return s.alertRepo.CreateAlertRule(ctx, rule)
}

func (s *AlertManagementService) ServiceUpdateAlertRule(ctx context.Context, rule *alerting.AlertRule) error {
	return s.alertRepo.UpdateAlertRule(ctx, rule)
}

func (s *AlertManagementService) ServiceDeleteAlertRule(ctx context.Context, sensorID int) error {
	return s.alertRepo.DeleteAlertRule(ctx, sensorID)
}

func (s *AlertManagementService) ServiceGetAlertHistory(ctx context.Context, sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	return s.alertRepo.GetAlertHistory(ctx, sensorID, limit)
}

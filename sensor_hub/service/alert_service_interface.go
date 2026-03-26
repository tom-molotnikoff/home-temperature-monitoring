package service

import (
	"context"
	"example/sensorHub/alerting"
	"example/sensorHub/types"
)

type AlertManagementServiceInterface interface {
	ServiceGetAllAlertRules(ctx context.Context) ([]alerting.AlertRule, error)
	ServiceGetAlertRuleBySensorID(ctx context.Context, sensorID int) (*alerting.AlertRule, error)
	ServiceCreateAlertRule(ctx context.Context, rule *alerting.AlertRule) error
	ServiceUpdateAlertRule(ctx context.Context, rule *alerting.AlertRule) error
	ServiceDeleteAlertRule(ctx context.Context, sensorID int) error
	ServiceGetAlertHistory(ctx context.Context, sensorID int, limit int) ([]types.AlertHistoryEntry, error)
}

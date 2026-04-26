package service

import (
	"context"
	"example/sensorHub/alerting"
	gen "example/sensorHub/gen"
)

type AlertManagementServiceInterface interface {
	ServiceGetAllAlertRules(ctx context.Context) ([]alerting.AlertRule, error)
	ServiceGetAlertRuleByID(ctx context.Context, ruleID int) (*alerting.AlertRule, error)
	ServiceGetAlertRuleBySensorID(ctx context.Context, sensorID int) (*alerting.AlertRule, error)
	ServiceGetAlertRulesBySensorID(ctx context.Context, sensorID int) ([]alerting.AlertRule, error)
	ServiceCreateAlertRule(ctx context.Context, rule *alerting.AlertRule) error
	ServiceUpdateAlertRule(ctx context.Context, rule *alerting.AlertRule) error
	ServiceDeleteAlertRule(ctx context.Context, ruleID int) error
	ServiceGetAlertHistory(ctx context.Context, sensorID int, limit int) ([]gen.AlertHistoryEntry, error)
}

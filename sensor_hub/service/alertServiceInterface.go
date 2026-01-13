package service

import (
	"example/sensorHub/alerting"
	"example/sensorHub/types"
)

type AlertManagementServiceInterface interface {
	ServiceGetAllAlertRules() ([]alerting.AlertRule, error)
	ServiceGetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error)
	ServiceCreateAlertRule(rule *alerting.AlertRule) error
	ServiceUpdateAlertRule(rule *alerting.AlertRule) error
	ServiceDeleteAlertRule(sensorID int) error
	ServiceGetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error)
}

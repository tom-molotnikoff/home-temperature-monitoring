package service

import (
	"example/sensorHub/alerting"
	database "example/sensorHub/db"
	"example/sensorHub/types"
)

type AlertManagementService struct {
	alertRepo database.AlertRepository
}

func NewAlertManagementService(alertRepo database.AlertRepository) AlertManagementServiceInterface {
	return &AlertManagementService{
		alertRepo: alertRepo,
	}
}

func (s *AlertManagementService) ServiceGetAllAlertRules() ([]alerting.AlertRule, error) {
	return s.alertRepo.GetAllAlertRules()
}

func (s *AlertManagementService) ServiceGetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error) {
	return s.alertRepo.GetAlertRuleBySensorID(sensorID)
}

func (s *AlertManagementService) ServiceCreateAlertRule(rule *alerting.AlertRule) error {
	return s.alertRepo.CreateAlertRule(rule)
}

func (s *AlertManagementService) ServiceUpdateAlertRule(rule *alerting.AlertRule) error {
	return s.alertRepo.UpdateAlertRule(rule)
}

func (s *AlertManagementService) ServiceDeleteAlertRule(sensorID int) error {
	return s.alertRepo.DeleteAlertRule(sensorID)
}

func (s *AlertManagementService) ServiceGetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	return s.alertRepo.GetAlertHistory(sensorID, limit)
}

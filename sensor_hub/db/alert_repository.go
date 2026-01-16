package database

import (
	"database/sql"
	"example/sensorHub/alerting"
	"example/sensorHub/types"
	"fmt"
	"time"
)

type AlertRepository interface {
	GetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error)
	UpdateLastAlertSent(ruleID int) error
	RecordAlertSent(ruleID, sensorID int, reason string, numericValue float64, statusValue string) error
	GetAllAlertRules() ([]alerting.AlertRule, error)
	GetAlertRuleBySensorName(sensorName string) (*alerting.AlertRule, error)
	CreateAlertRule(rule *alerting.AlertRule) error
	UpdateAlertRule(rule *alerting.AlertRule) error
	DeleteAlertRule(sensorID int) error
	GetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error)
}

type AlertRepositoryImpl struct {
	db *sql.DB
}

func NewAlertRepository(db *sql.DB) AlertRepository {
	return &AlertRepositoryImpl{db: db}
}

func (r *AlertRepositoryImpl) GetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error) {
	query := `
		SELECT 
			ar.id,
			ar.sensor_id,
			s.name,
			ar.alert_type,
			ar.high_threshold,
			ar.low_threshold,
			ar.trigger_status,
			ar.enabled,
			ar.rate_limit_hours,
			ah.sent_at
		FROM sensor_alert_rules ar
		JOIN sensors s ON ar.sensor_id = s.id
		LEFT JOIN (
			SELECT alert_rule_id, MAX(sent_at) as sent_at
			FROM alert_sent_history
			GROUP BY alert_rule_id
		) ah ON ar.id = ah.alert_rule_id
		WHERE ar.sensor_id = ? AND ar.enabled = TRUE
		LIMIT 1
	`

	var rule alerting.AlertRule
	var lastAlertSent sql.NullTime
	var triggerStatus sql.NullString

	err := r.db.QueryRow(query, sensorID).Scan(
		&rule.ID,
		&rule.SensorID,
		&rule.SensorName,
		&rule.AlertType,
		&rule.HighThreshold,
		&rule.LowThreshold,
		&triggerStatus,
		&rule.Enabled,
		&rule.RateLimitHours,
		&lastAlertSent,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule for sensor %d: %w", sensorID, err)
	}

	if lastAlertSent.Valid {
		rule.LastAlertSentAt = &lastAlertSent.Time
	}

	if triggerStatus.Valid {
		rule.TriggerStatus = triggerStatus.String
	}

	return &rule, nil
}

func (r *AlertRepositoryImpl) UpdateLastAlertSent(ruleID int) error {
	// This is handled by RecordAlertSent, kept for backwards compatibility
	return nil
}

func (r *AlertRepositoryImpl) RecordAlertSent(ruleID, sensorID int, reason string, numericValue float64, statusValue string) error {
	query := `
		INSERT INTO alert_sent_history 
		(alert_rule_id, sensor_id, alert_reason, reading_value, reading_status)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, ruleID, sensorID, reason, numericValue, statusValue)
	if err != nil {
		return fmt.Errorf("failed to record alert sent: %w", err)
	}

	return nil
}

func (r *AlertRepositoryImpl) GetAllAlertRules() ([]alerting.AlertRule, error) {
	query := `
		SELECT 
			sar.sensor_id,
			s.name,
			sar.alert_type,
			sar.high_threshold,
			sar.low_threshold,
			sar.trigger_status,
			sar.enabled,
			sar.rate_limit_hours,
			ash.sent_at
		FROM sensor_alert_rules sar
		INNER JOIN sensors s ON sar.sensor_id = s.id
		LEFT JOIN (
			SELECT sensor_id, MAX(sent_at) as sent_at
			FROM alert_sent_history
			GROUP BY sensor_id
		) ash ON sar.sensor_id = ash.sensor_id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all alert rules: %w", err)
	}
	defer rows.Close()

	var rules []alerting.AlertRule
	for rows.Next() {
		var rule alerting.AlertRule
		var lastAlertSentAt *time.Time
		var triggerStatus sql.NullString
		err := rows.Scan(
			&rule.SensorID,
			&rule.SensorName,
			&rule.AlertType,
			&rule.HighThreshold,
			&rule.LowThreshold,
			&triggerStatus,
			&rule.Enabled,
			&rule.RateLimitHours,
			&lastAlertSentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}
		if triggerStatus.Valid {
			rule.TriggerStatus = triggerStatus.String
		}
		rule.LastAlertSentAt = lastAlertSentAt
		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *AlertRepositoryImpl) GetAlertRuleBySensorName(sensorName string) (*alerting.AlertRule, error) {
	query := `
		SELECT 
			sar.sensor_id,
			s.name,
			sar.alert_type,
			sar.high_threshold,
			sar.low_threshold,
			sar.trigger_status,
			sar.enabled,
			sar.rate_limit_hours,
			ash.sent_at
		FROM sensor_alert_rules sar
		INNER JOIN sensors s ON sar.sensor_id = s.id
		LEFT JOIN (
			SELECT sensor_id, MAX(sent_at) as sent_at
			FROM alert_sent_history
			WHERE sensor_id = (SELECT id FROM sensors WHERE name = ?)
			GROUP BY sensor_id
		) ash ON sar.sensor_id = ash.sensor_id
		WHERE s.name = ?
	`

	var rule alerting.AlertRule
	var lastAlertSentAt *time.Time
	var triggerStatus sql.NullString
	err := r.db.QueryRow(query, sensorName, sensorName).Scan(
		&rule.SensorID,
		&rule.SensorName,
		&rule.AlertType,
		&rule.HighThreshold,
		&rule.LowThreshold,
		&triggerStatus,
		&rule.Enabled,
		&rule.RateLimitHours,
		&lastAlertSentAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert rule not found for sensor: %s", sensorName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule by sensor name: %w", err)
	}

	if triggerStatus.Valid {
		rule.TriggerStatus = triggerStatus.String
	}
	rule.LastAlertSentAt = lastAlertSentAt
	return &rule, nil
}

func (r *AlertRepositoryImpl) CreateAlertRule(rule *alerting.AlertRule) error {
	query := `
		INSERT INTO sensor_alert_rules 
		(sensor_id, alert_type, high_threshold, low_threshold, trigger_status, rate_limit_hours, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		rule.SensorID,
		rule.AlertType,
		rule.HighThreshold,
		rule.LowThreshold,
		rule.TriggerStatus,
		rule.RateLimitHours,
		rule.Enabled,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	return nil
}

func (r *AlertRepositoryImpl) UpdateAlertRule(rule *alerting.AlertRule) error {
	query := `
		UPDATE sensor_alert_rules
		SET alert_type = ?,
			high_threshold = ?,
			low_threshold = ?,
			trigger_status = ?,
			rate_limit_hours = ?,
			enabled = ?
		WHERE sensor_id = ?
	`

	_, err := r.db.Exec(query,
		rule.AlertType,
		rule.HighThreshold,
		rule.LowThreshold,
		rule.TriggerStatus,
		rule.RateLimitHours,
		rule.Enabled,
		rule.SensorID,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	return nil
}

func (r *AlertRepositoryImpl) DeleteAlertRule(sensorID int) error {
	query := `DELETE FROM sensor_alert_rules WHERE sensor_id = ?`
	_, err := r.db.Exec(query, sensorID)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}
	return nil
}

func (r *AlertRepositoryImpl) GetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	query := `
		SELECT 
			ash.id, 
			ash.sensor_id, 
			sar.alert_type, 
			ash.reading_value, 
			ash.sent_at
		FROM alert_sent_history ash
		JOIN sensor_alert_rules sar ON ash.alert_rule_id = sar.id
		WHERE ash.sensor_id = ?
		ORDER BY ash.sent_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, sensorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert history: %w", err)
	}
	defer rows.Close()

	var history []types.AlertHistoryEntry
	for rows.Next() {
		var entry types.AlertHistoryEntry
		var readingValue sql.NullFloat64
		err := rows.Scan(&entry.ID, &entry.SensorID, &entry.AlertType, &readingValue, &entry.SentAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert history entry: %w", err)
		}
		if readingValue.Valid {
			entry.ReadingValue = fmt.Sprintf("%.2f", readingValue.Float64)
		}
		history = append(history, entry)
	}

	return history, nil
}

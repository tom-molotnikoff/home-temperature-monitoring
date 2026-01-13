package alerting

import (
	"fmt"
	"time"
)

type AlertType string

const (
	AlertTypeNumericRange AlertType = "numeric_range"
	AlertTypeStatusBased  AlertType = "status_based"
)

type AlertRule struct {
	ID              int
	SensorID        int
	SensorName      string
	AlertType       AlertType
	HighThreshold   float64
	LowThreshold    float64
	TriggerStatus   string
	Enabled         bool
	RateLimitHours  int
	LastAlertSentAt *time.Time
}

func (r *AlertRule) Validate() error {
	if r.AlertType == AlertTypeNumericRange {
		if r.HighThreshold <= r.LowThreshold {
			return fmt.Errorf("high threshold must be greater than low threshold")
		}
	}
	if r.AlertType == AlertTypeStatusBased {
		if r.TriggerStatus == "" {
			return fmt.Errorf("trigger status must be set for status-based alerts")
		}
	}
	return nil
}

func (r *AlertRule) ShouldAlert(numericValue float64, statusValue string) (bool, string) {
	if !r.Enabled {
		return false, ""
	}

	switch r.AlertType {
	case AlertTypeNumericRange:
		if numericValue > r.HighThreshold {
			return true, fmt.Sprintf("value %.2f is above high threshold %.2f", numericValue, r.HighThreshold)
		}
		if numericValue < r.LowThreshold {
			return true, fmt.Sprintf("value %.2f is below low threshold %.2f", numericValue, r.LowThreshold)
		}
		return false, ""

	case AlertTypeStatusBased:
		if statusValue == r.TriggerStatus {
			return true, fmt.Sprintf("status is %s", statusValue)
		}
		return false, ""

	default:
		return false, ""
	}
}

func (r *AlertRule) IsRateLimited() bool {
	if r.RateLimitHours == 0 {
		return false
	}
	if r.LastAlertSentAt == nil {
		return false
	}
	elapsed := time.Since(*r.LastAlertSentAt)
	return elapsed < time.Duration(r.RateLimitHours)*time.Hour
}

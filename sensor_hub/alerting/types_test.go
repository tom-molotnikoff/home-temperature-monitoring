package alerting

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlertRule_ValidateNumericRange(t *testing.T) {
	rule := AlertRule{
		SensorID:       1,
		SensorName:     "TestSensor",
		AlertType:      AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		Enabled:        true,
		RateLimitHours: 1,
	}

	err := rule.Validate()
	assert.NoError(t, err)
}

func TestAlertRule_ValidateNumericRange_InvalidThresholds(t *testing.T) {
	rule := AlertRule{
		SensorID:       1,
		AlertType:      AlertTypeNumericRange,
		HighThreshold:  10.0,
		LowThreshold:   30.0,
		Enabled:        true,
	}

	err := rule.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "high threshold must be greater than low threshold")
}

func TestAlertRule_ValidateStatusBased(t *testing.T) {
	rule := AlertRule{
		SensorID:       2,
		SensorName:     "DoorSensor",
		AlertType:      AlertTypeStatusBased,
		TriggerStatus:  "open",
		Enabled:        true,
		RateLimitHours: 0,
	}

	err := rule.Validate()
	assert.NoError(t, err)
}

func TestAlertRule_ShouldAlert_NumericHigh(t *testing.T) {
	rule := AlertRule{
		AlertType:     AlertTypeNumericRange,
		HighThreshold: 30.0,
		LowThreshold:  10.0,
		Enabled:       true,
	}

	shouldAlert, reason := rule.ShouldAlert(35.0, "")
	assert.True(t, shouldAlert)
	assert.Contains(t, reason, "above high threshold")
}

func TestAlertRule_ShouldAlert_NumericLow(t *testing.T) {
	rule := AlertRule{
		AlertType:     AlertTypeNumericRange,
		HighThreshold: 30.0,
		LowThreshold:  10.0,
		Enabled:       true,
	}

	shouldAlert, reason := rule.ShouldAlert(5.0, "")
	assert.True(t, shouldAlert)
	assert.Contains(t, reason, "below low threshold")
}

func TestAlertRule_ShouldAlert_NumericInRange(t *testing.T) {
	rule := AlertRule{
		AlertType:     AlertTypeNumericRange,
		HighThreshold: 30.0,
		LowThreshold:  10.0,
		Enabled:       true,
	}

	shouldAlert, _ := rule.ShouldAlert(20.0, "")
	assert.False(t, shouldAlert)
}

func TestAlertRule_ShouldAlert_StatusMatch(t *testing.T) {
	rule := AlertRule{
		AlertType:     AlertTypeStatusBased,
		TriggerStatus: "open",
		Enabled:       true,
	}

	shouldAlert, reason := rule.ShouldAlert(0, "open")
	assert.True(t, shouldAlert)
	assert.Contains(t, reason, "status is")
}

func TestAlertRule_ShouldAlert_StatusNoMatch(t *testing.T) {
	rule := AlertRule{
		AlertType:     AlertTypeStatusBased,
		TriggerStatus: "open",
		Enabled:       true,
	}

	shouldAlert, _ := rule.ShouldAlert(0, "closed")
	assert.False(t, shouldAlert)
}

func TestAlertRule_IsRateLimited_NoLimit(t *testing.T) {
	rule := AlertRule{
		RateLimitHours: 0,
	}

	assert.False(t, rule.IsRateLimited())
}

func TestAlertRule_IsRateLimited_NeverSent(t *testing.T) {
	rule := AlertRule{
		RateLimitHours:  1,
		LastAlertSentAt: nil,
	}

	assert.False(t, rule.IsRateLimited())
}

func TestAlertRule_IsRateLimited_RecentlySent(t *testing.T) {
	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)

	rule := AlertRule{
		RateLimitHours:  1,
		LastAlertSentAt: &thirtyMinutesAgo,
	}

	assert.True(t, rule.IsRateLimited())
}

func TestAlertRule_IsRateLimited_OldEnough(t *testing.T) {
	twoHoursAgo := time.Now().Add(-2 * time.Hour)

	rule := AlertRule{
		RateLimitHours:  1,
		LastAlertSentAt: &twoHoursAgo,
	}

	assert.False(t, rule.IsRateLimited())
}

# Generic Sensor Alerting System Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Refactor the temperature-specific SMTP alerting system into a generic, database-driven alerting system that supports multiple sensor types (numeric ranges and status-based) with configurable rate limiting.

**Architecture:** Replace hardcoded temperature thresholds in application config with per-sensor alert rules stored in the database. Create a generic alerting interface that can handle both numeric range checks (temperature, humidity, pressure) and status-based alerts (door open/closed, motion detected). Implement rate limiting to prevent alert spam (configurable per sensor, default 1 hour). Use dependency injection to decouple SMTP sending from alert logic.

**Tech Stack:** Go 1.25, MySQL (Flyway migrations), net/smtp, OAuth2, testify

---

## Current State Analysis

**Problems with current implementation:**
1. Hardcoded temperature thresholds in `ApplicationConfiguration` (lines 9-10)
2. Temperature-specific function `SendTemperatureAlertEmailIfNeeded` in `smtp/smtp.go`
3. No rate limiting - sends email on every collection cycle if threshold breached
4. No database-driven configuration
5. SMTP package tightly coupled to `oauth` and `appProps` global state
6. No abstraction for different sensor types

**Files to be modified:**
- `sensor_hub/smtp/smtp.go` - Complete rewrite
- `sensor_hub/smtp/smtp_test.go` - Complete rewrite  
- `sensor_hub/service/sensorService.go:209` - Change integration point
- `sensor_hub/db/changesets/V14__add_sensor_alert_config.sql` - New migration
- `sensor_hub/types/types.go` - Add alert types
- `sensor_hub/application_properties/applicationConfiguration.go` - Remove temp thresholds

---

## Task 1: Create Alert Domain Types (TDD)

**Files:**
- Create: `sensor_hub/alerting/types.go`
- Create: `sensor_hub/alerting/types_test.go`

**Step 1: Write test for AlertRule validation**

Create `sensor_hub/alerting/types_test.go`:

```go
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
	}

	shouldAlert, _ := rule.ShouldAlert(20.0, "")
	assert.False(t, shouldAlert)
}

func TestAlertRule_ShouldAlert_StatusMatch(t *testing.T) {
	rule := AlertRule{
		AlertType:     AlertTypeStatusBased,
		TriggerStatus: "open",
	}

	shouldAlert, reason := rule.ShouldAlert(0, "open")
	assert.True(t, shouldAlert)
	assert.Contains(t, reason, "status is")
}

func TestAlertRule_ShouldAlert_StatusNoMatch(t *testing.T) {
	rule := AlertRule{
		AlertType:     AlertTypeStatusBased,
		TriggerStatus: "open",
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
	now := time.Now()
	thirtyMinutesAgo := now.Add(-30 * time.Minute)

	rule := AlertRule{
		RateLimitHours:  1,
		LastAlertSentAt: &thirtyMinutesAgo,
	}

	assert.True(t, rule.IsRateLimited())
}

func TestAlertRule_IsRateLimited_OldEnough(t *testing.T) {
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)

	rule := AlertRule{
		RateLimitHours:  1,
		LastAlertSentAt: &twoHoursAgo,
	}

	assert.False(t, rule.IsRateLimited())
}
```

**Step 2: Run test to verify it fails**

```bash
cd /home/tommolotnikoff/Documents/repositories/home-temperature-monitoring/sensor_hub
go test ./alerting/ -v
```

Expected: FAIL - package alerting does not exist

**Step 3: Create minimal types to make tests pass**

Create `sensor_hub/alerting/types.go`:

```go
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
```

**Step 4: Run tests to verify they pass**

```bash
cd /home/tommolotnikoff/Documents/repositories/home-temperature-monitoring/sensor_hub
go test ./alerting/ -v
```

Expected: PASS - all tests pass

**Step 5: Commit**

```bash
git add sensor_hub/alerting/
git commit -m "feat(alerting): add domain types for generic sensor alerting with rate limiting"
```

---

## Next Steps

After Task 1 is complete, continue with:
- Task 2: Database Migration
- Task 3: Alert Repository
- Task 4: Alert Service
- Task 5: SMTP Notifier
- Task 6-11: Integration and cleanup

Plan complete and saved to `docs/plans/2026-01-13-generic-sensor-alerting.md`. Ready to execute using iterative-development skill.

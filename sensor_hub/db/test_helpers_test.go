package database

import (
	"database/sql"
	"testing"
	"time"

	"example/sensorHub/alerting"
	"example/sensorHub/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// newMockDB creates a new sqlmock database connection for testing.
// Returns the db, mock, and a cleanup function.
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// newMockDBWithQueryMatcher creates a mock DB with custom query matching.
func newMockDBWithQueryMatcher(t *testing.T, matcher sqlmock.QueryMatcher) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// Test data factories

func testSensor() types.Sensor {
	return types.Sensor{
		Id:           1,
		Name:         "test-sensor",
		Type:         "temperature",
		URL:          "http://localhost:8080",
		HealthStatus: types.SensorGoodHealth,
		HealthReason: "ok",
		Enabled:      true,
	}
}

func testSensorWithID(id int, name string) types.Sensor {
	return types.Sensor{
		Id:           id,
		Name:         name,
		Type:         "temperature",
		URL:          "http://localhost:8080",
		HealthStatus: types.SensorGoodHealth,
		HealthReason: "ok",
		Enabled:      true,
	}
}

func testUser() types.User {
	return types.User{
		Id:                 1,
		Username:           "testuser",
		Email:              "test@example.com",
		Disabled:           false,
		MustChangePassword: false,
		Roles:              []string{"user"},
		CreatedAt:          time.Now(),
	}
}

func testUserWithID(id int, username string) types.User {
	return types.User{
		Id:                 id,
		Username:           username,
		Email:              username + "@example.com",
		Disabled:           false,
		MustChangePassword: false,
		Roles:              []string{"user"},
		CreatedAt:          time.Now(),
	}
}

func testAlertRule() alerting.AlertRule {
	return alerting.AlertRule{
		ID:             1,
		SensorID:       1,
		SensorName:     "test-sensor",
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		TriggerStatus:  "",
		Enabled:        true,
		RateLimitHours: 1,
	}
}

func testTemperatureReading() types.TemperatureReading {
	return types.TemperatureReading{
		Id:          1,
		SensorName:  "test-sensor",
		Time:        "2026-01-16 12:00:00",
		Temperature: 22.5,
	}
}

func testSessionInfo() SessionInfo {
	now := time.Now()
	return SessionInfo{
		Id:             1,
		UserId:         1,
		CreatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour),
		LastAccessedAt: now,
		IpAddress:      "192.168.1.1",
		UserAgent:      "Mozilla/5.0",
	}
}

func testSensorHealthHistory() types.SensorHealthHistory {
	return types.SensorHealthHistory{
		Id:           1,
		SensorId:     "1",
		HealthStatus: types.SensorGoodHealth,
		RecordedAt:   time.Now(),
	}
}

// Column definitions for sqlmock rows

var sensorColumns = []string{"id", "name", "type", "url", "health_status", "health_reason", "enabled"}

var userColumns = []string{"id", "username", "email", "must_change_password", "disabled", "created_at", "updated_at"}

var userColumnsWithHash = []string{"id", "username", "email", "must_change_password", "disabled", "created_at", "updated_at", "password_hash"}

var sessionColumns = []string{"id", "user_id", "created_at", "expires_at", "last_accessed_at", "ip_address", "user_agent"}

var temperatureReadingColumns = []string{"id", "sensor_name", "time", "temperature"}

var sensorHealthHistoryColumns = []string{"id", "sensor_id", "health_status", "recorded_at"}

var alertRuleColumns = []string{"id", "sensor_id", "name", "alert_type", "high_threshold", "low_threshold", "trigger_status", "enabled", "rate_limit_hours", "sent_at"}

var alertRuleColumnsNoID = []string{"sensor_id", "name", "alert_type", "high_threshold", "low_threshold", "trigger_status", "enabled", "rate_limit_hours", "sent_at"}

var alertHistoryColumns = []string{"id", "sensor_id", "alert_type", "reading_value", "sent_at"}

var roleColumns = []string{"id", "name"}

var permissionColumns = []string{"id", "name", "description"}

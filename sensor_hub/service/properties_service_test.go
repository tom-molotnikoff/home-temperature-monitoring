package service

import (
	"testing"
	"time"

	appProps "example/sensorHub/application_properties"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Test helpers
// ============================================================================

func setupPropertiesServiceTestConfig() func() {
	// Save original config
	originalConfig := appProps.AppConfig

	// Set up minimal test config with actual field names
	appProps.AppConfig = &appProps.ApplicationConfiguration{
		SensorCollectionInterval:          30,
		AuthSessionTTLMinutes:             60,
		AuthBcryptCost:                    4,
		HealthHistoryRetentionDays:        30,
		SensorDataRetentionDays:           90,
		DataCleanupIntervalHours:          24,
		SMTPUser:                          "testuser",
		SMTPRecipient:                     "test@example.com",
		DatabaseHostname:                  "localhost",
		DatabasePort:                      "5432",
		DatabaseUsername:                  "testuser",
		DatabasePassword:                  "testpassword",
		EmailAlertHighTemperatureThreshold: 30.0,
		EmailAlertLowTemperatureThreshold:  15.0,
	}

	return func() {
		appProps.AppConfig = originalConfig
	}
}

// ============================================================================
// ServiceGetProperties tests
// ============================================================================

func TestPropertiesService_ServiceGetProperties_Success(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	result, err := service.ServiceGetProperties()

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check some expected keys exist
	assert.Contains(t, result, "sensor.collection.interval")
	assert.Contains(t, result, "auth.session.ttl.minutes")
}

func TestPropertiesService_ServiceGetProperties_MasksSensitiveData(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	result, err := service.ServiceGetProperties()

	assert.NoError(t, err)

	// Verify sensitive properties are masked
	for _, key := range appProps.SensitivePropertiesKeys {
		if val, ok := result[key]; ok {
			assert.Equal(t, "*****", val, "Sensitive key %s should be masked", key)
		}
	}
}

func TestPropertiesService_ServiceGetProperties_IncludesAllPropertyTypes(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	result, err := service.ServiceGetProperties()

	assert.NoError(t, err)

	// Should include app properties
	assert.Contains(t, result, "sensor.collection.interval")

	// Should include SMTP properties (if configured)
	assert.Contains(t, result, "smtp.user")

	// Should include database properties
	assert.Contains(t, result, "database.hostname")
}

// ============================================================================
// ServiceUpdateProperties tests
// ============================================================================

func TestPropertiesService_ServiceUpdateProperties_Success(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	properties := map[string]string{
		"sensor.collection.interval": "60",
		"auth.session.ttl.minutes":   "120",
	}

	err := service.ServiceUpdateProperties(properties)

	assert.NoError(t, err)

	// Verify values were updated
	assert.Equal(t, 60, appProps.AppConfig.SensorCollectionInterval)
	assert.Equal(t, 120, appProps.AppConfig.AuthSessionTTLMinutes)
	
	time.Sleep(50 * time.Millisecond)
}

func TestPropertiesService_ServiceUpdateProperties_SkipsMaskedSensitive(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	originalPassword := appProps.AppConfig.DatabasePassword

	service := NewPropertiesService()

	// When sensitive properties have "*****", they should be skipped
	properties := map[string]string{
		"database.password": "*****",
	}

	err := service.ServiceUpdateProperties(properties)

	assert.NoError(t, err)
	// Password should remain unchanged
	assert.Equal(t, originalPassword, appProps.AppConfig.DatabasePassword)
	
	time.Sleep(50 * time.Millisecond)
}

func TestPropertiesService_ServiceUpdateProperties_UpdatesSensitiveWhenChanged(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	// When sensitive properties have actual values, they should be updated
	properties := map[string]string{
		"database.password": "newSecretPassword",
	}

	err := service.ServiceUpdateProperties(properties)

	assert.NoError(t, err)
	assert.Equal(t, "newSecretPassword", appProps.AppConfig.DatabasePassword)
	
	time.Sleep(50 * time.Millisecond)
}

func TestPropertiesService_ServiceUpdateProperties_InvalidValue(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	// Invalid numeric value
	properties := map[string]string{
		"sensor.collection.interval": "not-a-number",
	}

	err := service.ServiceUpdateProperties(properties)

	// Should return error for invalid values
	assert.Error(t, err)
	
	time.Sleep(50 * time.Millisecond)
}

func TestPropertiesService_ServiceUpdateProperties_PartialUpdate(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	originalSessionTTL := appProps.AppConfig.AuthSessionTTLMinutes

	service := NewPropertiesService()

	// Only update one property
	properties := map[string]string{
		"sensor.collection.interval": "45",
	}

	err := service.ServiceUpdateProperties(properties)

	assert.NoError(t, err)
	assert.Equal(t, 45, appProps.AppConfig.SensorCollectionInterval)
	// Other properties should remain unchanged
	assert.Equal(t, originalSessionTTL, appProps.AppConfig.AuthSessionTTLMinutes)
	
	time.Sleep(50 * time.Millisecond)
}

func TestPropertiesService_ServiceUpdateProperties_EmptyMap(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	properties := map[string]string{}

	err := service.ServiceUpdateProperties(properties)

	assert.NoError(t, err)
	
	time.Sleep(50 * time.Millisecond)
}

func TestPropertiesService_ServiceUpdateProperties_UnknownKey(t *testing.T) {
	cleanup := setupPropertiesServiceTestConfig()
	defer cleanup()

	service := NewPropertiesService()

	// Unknown keys should be ignored
	properties := map[string]string{
		"unknownProperty":            "someValue",
		"sensor.collection.interval": "45",
	}

	err := service.ServiceUpdateProperties(properties)

	assert.NoError(t, err)
	// Known property should still be updated
	assert.Equal(t, 45, appProps.AppConfig.SensorCollectionInterval)
	
	// Give async goroutines time to complete
	time.Sleep(50 * time.Millisecond)
}

// ============================================================================
// NewPropertiesService tests
// ============================================================================

func TestNewPropertiesService_ReturnsService(t *testing.T) {
	service := NewPropertiesService()

	assert.NotNil(t, service)
}

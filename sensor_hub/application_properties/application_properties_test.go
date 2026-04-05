package appProps

import (
	"os"
	"path/filepath"
	"testing"

	"example/sensorHub/utils"

	"github.com/stretchr/testify/assert"
)

// validAppPropsMap returns a complete valid application properties map
func validAppPropsMap() map[string]string {
	return map[string]string{
		"sensor.collection.interval":             "300",
		"sensor.discovery.skip":                  "true",
		"openapi.yaml.location":                  "/path/to/openapi.yaml",
		"health.history.retention.days":          "180",
		"sensor.data.retention.days":             "365",
		"data.cleanup.interval.hours":            "24",
		"health.history.default.response.number": "1000",
		"failed.login.retention.days":            "2",
		"auth.bcrypt.cost":                       "12",
		"auth.session.ttl.minutes":               "43200",
		"auth.session.cookie.name":               "sensor_hub_session",
		"auth.login.backoff.window.minutes":      "15",
		"auth.login.backoff.threshold":           "5",
		"auth.login.backoff.base.seconds":        "2",
		"auth.login.backoff.max.seconds":         "300",
		"oauth.credentials.file.path":            "configuration/credentials.json",
		"oauth.token.file.path":                  "configuration/token.json",
		"oauth.token.refresh.interval.minutes":   "30",
	}
}

func validSmtpPropsMap() map[string]string {
	return map[string]string{
		"smtp.user": "user@example.com",
	}
}

func validDbPropsMap() map[string]string {
	return map[string]string{
		"database.path": "test/sensor_hub.db",
	}
}

func TestLoadConfigurationFromMaps_Success(t *testing.T) {
	appProps := validAppPropsMap()
	smtpProps := validSmtpPropsMap()
	dbProps := validDbPropsMap()

	cfg, err := LoadConfigurationFromMaps(appProps, smtpProps, dbProps)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 300, cfg.SensorCollectionInterval)
	assert.True(t, cfg.SensorDiscoverySkip)
	assert.Equal(t, "/path/to/openapi.yaml", cfg.OpenAPILocation)
	assert.Equal(t, 180, cfg.HealthHistoryRetentionDays)
	assert.Equal(t, 365, cfg.SensorDataRetentionDays)
	assert.Equal(t, 24, cfg.DataCleanupIntervalHours)
	assert.Equal(t, 1000, cfg.HealthHistoryDefaultResponseNumber)
	assert.Equal(t, 2, cfg.FailedLoginRetentionDays)
	assert.Equal(t, 12, cfg.AuthBcryptCost)
	assert.Equal(t, 43200, cfg.AuthSessionTTLMinutes)
	assert.Equal(t, "sensor_hub_session", cfg.AuthSessionCookieName)
	assert.Equal(t, 15, cfg.AuthLoginBackoffWindowMinutes)
	assert.Equal(t, 5, cfg.AuthLoginBackoffThreshold)
	assert.Equal(t, 2, cfg.AuthLoginBackoffBaseSeconds)
	assert.Equal(t, 300, cfg.AuthLoginBackoffMaxSeconds)
	assert.Equal(t, "user@example.com", cfg.SMTPUser)
	assert.Equal(t, "test/sensor_hub.db", cfg.DatabasePath)
}

func TestLoadConfigurationFromMaps_EmptyMaps(t *testing.T) {
	cfg, err := LoadConfigurationFromMaps(map[string]string{}, map[string]string{}, map[string]string{})

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 0, cfg.SensorCollectionInterval)
}

func TestLoadConfigurationFromMaps_InvalidSensorCollectionInterval(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["sensor.collection.interval"] = "abc"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidSensorDiscoverySkip(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["sensor.discovery.skip"] = "maybe"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidHealthHistoryRetentionDays(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["health.history.retention.days"] = "not-int"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidSensorDataRetentionDays(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["sensor.data.retention.days"] = "xxx"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidDataCleanupIntervalHours(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["data.cleanup.interval.hours"] = "bad"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidHealthHistoryDefaultResponseNumber(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["health.history.default.response.number"] = "nope"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidFailedLoginRetentionDays(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["failed.login.retention.days"] = "invalid"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidAuthBcryptCost(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["auth.bcrypt.cost"] = "high"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidAuthSessionTTLMinutes(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["auth.session.ttl.minutes"] = "forever"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidAuthLoginBackoffWindowMinutes(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["auth.login.backoff.window.minutes"] = "bad"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidAuthLoginBackoffThreshold(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["auth.login.backoff.threshold"] = "x"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidAuthLoginBackoffBaseSeconds(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["auth.login.backoff.base.seconds"] = "slow"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigurationFromMaps_InvalidAuthLoginBackoffMaxSeconds(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["auth.login.backoff.max.seconds"] = "max"

	cfg, err := LoadConfigurationFromMaps(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConvertConfigurationToMaps_Success(t *testing.T) {
	cfg := &ApplicationConfiguration{
		SensorCollectionInterval:           300,
		SensorDiscoverySkip:                true,
		OpenAPILocation:                    "/path/to/openapi.yaml",
		HealthHistoryRetentionDays:         180,
		SensorDataRetentionDays:            365,
		DataCleanupIntervalHours:           24,
		HealthHistoryDefaultResponseNumber: 1000,
		FailedLoginRetentionDays:           2,
		AuthBcryptCost:                     12,
		AuthSessionTTLMinutes:              43200,
		AuthSessionCookieName:              "my_session",
		AuthLoginBackoffWindowMinutes:      15,
		AuthLoginBackoffThreshold:          5,
		AuthLoginBackoffBaseSeconds:        2,
		AuthLoginBackoffMaxSeconds:         300,
		SMTPUser:                           "user@test.com",
		DatabasePath:                       "test/path.db",
	}

	appProps, smtpProps, dbProps := ConvertConfigurationToMaps(cfg)

	assert.Equal(t, "300", appProps["sensor.collection.interval"])
	assert.Equal(t, "true", appProps["sensor.discovery.skip"])
	assert.Equal(t, "/path/to/openapi.yaml", appProps["openapi.yaml.location"])
	assert.Equal(t, "180", appProps["health.history.retention.days"])
	assert.Equal(t, "365", appProps["sensor.data.retention.days"])
	assert.Equal(t, "24", appProps["data.cleanup.interval.hours"])
	assert.Equal(t, "1000", appProps["health.history.default.response.number"])
	assert.Equal(t, "2", appProps["failed.login.retention.days"])
	assert.Equal(t, "12", appProps["auth.bcrypt.cost"])
	assert.Equal(t, "43200", appProps["auth.session.ttl.minutes"])
	assert.Equal(t, "my_session", appProps["auth.session.cookie.name"])
	assert.Equal(t, "15", appProps["auth.login.backoff.window.minutes"])
	assert.Equal(t, "5", appProps["auth.login.backoff.threshold"])
	assert.Equal(t, "2", appProps["auth.login.backoff.base.seconds"])
	assert.Equal(t, "300", appProps["auth.login.backoff.max.seconds"])

	assert.Equal(t, "user@test.com", smtpProps["smtp.user"])

	assert.Equal(t, "test/path.db", dbProps["database.path"])
}

func TestConvertConfigurationToMaps_ZeroValues(t *testing.T) {
	cfg := &ApplicationConfiguration{}

	appProps, smtpProps, dbProps := ConvertConfigurationToMaps(cfg)

	assert.Equal(t, "0", appProps["sensor.collection.interval"])
	assert.Equal(t, "false", appProps["sensor.discovery.skip"])
	assert.Equal(t, "", smtpProps["smtp.user"])
	assert.Equal(t, "", dbProps["database.path"])
}

func TestConvertConfigurationToMaps_RoundTrip(t *testing.T) {
	original := &ApplicationConfiguration{
		SensorCollectionInterval:           600,
		SensorDiscoverySkip:                false,
		OpenAPILocation:                    "/api/spec.yaml",
		HealthHistoryRetentionDays:         90,
		SensorDataRetentionDays:            180,
		DataCleanupIntervalHours:           12,
		HealthHistoryDefaultResponseNumber: 1000,
		FailedLoginRetentionDays:           7,
		AuthBcryptCost:                     14,
		AuthSessionTTLMinutes:              60,
		AuthSessionCookieName:              "test_session",
		AuthLoginBackoffWindowMinutes:      30,
		AuthLoginBackoffThreshold:          10,
		AuthLoginBackoffBaseSeconds:        5,
		AuthLoginBackoffMaxSeconds:         600,
		SMTPUser:                           "smtp@test.com",
		DatabasePath:                       "test/roundtrip.db",
	}

	appProps, smtpProps, dbProps := ConvertConfigurationToMaps(original)
	restored, err := LoadConfigurationFromMaps(appProps, smtpProps, dbProps)

	assert.NoError(t, err)
	assert.Equal(t, original.SensorCollectionInterval, restored.SensorCollectionInterval)
	assert.Equal(t, original.SensorDiscoverySkip, restored.SensorDiscoverySkip)
	assert.Equal(t, original.AuthBcryptCost, restored.AuthBcryptCost)
	assert.Equal(t, original.SMTPUser, restored.SMTPUser)
	assert.Equal(t, original.DatabasePath, restored.DatabasePath)
}

// ============================================================================
// Validation function tests
// ============================================================================

func TestValidateApplicationProperties_ValidConfig(t *testing.T) {
	applicationProperties = validAppPropsMap()

	err := validateApplicationProperties()

	assert.NoError(t, err)
}

func TestValidateApplicationProperties_InvalidSensorCollectionInterval(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.collection.interval"] = "zero"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sensor collection interval")
}

func TestValidateApplicationProperties_ZeroSensorCollectionInterval(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.collection.interval"] = "0"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sensor collection interval")
}

func TestValidateApplicationProperties_NegativeSensorCollectionInterval(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.collection.interval"] = "-5"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sensor collection interval")
}

func TestValidateApplicationProperties_InvalidSensorDiscoverySkip(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.discovery.skip"] = "yes"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sensor discovery skip")
}

func TestValidateApplicationProperties_EmptyOpenAPIWhenDiscoveryEnabled(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.discovery.skip"] = "false"
	applicationProperties["openapi.yaml.location"] = ""

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openapi.yaml.location cannot be empty")
}

func TestValidateApplicationProperties_EmptyOpenAPIWhenDiscoverySkipped(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.discovery.skip"] = "true"
	applicationProperties["openapi.yaml.location"] = ""

	err := validateApplicationProperties()

	assert.NoError(t, err)
}

func TestValidateApplicationProperties_InvalidSensorDataRetentionDays(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.data.retention.days"] = "abc"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sensor data retention days")
}

func TestValidateApplicationProperties_NegativeSensorDataRetentionDays(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["sensor.data.retention.days"] = "-1"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sensor data retention days")
}

func TestValidateApplicationProperties_InvalidHealthHistoryRetentionDays(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["health.history.retention.days"] = "bad"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid health history retention days")
}

func TestValidateApplicationProperties_InvalidDataCleanupIntervalHours(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["data.cleanup.interval.hours"] = "nope"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data cleanup interval hours")
}

func TestValidateApplicationProperties_ZeroDataCleanupIntervalHours(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["data.cleanup.interval.hours"] = "0"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid data cleanup interval hours")
}

func TestValidateApplicationProperties_InvalidHealthHistoryDefaultResponseNumber(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["health.history.default.response.number"] = "x"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid health history default response number")
}

func TestValidateApplicationProperties_ZeroHealthHistoryDefaultResponseNumber(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["health.history.default.response.number"] = "0"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid health history default response number")
}

func TestValidateApplicationProperties_InvalidFailedLoginRetentionDays(t *testing.T) {
	applicationProperties = validAppPropsMap()
	applicationProperties["failed.login.retention.days"] = "bad"

	err := validateApplicationProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid failed login retention days")
}

// Tests for validateSMTPProperties

func TestValidateSMTPProperties_ValidConfig(t *testing.T) {
	smtpProperties = validSmtpPropsMap()

	err := validateSMTPProperties()

	assert.NoError(t, err)
}

func TestValidateSMTPProperties_EmptyValues(t *testing.T) {
	smtpProperties = map[string]string{
		"smtp.user": "",
	}

	err := validateSMTPProperties()

	assert.NoError(t, err)
}

// Tests for dbValidateDatabaseProperties

func TestDbValidateDatabaseProperties_ValidConfig(t *testing.T) {
	databaseProperties = validDbPropsMap()

	err := dbValidateDatabaseProperties()

	assert.NoError(t, err)
}

func TestDbValidateDatabaseProperties_MissingPath(t *testing.T) {
	databaseProperties = validDbPropsMap()
	databaseProperties["database.path"] = ""

	err := dbValidateDatabaseProperties()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database.path is not set")
}

// ============================================================================
// File reading tests (with mocked utils.ReadPropertiesFile)
// ============================================================================

func TestReadApplicationPropertiesFile_Success(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return map[string]string{
			"sensor.collection.interval": "600",
		}, nil
	}

	props, err := ReadApplicationPropertiesFile()

	assert.NoError(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "600", props["sensor.collection.interval"])
}

func TestReadApplicationPropertiesFile_FileReadError(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return nil, os.ErrNotExist
	}

	props, err := ReadApplicationPropertiesFile()

	assert.Error(t, err)
	assert.Nil(t, props)
	assert.Contains(t, err.Error(), "failed to read application properties file")
}

func TestReadApplicationPropertiesFile_ValidationError(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return map[string]string{
			"sensor.collection.interval": "-999",
		}, nil
	}

	props, err := ReadApplicationPropertiesFile()

	assert.Error(t, err)
	assert.Nil(t, props)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestReadDatabasePropertiesFile_Success(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return map[string]string{
			"database.path": "data/my_hub.db",
		}, nil
	}

	props, err := ReadDatabasePropertiesFile()

	assert.NoError(t, err)
	assert.Equal(t, "data/my_hub.db", props["database.path"])
}

func TestReadDatabasePropertiesFile_FileReadError(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return nil, os.ErrNotExist
	}

	props, err := ReadDatabasePropertiesFile()

	assert.Error(t, err)
	assert.Nil(t, props)
	assert.Contains(t, err.Error(), "failed to read database properties file")
}

func TestReadDatabasePropertiesFile_ValidationError(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return map[string]string{
			"database.path": "",
		}, nil
	}

	props, err := ReadDatabasePropertiesFile()

	assert.Error(t, err)
	assert.Nil(t, props)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestReadSMTPPropertiesFile_Success(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return map[string]string{
			"smtp.user": "sender@test.com",
		}, nil
	}

	props, err := ReadSMTPPropertiesFile()

	assert.NoError(t, err)
	assert.Equal(t, "sender@test.com", props["smtp.user"])
}

func TestReadSMTPPropertiesFile_FileReadError(t *testing.T) {
	originalReadPropertiesFile := utils.ReadPropertiesFile
	defer func() { utils.ReadPropertiesFile = originalReadPropertiesFile }()

	utils.ReadPropertiesFile = func(path string) (map[string]string, error) {
		return nil, os.ErrNotExist
	}

	props, err := ReadSMTPPropertiesFile()

	assert.Error(t, err)
	assert.Nil(t, props)
	assert.Contains(t, err.Error(), "failed to read SMTP properties file")
}

// ============================================================================
// SaveConfigurationToFiles tests (with temp directory)
// ============================================================================

func TestSaveConfigurationToFiles_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "app-props-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	origAppPath := applicationPropertiesFilePath
	origSmtpPath := smtpPropertiesFilePath
	origDbPath := databasePropertiesFilePath
	defer func() {
		applicationPropertiesFilePath = origAppPath
		smtpPropertiesFilePath = origSmtpPath
		databasePropertiesFilePath = origDbPath
	}()

	applicationPropertiesFilePath = filepath.Join(tempDir, "application.properties")
	smtpPropertiesFilePath = filepath.Join(tempDir, "smtp.properties")
	databasePropertiesFilePath = filepath.Join(tempDir, "database.properties")

	origConfig := AppConfig
	defer func() { AppConfig = origConfig }()

	AppConfig = &ApplicationConfiguration{
		SensorCollectionInterval:           120,
		SensorDiscoverySkip:                false,
		OpenAPILocation:                    "/test/openapi.yaml",
		HealthHistoryRetentionDays:         90,
		SensorDataRetentionDays:            180,
		DataCleanupIntervalHours:           12,
		HealthHistoryDefaultResponseNumber: 1000,
		FailedLoginRetentionDays:           3,
		AuthBcryptCost:                     10,
		AuthSessionTTLMinutes:              60,
		AuthSessionCookieName:              "test_cookie",
		AuthLoginBackoffWindowMinutes:      10,
		AuthLoginBackoffThreshold:          3,
		AuthLoginBackoffBaseSeconds:        1,
		AuthLoginBackoffMaxSeconds:         60,
		SMTPUser:                           "test@smtp.com",
		DatabasePath:                       "test/save.db",
	}

	err = SaveConfigurationToFiles()

	assert.NoError(t, err)

	_, err = os.Stat(applicationPropertiesFilePath)
	assert.NoError(t, err)
	_, err = os.Stat(smtpPropertiesFilePath)
	assert.NoError(t, err)
	_, err = os.Stat(databasePropertiesFilePath)
	assert.NoError(t, err)

	appContent, err := os.ReadFile(applicationPropertiesFilePath)
	assert.NoError(t, err)
	assert.Contains(t, string(appContent), "sensor.collection.interval=120")

	smtpContent, err := os.ReadFile(smtpPropertiesFilePath)
	assert.NoError(t, err)
	assert.Contains(t, string(smtpContent), "smtp.user=test@smtp.com")

	dbContent, err := os.ReadFile(databasePropertiesFilePath)
	assert.NoError(t, err)
	assert.Contains(t, string(dbContent), "database.path=test/save.db")
}

func TestSaveConfigurationToFiles_NilAppConfig(t *testing.T) {
	origConfig := AppConfig
	defer func() { AppConfig = origConfig }()

	AppConfig = nil

	err := SaveConfigurationToFiles()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no application configuration loaded")
}

func TestSaveConfigurationToFiles_InvalidPath(t *testing.T) {
	origAppPath := applicationPropertiesFilePath
	defer func() { applicationPropertiesFilePath = origAppPath }()

	applicationPropertiesFilePath = "/nonexistent/dir/application.properties"

	origConfig := AppConfig
	defer func() { AppConfig = origConfig }()

	AppConfig = &ApplicationConfiguration{
		SensorCollectionInterval: 300,
	}

	err := SaveConfigurationToFiles()

	assert.Error(t, err)
}

// ============================================================================
// ReloadConfig and edge case tests
// ============================================================================

func TestReloadConfig_Success(t *testing.T) {
	origConfig := AppConfig
	defer func() { AppConfig = origConfig }()

	appProps := validAppPropsMap()
	smtpProps := validSmtpPropsMap()
	dbProps := validDbPropsMap()

	ReloadConfig(appProps, smtpProps, dbProps)

	assert.NotNil(t, AppConfig)
	assert.Equal(t, 300, AppConfig.SensorCollectionInterval)
	assert.Equal(t, "test/sensor_hub.db", AppConfig.DatabasePath)
}

func TestReloadConfig_InvalidConfig(t *testing.T) {
	origConfig := AppConfig
	defer func() { AppConfig = origConfig }()

	AppConfig = &ApplicationConfiguration{SensorCollectionInterval: 100}

	appProps := validAppPropsMap()
	appProps["sensor.collection.interval"] = "invalid"

	ReloadConfig(appProps, validSmtpPropsMap(), validDbPropsMap())

	assert.Equal(t, 100, AppConfig.SensorCollectionInterval)
}

func TestSensitivePropertiesKeys(t *testing.T) {
	assert.Empty(t, SensitivePropertiesKeys)
}

func TestApplicationPropertiesDefaults_HasExpectedKeys(t *testing.T) {
	// Note: ApplicationPropertiesDefaults may be modified by ReadApplicationPropertiesFile
	// due to direct map assignment in the implementation
	// Test the map has the expected keys
	_, hasInterval := ApplicationPropertiesDefaults["sensor.collection.interval"]
	_, hasBcryptCost := ApplicationPropertiesDefaults["auth.bcrypt.cost"]
	_, hasCookieName := ApplicationPropertiesDefaults["auth.session.cookie.name"]
	assert.True(t, hasInterval)
	assert.True(t, hasBcryptCost)
	assert.True(t, hasCookieName)
}

func TestSmtpPropertiesDefaults_Initial(t *testing.T) {
	// Note: SmtpPropertiesDefaults may be modified by ReadSMTPPropertiesFile
	// due to direct map assignment in the implementation
	// Test the map has the expected keys
	_, hasUser := SmtpPropertiesDefaults["smtp.user"]
	assert.True(t, hasUser)
}

func TestDatabasePropertiesDefaults_Initial(t *testing.T) {
	// Note: DatabasePropertiesDefaults may be modified by ReadDatabasePropertiesFile
	// due to direct map assignment in the implementation
	// Test the map has the expected keys
	_, hasPath := DatabasePropertiesDefaults["database.path"]
	assert.True(t, hasPath)
}

func TestLoadConfigurationFromMaps_OAuthConfig(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["oauth.credentials.file.path"] = "/custom/creds.json"
	appProps["oauth.token.file.path"] = "/custom/token.json"
	appProps["oauth.token.refresh.interval.minutes"] = "45"
	smtpProps := validSmtpPropsMap()
	dbProps := validDbPropsMap()

	cfg, err := LoadConfigurationFromMaps(appProps, smtpProps, dbProps)

	assert.NoError(t, err)
	assert.Equal(t, "/custom/creds.json", cfg.OAuthCredentialsFilePath)
	assert.Equal(t, "/custom/token.json", cfg.OAuthTokenFilePath)
	assert.Equal(t, 45, cfg.OAuthTokenRefreshIntervalMinutes)
}

func TestLoadConfigurationFromMaps_InvalidOAuthTokenRefreshInterval(t *testing.T) {
	appProps := validAppPropsMap()
	appProps["oauth.token.refresh.interval.minutes"] = "not-a-number"
	smtpProps := validSmtpPropsMap()
	dbProps := validDbPropsMap()

	cfg, err := LoadConfigurationFromMaps(appProps, smtpProps, dbProps)

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConvertConfigurationToMaps_OAuthConfig(t *testing.T) {
	cfg := &ApplicationConfiguration{
		SensorCollectionInterval:           300,
		SensorDiscoverySkip:                true,
		OpenAPILocation:                    "/path/to/openapi.yaml",
		HealthHistoryRetentionDays:         180,
		SensorDataRetentionDays:            365,
		DataCleanupIntervalHours:           24,
		HealthHistoryDefaultResponseNumber: 1000,
		FailedLoginRetentionDays:           2,
		AuthBcryptCost:                     12,
		AuthSessionTTLMinutes:              43200,
		AuthSessionCookieName:              "sensor_hub_session",
		AuthLoginBackoffWindowMinutes:      15,
		AuthLoginBackoffThreshold:          5,
		AuthLoginBackoffBaseSeconds:        2,
		AuthLoginBackoffMaxSeconds:         300,
		OAuthCredentialsFilePath:           "/my/creds.json",
		OAuthTokenFilePath:                 "/my/token.json",
		OAuthTokenRefreshIntervalMinutes:   60,
		SMTPUser:                           "user@example.com",
		DatabasePath:                       "test/oauth.db",
	}

	appProps, _, _ := ConvertConfigurationToMaps(cfg)

	assert.Equal(t, "/my/creds.json", appProps["oauth.credentials.file.path"])
	assert.Equal(t, "/my/token.json", appProps["oauth.token.file.path"])
	assert.Equal(t, "60", appProps["oauth.token.refresh.interval.minutes"])
}

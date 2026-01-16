package appProps

import (
	"example/sensorHub/utils"
	"fmt"
	"log"
	"os"
	"strconv"
)

var applicationProperties map[string]string
var smtpProperties map[string]string
var databaseProperties map[string]string

var applicationPropertiesFilePath = "configuration/application.properties"
var smtpPropertiesFilePath = "configuration/smtp.properties"
var databasePropertiesFilePath = "configuration/database.properties"

func validateApplicationProperties() error {
	_, err := strconv.ParseFloat(applicationProperties["email.alert.high.temperature.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid email high threshold value: %s", applicationProperties["email.alert.high.temperature.threshold"])
	}

	_, err = strconv.ParseFloat(applicationProperties["email.alert.low.temperature.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid email low threshold value: %s", applicationProperties["email.alert.low.temperature.threshold"])
	}

	sensorCollectionInterval, err := strconv.Atoi(applicationProperties["sensor.collection.interval"])
	if err != nil || sensorCollectionInterval <= 0 {
		return fmt.Errorf("invalid sensor collection interval value: %s", applicationProperties["sensor.collection.interval"])
	}

	sensorDiscoverySkip := applicationProperties["sensor.discovery.skip"]
	if sensorDiscoverySkip != "true" && sensorDiscoverySkip != "false" {
		return fmt.Errorf("invalid sensor discovery skip value: %s. must be 'true' or 'false'", sensorDiscoverySkip)
	}

	openAPILocation := applicationProperties["openapi.yaml.location"]
	if openAPILocation == "" && sensorDiscoverySkip == "false" {
		return fmt.Errorf("openapi.yaml.location cannot be empty if sensor discovery is not skipped")
	}

	sensorDataRetentionDaysStr := applicationProperties["sensor.data.retention.days"]
	if sensorDataRetentionDaysStr != "" {
		sensorDataRetentionDays, err := strconv.Atoi(sensorDataRetentionDaysStr)
		if err != nil || sensorDataRetentionDays < 0 {
			return fmt.Errorf("invalid sensor data retention days value: %s", sensorDataRetentionDaysStr)
		}
	}

	healthHistoryRetentionDaysStr := applicationProperties["health.history.retention.days"]
	if healthHistoryRetentionDaysStr != "" {
		healthHistoryRetentionDays, err := strconv.Atoi(healthHistoryRetentionDaysStr)
		if err != nil || healthHistoryRetentionDays < 0 {
			return fmt.Errorf("invalid health history retention days value: %s", healthHistoryRetentionDaysStr)
		}
	}

	dataCleanupIntervalHoursStr := applicationProperties["data.cleanup.interval.hours"]
	if dataCleanupIntervalHoursStr != "" {
		dataCleanupIntervalHours, err := strconv.Atoi(dataCleanupIntervalHoursStr)
		if err != nil || dataCleanupIntervalHours <= 0 {
			return fmt.Errorf("invalid data cleanup interval hours value: %s", dataCleanupIntervalHoursStr)
		}
	}

	healthHistoryDefaultResponseNumberStr := applicationProperties["health.history.default.response.number"]
	if healthHistoryDefaultResponseNumberStr != "" {
		healthHistoryDefaultResponseNumber, err := strconv.Atoi(healthHistoryDefaultResponseNumberStr)
		if err != nil || healthHistoryDefaultResponseNumber <= 0 {
			return fmt.Errorf("invalid health history default response number value: %s", healthHistoryDefaultResponseNumberStr)
		}
	}

	failedLoginRetentionDaysStr := applicationProperties["failed.login.retention.days"]
	if failedLoginRetentionDaysStr != "" {
		failedLoginRetentionDays, err := strconv.Atoi(failedLoginRetentionDaysStr)
		if err != nil || failedLoginRetentionDays < 0 {
			return fmt.Errorf("invalid failed login retention days value: %s", failedLoginRetentionDaysStr)
		}
	}

	return nil
}

func validateSMTPProperties() error {
	if smtpProperties["smtp.user"] == "" || smtpProperties["smtp.recipient"] == "" {
		log.Printf("smtp.user or smtp.recipient is empty, email alerts will not be sent. Please check your smtp.properties file")
	}
	return nil
}

func dbValidateDatabaseProperties() error {
	if databaseProperties["database.username"] == "" || databaseProperties["database.password"] == "" ||
		databaseProperties["database.hostname"] == "" || databaseProperties["database.port"] == "" {
		return fmt.Errorf("database properties are not set correctly. please check your database.properties file")
	}
	return nil
}

func ReadApplicationPropertiesFile() (map[string]string, error) {
	applicationProperties = ApplicationPropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(applicationPropertiesFilePath)

	if err != nil {
		return nil, fmt.Errorf("failed to read application properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		applicationProperties[k] = v
	}

	if err := validateApplicationProperties(); err != nil {
		return nil, fmt.Errorf("validation failed on application properties: %w", err)
	}

	return applicationProperties, nil
}

func ReadDatabasePropertiesFile() (map[string]string, error) {
	databaseProperties = DatabasePropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(databasePropertiesFilePath)

	if err != nil {
		return nil, fmt.Errorf("failed to read database properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		databaseProperties[k] = v
	}

	if err := dbValidateDatabaseProperties(); err != nil {
		return nil, fmt.Errorf("validation failed on database properties: %w", err)
	}

	return databaseProperties, nil
}

func ReadSMTPPropertiesFile() (map[string]string, error) {
	smtpProperties = SmtpPropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(smtpPropertiesFilePath)

	if err != nil {
		return nil, fmt.Errorf("failed to read SMTP properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		smtpProperties[k] = v
	}

	if err := validateSMTPProperties(); err != nil {
		return nil, fmt.Errorf("validation failed on SMTP properties: %w", err)
	}

	return smtpProperties, nil
}

func SaveConfigurationToFiles() error {
	if AppConfig == nil {
		log.Printf("No application configuration loaded; cannot save")
		return fmt.Errorf("no application configuration loaded; cannot save")
	}

	applicationPropertiesFile, err := os.OpenFile(applicationPropertiesFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := applicationPropertiesFile.Close(); cerr != nil {
			log.Printf("error closing application properties file: %v", cerr)
		}
	}()

	smtpPropertiesFile, err := os.OpenFile(smtpPropertiesFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := smtpPropertiesFile.Close(); cerr != nil {
			log.Printf("error closing smtp properties file: %v", cerr)
		}
	}()

	databasePropertiesFile, err := os.OpenFile(databasePropertiesFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := databasePropertiesFile.Close(); cerr != nil {
			log.Printf("error closing database properties file: %v", cerr)
		}
	}()

	// helper for writing and checking errors
	writeLine := func(f *os.File, line string) error {
		if _, werr := f.WriteString(line); werr != nil {
			return werr
		}
		return nil
	}

	if err := writeLine(applicationPropertiesFile, "email.alert.high.temperature.threshold="+strconv.FormatFloat(AppConfig.EmailAlertHighTemperatureThreshold, 'f', -1, 64)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "email.alert.low.temperature.threshold="+strconv.FormatFloat(AppConfig.EmailAlertLowTemperatureThreshold, 'f', -1, 64)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "sensor.collection.interval="+strconv.Itoa(AppConfig.SensorCollectionInterval)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "sensor.discovery.skip="+strconv.FormatBool(AppConfig.SensorDiscoverySkip)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "openapi.yaml.location="+AppConfig.OpenAPILocation+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "health.history.retention.days="+strconv.Itoa(AppConfig.HealthHistoryRetentionDays)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "sensor.data.retention.days="+strconv.Itoa(AppConfig.SensorDataRetentionDays)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "data.cleanup.interval.hours="+strconv.Itoa(AppConfig.DataCleanupIntervalHours)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "health.history.default.response.number="+strconv.Itoa(AppConfig.HealthHistoryDefaultResponseNumber)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "failed.login.retention.days="+strconv.Itoa(AppConfig.FailedLoginRetentionDays)+"\n"); err != nil {
		return err
	}

	// auth related
	if err := writeLine(applicationPropertiesFile, "auth.bcrypt.cost="+strconv.Itoa(AppConfig.AuthBcryptCost)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "auth.session.ttl.minutes="+strconv.Itoa(AppConfig.AuthSessionTTLMinutes)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "auth.session.cookie.name="+AppConfig.AuthSessionCookieName+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "auth.login.backoff.window.minutes="+strconv.Itoa(AppConfig.AuthLoginBackoffWindowMinutes)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "auth.login.backoff.threshold="+strconv.Itoa(AppConfig.AuthLoginBackoffThreshold)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "auth.login.backoff.base.seconds="+strconv.Itoa(AppConfig.AuthLoginBackoffBaseSeconds)+"\n"); err != nil {
		return err
	}
	if err := writeLine(applicationPropertiesFile, "auth.login.backoff.max.seconds="+strconv.Itoa(AppConfig.AuthLoginBackoffMaxSeconds)+"\n"); err != nil {
		return err
	}

	if err := writeLine(smtpPropertiesFile, "smtp.user="+AppConfig.SMTPUser+"\n"); err != nil {
		return err
	}
	if err := writeLine(smtpPropertiesFile, "smtp.recipient="+AppConfig.SMTPRecipient+"\n"); err != nil {
		return err
	}

	if err := writeLine(databasePropertiesFile, "database.username="+AppConfig.DatabaseUsername+"\n"); err != nil {
		return err
	}
	if err := writeLine(databasePropertiesFile, "database.password="+AppConfig.DatabasePassword+"\n"); err != nil {
		return err
	}
	if err := writeLine(databasePropertiesFile, "database.hostname="+AppConfig.DatabaseHostname+"\n"); err != nil {
		return err
	}
	if err := writeLine(databasePropertiesFile, "database.port="+AppConfig.DatabasePort+"\n"); err != nil {
		return err
	}

	return nil
}

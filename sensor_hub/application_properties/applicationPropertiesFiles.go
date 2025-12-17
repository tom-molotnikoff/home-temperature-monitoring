package appProps

import (
	"example/sensorHub/utils"
	"fmt"
	"log"
	"strconv"
)

var applicationProperties map[string]string
var smtpProperties map[string]string
var databaseProperties map[string]string

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

func ReadApplicationPropertiesFile(filePath string) error {
	applicationProperties = ApplicationPropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(filePath)

	if err != nil {
		return fmt.Errorf("failed to read application properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		applicationProperties[k] = v
	}

	if err := validateApplicationProperties(); err != nil {
		return fmt.Errorf("validation failed on application properties: %w", err)
	}

	return nil
}

func ReadDatabasePropertiesFile(filePath string) error {
	databaseProperties = DatabasePropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(filePath)

	if err != nil {
		return fmt.Errorf("failed to read database properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		databaseProperties[k] = v
	}

	if err := dbValidateDatabaseProperties(); err != nil {
		return fmt.Errorf("validation failed on database properties: %w", err)
	}

	return nil
}

func ReadSMTPPropertiesFile(filePath string) error {
	smtpProperties = SmtpPropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(filePath)

	if err != nil {
		return fmt.Errorf("failed to read SMTP properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		smtpProperties[k] = v
	}

	if err := validateSMTPProperties(); err != nil {
		return fmt.Errorf("validation failed on SMTP properties: %w", err)
	}

	return nil
}

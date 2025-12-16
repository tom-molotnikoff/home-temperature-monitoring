package appProps

import (
	"example/sensorHub/utils"
	"fmt"
	"log"
	"slices"
	"strconv"
)

var ApplicationProperties map[string]string
var SmtpProperties map[string]string
var DatabaseProperties map[string]string

func validateApplicationProperties() error {
	_, err := strconv.ParseFloat(ApplicationProperties["email.alert.high.temperature.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid email high threshold value: %s", ApplicationProperties["email.alert.high.temperature.threshold"])
	}

	_, err = strconv.ParseFloat(ApplicationProperties["email.alert.low.temperature.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid email low threshold value: %s", ApplicationProperties["email.alert.low.temperature.threshold"])
	}

	sensorCollectionInterval, err := strconv.Atoi(ApplicationProperties["sensor.collection.interval"])
	if err != nil || sensorCollectionInterval <= 0 {
		return fmt.Errorf("invalid sensor collection interval value: %s", ApplicationProperties["sensor.collection.interval"])
	}

	sensorDiscoverySkip := ApplicationProperties["sensor.discovery.skip"]
	if sensorDiscoverySkip != "true" && sensorDiscoverySkip != "false" {
		return fmt.Errorf("invalid sensor discovery skip value: %s. must be 'true' or 'false'", sensorDiscoverySkip)
	}

	openAPILocation := ApplicationProperties["openapi.yaml.location"]
	if openAPILocation == "" && sensorDiscoverySkip == "false" {
		return fmt.Errorf("openapi.yaml.location cannot be empty if sensor discovery is not skipped")
	}

	return nil
}

func validateSMTPProperties() error {
	if SmtpProperties["smtp.user"] == "" || SmtpProperties["smtp.recipient"] == "" {
		log.Printf("smtp.user or smtp.recipient is empty, email alerts will not be sent. Please check your smtp.properties file")
	}
	return nil
}

func dbValidateDatabaseProperties() error {
	if DatabaseProperties["database.username"] == "" || DatabaseProperties["database.password"] == "" ||
		DatabaseProperties["database.hostname"] == "" || DatabaseProperties["database.port"] == "" {
		return fmt.Errorf("database properties are not set correctly. please check your database.properties file")
	}
	return nil
}

func ReadApplicationPropertiesFile(filePath string) error {
	ApplicationProperties = ApplicationPropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(filePath)

	if err != nil {
		return fmt.Errorf("failed to read application properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		ApplicationProperties[k] = v
	}

	if err := validateApplicationProperties(); err != nil {
		return fmt.Errorf("validation failed on application properties: %w", err)
	}

	logPropertiesFilterSensitive("ApplicationProperties", ApplicationProperties)
	return nil
}

func ReadDatabasePropertiesFile(filePath string) error {
	DatabaseProperties = DatabasePropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(filePath)

	if err != nil {
		return fmt.Errorf("failed to read database properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		DatabaseProperties[k] = v
	}

	if err := dbValidateDatabaseProperties(); err != nil {
		return fmt.Errorf("validation failed on database properties: %w", err)
	}

	logPropertiesFilterSensitive("DatabaseProperties", DatabaseProperties)
	return nil
}

func ReadSMTPPropertiesFile(filePath string) error {
	SmtpProperties = SmtpPropertiesDefaults
	propertiesFromFile, err := utils.ReadPropertiesFile(filePath)

	if err != nil {
		return fmt.Errorf("failed to read SMTP properties file: %w", err)
	}

	for k, v := range propertiesFromFile {
		SmtpProperties[k] = v
	}

	if err := validateSMTPProperties(); err != nil {
		return fmt.Errorf("validation failed on SMTP properties: %w", err)
	}

	logPropertiesFilterSensitive("SmtpProperties", SmtpProperties)
	return nil
}

func logPropertiesFilterSensitive(name string, propsMap map[string]string) {
	propsFiltered := make(map[string]string)
	for k, v := range propsMap {
		if slices.Contains(SensitivePropertiesKeys, k) {
			propsFiltered[k] = "****"
		} else {
			propsFiltered[k] = v
		}
	}
	log.Printf("%s properties: %v", name, propsFiltered)
}

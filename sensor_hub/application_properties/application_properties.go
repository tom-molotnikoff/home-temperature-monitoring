package appProps

import (
	"example/sensorHub/utils"
	"fmt"
	"log"
	"slices"
)

var APPLICATION_PROPERTIES map[string]string
var SMTP_PROPERTIES map[string]string
var DATABASE_PROPERTIES map[string]string

func validateSMTPProperties(propsMap map[string]string) error {
	if propsMap["smtp.user"] == "" || propsMap["smtp.recipient"] == "" {
		return fmt.Errorf("smtp properties are not set correctly. Please check your smtp.properties file")
	}
	return nil
}

// This function validates the database properties by checking if the required fields are set.
func dbValidateDatabaseProperties(propsMap map[string]string) error {
	if propsMap["database.username"] == "" || propsMap["database.password"] == "" ||
		propsMap["database.hostname"] == "" || propsMap["database.port"] == "" {
		return fmt.Errorf("database properties are not set correctly. please check your database.properties file")
	}
	return nil
}

func ReadApplicationPropertiesFile(filePath string) error {
	propsMap, err := utils.ReadPropertiesFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read application properties file: %w", err)
	}
	APPLICATION_PROPERTIES = propsMap
	logPropertiesFilterSensitive("APPLICATION_PROPERTIES", APPLICATION_PROPERTIES)
	return nil
}

func ReadDatabasePropertiesFile(filePath string) error {
	propsMap, err := utils.ReadPropertiesFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read database properties file: %w", err)
	}
	if err := dbValidateDatabaseProperties(propsMap); err != nil {
		return fmt.Errorf("validation failed on database properties: %w", err)
	}
	DATABASE_PROPERTIES = propsMap
	logPropertiesFilterSensitive("DATABASE_PROPERTIES", DATABASE_PROPERTIES)
	return nil
}

func ReadSMTPPropertiesFile(filePath string) error {
	propsMap, err := utils.ReadPropertiesFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SMTP properties file: %w", err)
	}
	if err := validateSMTPProperties(propsMap); err != nil {
		return fmt.Errorf("validation failed on SMTP properties: %w", err)
	}
	SMTP_PROPERTIES = propsMap
	logPropertiesFilterSensitive("SMTP", SMTP_PROPERTIES)
	return nil
}

func logPropertiesFilterSensitive(name string, propsMap map[string]string) {
	sensitiveProperties := []string{"database.password"}
	propsFiltered := make(map[string]string)
	for k, v := range propsMap {
		if slices.Contains(sensitiveProperties, k) {
			propsFiltered[k] = "****"
		} else {
			propsFiltered[k] = v
		}
	}
	log.Printf("%s properties: %v", name, propsFiltered)
}

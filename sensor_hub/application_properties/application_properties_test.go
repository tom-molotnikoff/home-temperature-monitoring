package appProps

import (
	"bytes"
	"example/sensorHub/utils"
	"fmt"
	"log"
	"strings"
	"testing"
)

func mockReadPropertiesFile_applicationProps(filePath string) (map[string]string, error) {
	return map[string]string{
		"email.alert.high.threshold":          "30",
		"email.alert.low.threshold":           "12",
		"openapi.yaml.location":               "./docker_tests/openapi.yaml",
		"sensor.collection.interval":          "300",
		"current.temperature.websocket.interval": "1",
	}, nil
}

func mockReadPropertiesFile_error(filePath string) (map[string]string, error) {
	return nil, fmt.Errorf("failed to open properties file: %w", fmt.Errorf("file not found"))
}

func TestReadApplicationPropertiesFile(t *testing.T) {
	var originalMethod = utils.ReadPropertiesFile
	utils.ReadPropertiesFile = mockReadPropertiesFile_applicationProps
	defer func() { utils.ReadPropertiesFile = originalMethod }()

	err := ReadApplicationPropertiesFile("application.properties")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(APPLICATION_PROPERTIES) != 5 {
		t.Fatalf("Expected 5 properties, got %d", len(APPLICATION_PROPERTIES))
	}
	expectedKeys := []string{
		"email.alert.high.threshold",
		"email.alert.low.threshold",
		"openapi.yaml.location",
		"sensor.collection.interval",
		"current.temperature.websocket.interval",
	}
	for _, key := range expectedKeys {
		if _, exists := APPLICATION_PROPERTIES[key]; !exists {
			t.Errorf("Expected key %s to be present", key)
		}
	}
}

func TestReadApplicationPropertiesFile_ErrorOpeningFile(t *testing.T) {
	var originalMethod = utils.ReadPropertiesFile
	utils.ReadPropertiesFile = mockReadPropertiesFile_error
	defer func() { utils.ReadPropertiesFile = originalMethod }()

	err := ReadApplicationPropertiesFile("application.properties")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestValidateSMTPProperties(t *testing.T) {
	propsMap := map[string]string{
		"smtp.user":      "user@example.com",
		"smtp.recipient": "recipient@example.com",
	}
	err := validateSMTPProperties(propsMap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestValidateSMTPProperties_MissingUser(t *testing.T) {
	propsMap := map[string]string{
		"smtp.user":      "",
		"smtp.recipient": "recipient@example.com",
	}
	err := validateSMTPProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}
func TestValidateSMTPProperties_MissingRecipient(t *testing.T) {
	propsMap := map[string]string{
		"smtp.user":      "user@example.com",
		"smtp.recipient": "",
	}
	err := validateSMTPProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}
func TestValidateSMTPProperties_MissingBoth(t *testing.T) {
	propsMap := map[string]string{
		"smtp.user":      "",
		"smtp.recipient": "",
	}
	err := validateSMTPProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestValidateSMTPProperties_MissingKeys(t *testing.T) {
	propsMap := map[string]string{}
	err := validateSMTPProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestDbValidateDatabaseProperties(t *testing.T) {
	propsMap := map[string]string{
		"database.username": "user",
		"database.password": "password",
		"database.hostname": "localhost",
		"database.port":     "3306",
	}
	err := dbValidateDatabaseProperties(propsMap)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDbValidateDatabaseProperties_MissingUsername(t *testing.T) {
	propsMap := map[string]string{
		"database.username": "",
		"database.password": "password",
		"database.hostname": "localhost",
		"database.port":     "3306",
	}
	err := dbValidateDatabaseProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestDbValidateDatabaseProperties_MissingPassword(t *testing.T) {
	propsMap := map[string]string{
		"database.username": "user",
		"database.password": "",
		"database.hostname": "localhost",
		"database.port":     "3306",
	}
	err := dbValidateDatabaseProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestDbValidateDatabaseProperties_MissingHostname(t *testing.T) {
	propsMap := map[string]string{
		"database.username": "user",
		"database.password": "password",
		"database.hostname": "",
		"database.port":     "3306",
	}
	err := dbValidateDatabaseProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestDbValidateDatabaseProperties_MissingPort(t *testing.T) {
	propsMap := map[string]string{
		"database.username": "user",
		"database.password": "password",
		"database.hostname": "localhost",
		"database.port":     "",
	}
	err := dbValidateDatabaseProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestDbValidateDatabaseProperties_MissingAll(t *testing.T) {
	propsMap := map[string]string{}
	err := dbValidateDatabaseProperties(propsMap)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestLogPropertiesFilterSensitive(t *testing.T) {
	propsMap := map[string]string{
		"database.username": "user",
		"database.password": "secretstuff!",
		"database.hostname": "localhost",
		"database.port":     "3306",
		"smtp.user":        "smtp_user",
		"smtp.recipient":   "smtp_recipient",
		"other.key":        "other_value",
	}

    var buf bytes.Buffer
    log.SetFlags(0) 
    log.SetOutput(&buf)
	defer func() { log.SetOutput(nil) }()

	logPropertiesFilterSensitive("props", propsMap)

	
	for _, line := range strings.Split(buf.String(), "\n") {

		if strings.Contains(line, "secretstuff!") {
			t.Errorf("Sensitive information logged: %s", line)
		}
	}
}
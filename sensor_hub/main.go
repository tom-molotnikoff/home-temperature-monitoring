package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

var APPLICATION_PROPERTIES map[string]string
var DATABASE_PROPERTIES map[string]string
var SMTP_PROPERTIES map[string]string

// initialise_application reads the properties files, validates them, and sets up the database connection.
// It returns an error if any of the properties files are missing or not set correctly.
func initialise_application() error {
	db_props, err := read_properties_file("configuration/database.properties")
	if err != nil {
		return fmt.Errorf("failed to read database properties file: %w", err)
	}
	DATABASE_PROPERTIES = db_props
	err = validateDatabaseProperties()
	if err != nil {
		return fmt.Errorf("database properties are not set correctly: %w", err)
	}

	SMTP_PROPERTIES, err = read_properties_file("configuration/smtp.properties")
	if err != nil {
		log.Printf("SMTP properties file missing, email alerts will be disabled: %s", err)
	} else {
		validationErr := validateSMTPProperties()
		if validationErr != nil {
			log.Printf("SMTP properties are present but not set correctly, email alerts will be disabled: %s", validationErr)
		}
	}
	APPLICATION_PROPERTIES, err = read_properties_file("configuration/application.properties")
	if err != nil {
		log.Printf("Failed to read application properties file: %s", err)
	}
	log.Printf("Application properties: %v", APPLICATION_PROPERTIES)

	DB, err = initialise_database(DATABASE_PROPERTIES)
	if err != nil {
		return fmt.Errorf("failed to initialise database: %w", err)
	}

	err = create_temperature_readings_table()
	if err != nil {
		return fmt.Errorf("failed to create temperature readings table: %w", err)
	}
	return nil
}

func startPeriodicSensorCollection() {
	intervalStr := APPLICATION_PROPERTIES["sensor.collection.interval"]
	intervalSec, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("Invalid sensor.collection.interval value: %s, defaulting to 60 seconds", intervalStr)
		intervalSec = 60
	}
	go func() {
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()
		for {
			_, err := take_readings()
			if err != nil {
				log.Printf("Error taking readings: %s", err)
			}

			<-ticker.C
		}
	}()
}

// This application will read the temperature from sensors through their APIs
// persist the readings to a database, and send an email alert if the
// temperature exceeds a threshold defined in the application properties.
func main() {

	log.SetPrefix("sensor-hub: ")

	err := initialise_application()
	if err != nil {
		log.Fatalf("Failed to initialise application: %s", err)
	}

	err = discover_sensor_urls()
	if err != nil {
		log.Fatalf("Failed to discover sensor URLs: %s", err)
	}

	err = initialise_oauth()
	if err != nil {
		log.Printf("Failed to initialise OAuth: %s", err)
	}
	startPeriodicSensorCollection()
	initalise_api_and_listen()
}

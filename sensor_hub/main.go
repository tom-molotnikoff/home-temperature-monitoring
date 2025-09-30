package main

import (
	"example/sensorHub/api"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/oauth"
	"example/sensorHub/sensors"
	"fmt"
	"log"
)

// initialise_application reads the properties files, validates them, and sets up the database connection.
// It returns an error if any of the properties files are missing or not set correctly.
func mainInitialiseApplication() error {

	err := appProps.ReadDatabasePropertiesFile("configuration/database.properties")
	if err != nil {
		return fmt.Errorf("failed to read database properties file: %w", err)
	}

	err = appProps.ReadSMTPPropertiesFile("configuration/smtp.properties")
	if err != nil {
		// non-fatal error - email alerts will not work
		log.Printf("Failed to read SMTP properties file: %v", err)
	}

	err = appProps.ReadApplicationPropertiesFile("configuration/application.properties")
	if err != nil {
		return fmt.Errorf("failed to read application properties file: %w", err)
	}

	err = database.InitialiseDatabase()
	if err != nil {
		return fmt.Errorf("failed to initialise database: %w", err)
	}

	return nil
}

// This application will read the temperature from sensors through their APIs
// persist the readings to a database, and send an email alert if the
// temperature exceeds a threshold defined in the application properties.
func main() {

	log.SetPrefix("sensor-hub: ")

	err := mainInitialiseApplication()
	if err != nil {
		log.Fatalf("Failed to initialise application: %v", err)
	}

	// Clean up after yourself!
	defer database.DB.Close()

	err = sensors.DiscoverSensors()
	if err != nil {
		log.Fatalf("Failed to discover sensors: %v", err)
	}

	err = oauth.InitialiseOauth()
	if err != nil {
		log.Printf("Failed to initialise OAuth: %v", err)
	}
	sensors.StartPeriodicSensorCollection()
	api.InitialiseAndListen()
}

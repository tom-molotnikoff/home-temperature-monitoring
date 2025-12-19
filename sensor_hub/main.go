package main

import (
	"database/sql"
	"example/sensorHub/api"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/oauth"
	"example/sensorHub/service"
	"log"
)

func main() {

	log.SetPrefix("sensor-hub: ")

	err := appProps.InitialiseConfig()
	if err != nil {
		log.Fatalf("failed to initialise application configuration: %v", err)
	}

	db, err := database.InitialiseDatabase()
	if err != nil {
		log.Fatalf("failed to initialise database: %v", err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}(db)

	sensorRepo := database.NewSensorRepository(db)
	tempRepo := database.NewTemperatureRepository(db, sensorRepo)

	sensorService := service.NewSensorService(sensorRepo, tempRepo)
	tempService := service.NewTemperatureService(tempRepo)
	propertiesService := service.NewPropertiesService()

	api.InitTemperatureAPI(tempService)
	api.InitSensorAPI(sensorService)
	api.InitPropertiesAPI(propertiesService)

	err = sensorService.ServiceDiscoverSensors()

	if err != nil {
		log.Fatalf("Failed to discover sensors: %v", err)
	}

	err = oauth.InitialiseOauth()
	if err != nil {
		log.Printf("Failed to initialise OAuth: %v", err)
	}
	sensorService.ServiceStartPeriodicSensorCollection()

	err = api.InitialiseAndListen()
	if err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

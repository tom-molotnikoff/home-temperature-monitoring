package service

import (
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"log"
	"time"
)

type cleanupService struct {
	sensorRepo      database.SensorRepositoryInterface[types.Sensor]
	temperatureRepo database.ReadingsRepository[types.TemperatureReading]
}

func NewCleanupService(sensorRepo database.SensorRepositoryInterface[types.Sensor], temperatureRepo database.ReadingsRepository[types.TemperatureReading]) CleanupServiceInterface {
	return &cleanupService{
		sensorRepo:      sensorRepo,
		temperatureRepo: temperatureRepo,
	}
}

func (cs *cleanupService) StartPeriodicCleanup() {
	healthHistoryRetentionDays := appProps.AppConfig.HealthHistoryRetentionDays
	sensorDataRetentionDays := appProps.AppConfig.SensorDataRetentionDays

	go func() {
		ticker := time.NewTicker(time.Duration(appProps.AppConfig.DataCleanupIntervalHours) * time.Hour)
		defer ticker.Stop()
		for {
			err := cs.performCleanup(healthHistoryRetentionDays, sensorDataRetentionDays)
			if err != nil {
				log.Printf("Error during periodic cleanup: %v", err)
				continue
			}
			<-ticker.C
		}
	}()
}

func (cs *cleanupService) performCleanup(healthHistoryRetentionDays int, sensorDataRetentionDays int) error {
	log.Printf("Starting data cleanup: Health history retention = %d days, Sensor data retention = %d days", healthHistoryRetentionDays, sensorDataRetentionDays)
	log.Printf("Cleaning up old temperature readings...")
	if sensorDataRetentionDays > 0 {
		err := cs.temperatureRepo.DeleteReadingsOlderThan(time.Now().AddDate(0, 0, -sensorDataRetentionDays))
		if err != nil {
			return err
		}
		log.Printf("Deleted temperature readings older than %d days", sensorDataRetentionDays)
	}
	log.Printf("Temperature readings cleanup completed successfully")
	log.Printf("Cleaning up old health history...")
	if healthHistoryRetentionDays > 0 {
		err := cs.sensorRepo.DeleteHealthHistoryOlderThan(time.Now().AddDate(0, 0, -healthHistoryRetentionDays))
		if err != nil {
			return err
		}
		log.Printf("Deleted health history records older than %d days", healthHistoryRetentionDays)
	}
	log.Printf("Health history cleanup completed successfully")
	return nil
}

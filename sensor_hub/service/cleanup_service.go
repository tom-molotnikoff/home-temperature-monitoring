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
	failedRepo      database.FailedLoginRepository
}

func NewCleanupService(sensorRepo database.SensorRepositoryInterface[types.Sensor], temperatureRepo database.ReadingsRepository[types.TemperatureReading], failedRepo database.FailedLoginRepository) CleanupServiceInterface {
	return &cleanupService{
		sensorRepo:      sensorRepo,
		temperatureRepo: temperatureRepo,
		failedRepo:      failedRepo,
	}
}

func (cs *cleanupService) StartPeriodicCleanup() {
	healthHistoryRetentionDays := appProps.AppConfig.HealthHistoryRetentionDays
	sensorDataRetentionDays := appProps.AppConfig.SensorDataRetentionDays
	failedLoginRetentionDays := appProps.AppConfig.FailedLoginRetentionDays

	go func() {
		ticker := time.NewTicker(time.Duration(appProps.AppConfig.DataCleanupIntervalHours) * time.Hour)
		defer ticker.Stop()
		for {
			err := cs.performCleanup(healthHistoryRetentionDays, sensorDataRetentionDays, failedLoginRetentionDays)
			if err != nil {
				log.Printf("Error during periodic cleanup: %v", err)
				continue
			}
			<-ticker.C
		}
	}()
}

func (cs *cleanupService) performCleanup(healthHistoryRetentionDays int, sensorDataRetentionDays int, failedLoginRetentionDays int) error {
	if sensorDataRetentionDays > 0 {
		log.Printf("Cleaning up old temperature readings older than %d days...", sensorDataRetentionDays)
		err := cs.temperatureRepo.DeleteReadingsOlderThan(time.Now().AddDate(0, 0, -sensorDataRetentionDays))
		if err != nil {
			return err
		}
		log.Printf("Deleted temperature readings older than %d days", sensorDataRetentionDays)
	}
	log.Printf("Temperature readings cleanup completed successfully")

	if healthHistoryRetentionDays > 0 {
		log.Printf("Cleaning up old health history older than %d days...", healthHistoryRetentionDays)
		err := cs.sensorRepo.DeleteHealthHistoryOlderThan(time.Now().AddDate(0, 0, -healthHistoryRetentionDays))
		if err != nil {
			return err
		}
		log.Printf("Deleted health history records older than %d days", healthHistoryRetentionDays)
	}
	log.Printf("Health history cleanup completed successfully")

	if failedLoginRetentionDays > 0 {
		log.Printf("Cleaning up failed login attempts older than %d days...", failedLoginRetentionDays)
		threshold := time.Now().AddDate(0, 0, -failedLoginRetentionDays)
		if err := cs.failedRepo.DeleteAttemptsOlderThan(threshold); err != nil {
			return err
		}
		log.Printf("Deleted failed login attempts older than %d days", failedLoginRetentionDays)
	}

	return nil
}

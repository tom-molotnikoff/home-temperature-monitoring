package service

import (
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"log"
	"time"
)

type cleanupService struct {
	sensorRepo       database.SensorRepositoryInterface[types.Sensor]
	temperatureRepo  database.ReadingsRepository[types.TemperatureReading]
	failedRepo       database.FailedLoginRepository
	notificationRepo database.NotificationRepository
}

func NewCleanupService(sensorRepo database.SensorRepositoryInterface[types.Sensor], temperatureRepo database.ReadingsRepository[types.TemperatureReading], failedRepo database.FailedLoginRepository, notificationRepo database.NotificationRepository) CleanupServiceInterface {
	return &cleanupService{
		sensorRepo:       sensorRepo,
		temperatureRepo:  temperatureRepo,
		failedRepo:       failedRepo,
		notificationRepo: notificationRepo,
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

	// Clean up old notifications (90 days retention for dismissed notifications)
	if cs.notificationRepo != nil {
		notificationRetentionDays := 90
		log.Printf("Cleaning up old notifications older than %d days...", notificationRetentionDays)
		threshold := time.Now().AddDate(0, 0, -notificationRetentionDays)
		deleted, err := cs.notificationRepo.DeleteOldNotifications(threshold)
		if err != nil {
			log.Printf("Warning: failed to cleanup old notifications: %v", err)
		} else if deleted > 0 {
			log.Printf("Deleted %d notifications older than %d days", deleted, notificationRetentionDays)
		}
	}

	return nil
}

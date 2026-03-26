package service

import (
	"context"

	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"log/slog"
	"time"
)

type cleanupService struct {
	sensorRepo       database.SensorRepositoryInterface[types.Sensor]
	temperatureRepo  database.ReadingsRepository[types.TemperatureReading]
	failedRepo       database.FailedLoginRepository
	notificationRepo database.NotificationRepository
	logger           *slog.Logger
}

func NewCleanupService(sensorRepo database.SensorRepositoryInterface[types.Sensor], temperatureRepo database.ReadingsRepository[types.TemperatureReading], failedRepo database.FailedLoginRepository, notificationRepo database.NotificationRepository, logger *slog.Logger) CleanupServiceInterface {
	return &cleanupService{
		sensorRepo:       sensorRepo,
		temperatureRepo:  temperatureRepo,
		failedRepo:       failedRepo,
		notificationRepo: notificationRepo,
		logger:           logger.With("component", "cleanup_service"),
	}
}

func (cs *cleanupService) StartPeriodicCleanup(ctx context.Context) {
	healthHistoryRetentionDays := appProps.AppConfig.HealthHistoryRetentionDays
	sensorDataRetentionDays := appProps.AppConfig.SensorDataRetentionDays
	failedLoginRetentionDays := appProps.AppConfig.FailedLoginRetentionDays

	go func() {
		ticker := time.NewTicker(time.Duration(appProps.AppConfig.DataCleanupIntervalHours) * time.Hour)
		defer ticker.Stop()
		for {
			err := cs.performCleanup(context.Background(), healthHistoryRetentionDays, sensorDataRetentionDays, failedLoginRetentionDays)
			if err != nil {
				cs.logger.Error("error during periodic cleanup", "error", err)
				continue
			}
			<-ticker.C
		}
	}()

	// Hourly average computation (replaces MySQL EVENT)
	go func() {
		// Run once at startup to catch any missed hours
		if err := cs.temperatureRepo.ComputeHourlyAverages(context.Background()); err != nil {
			cs.logger.Error("error computing hourly averages at startup", "error", err)
		} else {
			cs.logger.Info("hourly average computation completed", "trigger", "startup")
		}

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := cs.temperatureRepo.ComputeHourlyAverages(context.Background()); err != nil {
				cs.logger.Error("error computing hourly averages", "error", err)
			} else {
				cs.logger.Info("hourly average computation completed", "trigger", "periodic")
			}
		}
	}()
}

func (cs *cleanupService) performCleanup(ctx context.Context, healthHistoryRetentionDays int, sensorDataRetentionDays int, failedLoginRetentionDays int) error {
	if sensorDataRetentionDays > 0 {
		cs.logger.Debug("cleaning up old temperature readings", "retention_days", sensorDataRetentionDays)
		err := cs.temperatureRepo.DeleteReadingsOlderThan(ctx, time.Now().AddDate(0, 0, -sensorDataRetentionDays))
		if err != nil {
			return err
		}
		cs.logger.Info("deleted old temperature readings", "retention_days", sensorDataRetentionDays)
	}
	cs.logger.Info("temperature readings cleanup completed")

	if healthHistoryRetentionDays > 0 {
		cs.logger.Debug("cleaning up old health history", "retention_days", healthHistoryRetentionDays)
		err := cs.sensorRepo.DeleteHealthHistoryOlderThan(ctx, time.Now().AddDate(0, 0, -healthHistoryRetentionDays))
		if err != nil {
			return err
		}
		cs.logger.Info("deleted old health history records", "retention_days", healthHistoryRetentionDays)
	}
	cs.logger.Info("health history cleanup completed")

	if failedLoginRetentionDays > 0 {
		cs.logger.Debug("cleaning up old failed login attempts", "retention_days", failedLoginRetentionDays)
		threshold := time.Now().AddDate(0, 0, -failedLoginRetentionDays)
		if err := cs.failedRepo.DeleteAttemptsOlderThan(ctx, threshold); err != nil {
			return err
		}
		cs.logger.Info("deleted old failed login attempts", "retention_days", failedLoginRetentionDays)
	}

	// Clean up old notifications (90 days retention for dismissed notifications)
	if cs.notificationRepo != nil {
		notificationRetentionDays := 90
		cs.logger.Debug("cleaning up old notifications", "retention_days", notificationRetentionDays)
		threshold := time.Now().AddDate(0, 0, -notificationRetentionDays)
		deleted, err := cs.notificationRepo.DeleteOldNotifications(ctx, threshold)
		if err != nil {
			cs.logger.Warn("failed to cleanup old notifications", "error", err)
		} else if deleted > 0 {
			cs.logger.Info("deleted old notifications", "count", deleted, "retention_days", notificationRetentionDays)
		}
	}

	return nil
}

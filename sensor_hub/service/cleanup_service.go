package service

import (
	"context"
	"fmt"

	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/periodic"
	"example/sensorHub/types"
	"log/slog"
	"time"
)

type cleanupService struct {
	sensorRepo       database.SensorRepositoryInterface[types.Sensor]
	readingsRepo     database.ReadingsRepository
	failedRepo       database.FailedLoginRepository
	notificationRepo database.NotificationRepository
	logger           *slog.Logger
}

func NewCleanupService(sensorRepo database.SensorRepositoryInterface[types.Sensor], readingsRepo database.ReadingsRepository, failedRepo database.FailedLoginRepository, notificationRepo database.NotificationRepository, logger *slog.Logger) CleanupServiceInterface {
	return &cleanupService{
		sensorRepo:       sensorRepo,
		readingsRepo:     readingsRepo,
		failedRepo:       failedRepo,
		notificationRepo: notificationRepo,
		logger:           logger.With("component", "cleanup_service"),
	}
}

func (cs *cleanupService) StartPeriodicCleanup(ctx context.Context) {
	healthHistoryRetentionDays := appProps.AppConfig.HealthHistoryRetentionDays
	sensorDataRetentionDays := appProps.AppConfig.SensorDataRetentionDays
	failedLoginRetentionDays := appProps.AppConfig.FailedLoginRetentionDays

	periodic.RunTask(ctx, periodic.TaskConfig{
		Name:           "data_cleanup",
		Interval:       time.Duration(appProps.AppConfig.DataCleanupIntervalHours) * time.Hour,
		Logger:         cs.logger,
		RunImmediately: true,
	}, func(ctx context.Context) error {
		return cs.performCleanup(ctx, healthHistoryRetentionDays, sensorDataRetentionDays, failedLoginRetentionDays)
	})

	periodic.RunTask(ctx, periodic.TaskConfig{
		Name:           "hourly_averages",
		Interval:       1 * time.Hour,
		Logger:         cs.logger,
		RunImmediately: true,
	}, func(ctx context.Context) error {
		if err := cs.readingsRepo.ComputeHourlyAverages(ctx); err != nil {
			return err
		}
		return cs.readingsRepo.ComputeHourlyEvents(ctx)
	})
}

func (cs *cleanupService) performCleanup(ctx context.Context, healthHistoryRetentionDays int, sensorDataRetentionDays int, failedLoginRetentionDays int) error {
	if sensorDataRetentionDays > 0 {
		cs.logger.Debug("cleaning up old sensor readings", "retention_days", sensorDataRetentionDays)

		// Apply per-sensor retention first, collecting the IDs of sensors that have a custom value.
		customSensors, err := cs.sensorRepo.GetSensorsWithRetention(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch sensors with custom retention: %w", err)
		}

		customSensorIds := make([]int, 0, len(customSensors))
		for _, sensor := range customSensors {
			if sensor.RetentionHours == nil {
				continue
			}
			cutoff := time.Now().Add(-time.Duration(*sensor.RetentionHours) * time.Hour)
			if err := cs.readingsRepo.DeleteReadingsOlderThanForSensor(ctx, cutoff, sensor.Id); err != nil {
				return fmt.Errorf("failed per-sensor cleanup for sensor %d: %w", sensor.Id, err)
			}
			customSensorIds = append(customSensorIds, sensor.Id)
		}

		// Apply global retention to all remaining sensors (excluding those already handled above).
		globalCutoff := time.Now().AddDate(0, 0, -sensorDataRetentionDays)
		if err := cs.readingsRepo.DeleteReadingsOlderThanExcludingSensors(ctx, globalCutoff, customSensorIds); err != nil {
			return fmt.Errorf("failed global cleanup: %w", err)
		}
		cs.logger.Info("deleted old sensor readings", "retention_days", sensorDataRetentionDays, "custom_sensors", len(customSensorIds))
	}
	cs.logger.Info("sensor readings cleanup completed")

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

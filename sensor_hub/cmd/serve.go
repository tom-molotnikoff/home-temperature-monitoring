package cmd

import (
	"context"
	"database/sql"
	"example/sensorHub/api"
	"example/sensorHub/api/middleware"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/notifications"
	"example/sensorHub/oauth"
	"example/sensorHub/service"
	"example/sensorHub/smtp"
	"example/sensorHub/telemetry"
	"example/sensorHub/ws"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var configDir string
var logFile string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Sensor Hub server",
	Long:  "Starts the HTTP API server, sensor discovery, periodic collection, and serves the embedded UI.",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&configDir, "config-dir", "configuration", "Path to configuration directory")
	serveCmd.Flags().StringVar(&logFile, "log-file", "", "Path to log file (default: stdout)")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := appProps.InitialiseConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to initialise application configuration: %w", err)
	}

	logLevel := telemetry.ParseLogLevel(appProps.AppConfig.LogLevel)

	tel, err := telemetry.Init(context.Background(), telemetry.Config{
		ServiceName: "sensor-hub",
		Version:     Version,
		LogLevel:    logLevel,
		LogFilePath: logFile,
	})
	if err != nil {
		return fmt.Errorf("failed to initialise telemetry: %w", err)
	}
	defer tel.Shutdown()

	logger := tel.Logger

	db, err := database.InitialiseDatabase(logger)
	if err != nil {
		return fmt.Errorf("failed to initialise database: %w", err)
	}

	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			logger.Error("error closing database", "error", err)
		}
	}(db)

	sensorRepo := database.NewSensorRepository(db, logger)
	tempRepo := database.NewTemperatureRepository(db, sensorRepo, logger)
	alertRepo := database.NewAlertRepository(db, logger)
	notificationRepo := database.NewNotificationRepository(db, logger)

	userRepo := database.NewUserRepository(db, logger)
	sessionRepo := database.NewSessionRepository(db, logger)
	failedRepo := database.NewFailedLoginRepository(db, logger)
	roleRepo := database.NewRoleRepository(db, logger)

	smtpNotifier := smtp.NewSMTPNotifier(logger)
	wsBroadcaster := ws.NewNotificationBroadcaster(logger)
	notificationService := service.NewNotificationService(notificationRepo, wsBroadcaster, logger)
	notificationService.SetEmailNotifier(smtpNotifier)
	sensorService := service.NewSensorService(sensorRepo, tempRepo, alertRepo, notificationService, logger)

	sensorService.GetAlertService().SetInAppNotificationCallback(func(sensorName, sensorType, reason string, numericValue float64) {
		notif := notifications.Notification{
			Category: notifications.CategoryThresholdAlert,
			Severity: notifications.SeverityWarning,
			Title:    fmt.Sprintf("Alert: %s", sensorName),
			Message:  fmt.Sprintf("%s (value: %.2f)", reason, numericValue),
			Metadata: map[string]interface{}{
				"sensor_name":   sensorName,
				"sensor_type":   sensorType,
				"numeric_value": numericValue,
			},
		}
		notificationService.CreateNotification(context.Background(), notif, "view_alerts")
	})

	tempService := service.NewTemperatureService(tempRepo, logger)
	propertiesService := service.NewPropertiesService(logger)
	cleanupService := service.NewCleanupService(sensorRepo, tempRepo, failedRepo, notificationRepo, logger)

	userService := service.NewUserService(userRepo, notificationService, logger)
	authService := service.NewAuthService(userRepo, sessionRepo, failedRepo, roleRepo, logger)
	roleService := service.NewRoleService(roleRepo, logger)
	alertManagementService := service.NewAlertManagementService(alertRepo, logger)

	apiKeyRepo := database.NewApiKeyRepository(db, logger)
	apiKeyService := service.NewApiKeyService(apiKeyRepo, userRepo, roleRepo, logger)

	api.InitTemperatureAPI(tempService)
	api.InitSensorAPI(sensorService)
	api.InitPropertiesAPI(propertiesService)
	api.InitAuthAPI(authService)
	api.InitUsersAPI(userService)
	api.InitRolesAPI(roleService)
	api.InitAlertAPI(alertManagementService)
	api.InitNotificationsAPI(notificationService)
	api.InitApiKeyAPI(apiKeyService)

	api.InitOAuthAPI(nil)

	middleware.InitAuthMiddleware(authService)
	middleware.InitPermissionMiddleware(roleRepo)
	middleware.InitApiKeyMiddleware(apiKeyService)

	initialAdmin := os.Getenv("SENSOR_HUB_INITIAL_ADMIN")
	if initialAdmin != "" {
		var username, password string
		for i, c := range initialAdmin {
			if c == ':' {
				username = initialAdmin[:i]
				password = initialAdmin[i+1:]
				break
			}
		}
		if username != "" && password != "" {
			err = authService.CreateInitialAdminIfNone(context.Background(), username, password)
			if err != nil {
				return fmt.Errorf("failed to create initial admin user: %w", err)
			}
			logger.Info("initial admin user ready", "username", username)
		}
	}

	err = sensorService.ServiceDiscoverSensors(context.Background())
	if err != nil {
		return fmt.Errorf("failed to discover sensors: %w", err)
	}

	err = oauth.InitialiseOauth()
	if err != nil {
		logger.Warn("failed to initialise OAuth", "error", err)
	}

	oauthAdapter := service.NewOAuthServiceAdapter(oauth.GetService())
	api.InitOAuthAPI(oauthAdapter)

	sensorService.ServiceStartPeriodicSensorCollection(ctx)

	cleanupService.StartPeriodicCleanup(ctx)

	return api.InitialiseAndListen(logger, tel.PrometheusHandler)
}

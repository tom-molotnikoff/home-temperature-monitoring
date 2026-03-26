package cmd

import (
	"database/sql"
	"example/sensorHub/api"
	"example/sensorHub/api/middleware"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/notifications"
	"example/sensorHub/oauth"
	"example/sensorHub/service"
	"example/sensorHub/smtp"
	"example/sensorHub/ws"
	"fmt"
	"io"
	"log"
	"os"

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
	log.SetPrefix("sensor-hub: ")

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", logFile, err)
		}
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	err := appProps.InitialiseConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to initialise application configuration: %w", err)
	}

	db, err := database.InitialiseDatabase()
	if err != nil {
		return fmt.Errorf("failed to initialise database: %w", err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}(db)

	sensorRepo := database.NewSensorRepository(db)
	tempRepo := database.NewTemperatureRepository(db, sensorRepo)
	alertRepo := database.NewAlertRepository(db)
	notificationRepo := database.NewNotificationRepository(db)

	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	failedRepo := database.NewFailedLoginRepository(db)
	roleRepo := database.NewRoleRepository(db)

	smtpNotifier := smtp.NewSMTPNotifier()
	wsBroadcaster := ws.NewNotificationBroadcaster()
	notificationService := service.NewNotificationService(notificationRepo, wsBroadcaster)
	notificationService.SetEmailNotifier(smtpNotifier)
	sensorService := service.NewSensorService(sensorRepo, tempRepo, alertRepo, notificationService)

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
		notificationService.CreateNotification(notif, "view_alerts")
	})

	tempService := service.NewTemperatureService(tempRepo)
	propertiesService := service.NewPropertiesService()
	cleanupService := service.NewCleanupService(sensorRepo, tempRepo, failedRepo, notificationRepo)

	userService := service.NewUserService(userRepo, notificationService)
	authService := service.NewAuthService(userRepo, sessionRepo, failedRepo, roleRepo)
	roleService := service.NewRoleService(roleRepo)
	alertManagementService := service.NewAlertManagementService(alertRepo)

	apiKeyRepo := database.NewApiKeyRepository(db)
	apiKeyService := service.NewApiKeyService(apiKeyRepo, userRepo, roleRepo)

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
			err = authService.CreateInitialAdminIfNone(username, password)
			if err != nil {
				return fmt.Errorf("failed to create initial admin user: %w", err)
			}
			log.Printf("Initial admin user '%s' created (or already exists)", username)
		}
	}

	err = sensorService.ServiceDiscoverSensors()
	if err != nil {
		return fmt.Errorf("failed to discover sensors: %w", err)
	}

	err = oauth.InitialiseOauth()
	if err != nil {
		log.Printf("Failed to initialise OAuth: %v", err)
	}

	oauthAdapter := service.NewOAuthServiceAdapter(oauth.GetService())
	api.InitOAuthAPI(oauthAdapter)

	sensorService.ServiceStartPeriodicSensorCollection()

	cleanupService.StartPeriodicCleanup()

	return api.InitialiseAndListen()
}

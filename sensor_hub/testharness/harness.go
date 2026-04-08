//go:build integration

package testharness

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"example/sensorHub/api"
	"example/sensorHub/api/middleware"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	_ "example/sensorHub/drivers" // register sensor drivers
	"example/sensorHub/notifications"
	"example/sensorHub/service"
	"example/sensorHub/smtp"
	"example/sensorHub/ws"

	"github.com/gin-gonic/gin"
)

// Env holds references to the running test server and its components.
type Env struct {
	ServerURL string
	AdminUser string
	AdminPass string
	DB        *sql.DB
}

const (
	DefaultAdminUser = "testadmin"
	DefaultAdminPass = "testpassword123"
)

// StartServer creates a temp DB, wires up all services, starts the Gin server
// on a random port, and creates an admin user. Cleanup via t.Cleanup.
func StartServer(t interface{ Helper(); Fatalf(string, ...any); Cleanup(func()) }, sensorURLs []string) *Env {
	env, cleanup, err := startServer(sensorURLs)
	if err != nil {
		cleanup()
		if th, ok := t.(interface{ Fatalf(string, ...any) }); ok {
			th.Fatalf("failed to start server: %v", err)
		}
		return nil
	}
	t.Cleanup(cleanup)
	return env
}

// StartServerForMain is like StartServer but for use in TestMain where
// *testing.T is not available. Returns a cleanup function.
func StartServerForMain(sensorURLs []string) (*Env, func(), error) {
	return startServer(sensorURLs)
}

func startServer(sensorURLs []string) (*Env, func(), error) {
	tmpDir, err := os.MkdirTemp("", "sensor-hub-integration-*")
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cleanupDir := func() { os.RemoveAll(tmpDir) }

	dbPath := filepath.Join(tmpDir, "test.db")
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		cleanupDir()
		return nil, func() {}, fmt.Errorf("failed to create config dir: %w", err)
	}

	// Write minimal config files
	appPropsContent := fmt.Sprintf(
		"sensor.collection.interval=300\nsensor.discovery.skip=true\ndatabase.path=%s\nlog.level=debug\nauth.bcrypt.cost=4\n", dbPath)
	writeFileOrErr(filepath.Join(configDir, "application.properties"), appPropsContent)
	writeFileOrErr(filepath.Join(configDir, "database.properties"), fmt.Sprintf("database.path=%s\n", dbPath))
	writeFileOrErr(filepath.Join(configDir, "smtp.properties"), "smtp.user=\n")

	if err := appProps.InitialiseConfig(configDir); err != nil {
		cleanupDir()
		return nil, func() {}, fmt.Errorf("failed to initialise config: %w", err)
	}

	logger := slog.Default()

	db, err := database.InitialiseDatabase(logger)
	if err != nil {
		cleanupDir()
		return nil, func() {}, fmt.Errorf("failed to initialise database: %w", err)
	}

	// Build the full service graph, mirroring cmd/serve.go
	sensorRepo := database.NewSensorRepository(db, logger)
	readingsRepo := database.NewReadingsRepository(db, logger)
	mtRepo := database.NewMeasurementTypeRepository(db, logger)
	alertRepo := database.NewAlertRepository(db, logger)
	notificationRepo := database.NewNotificationRepository(db, logger)
	userRepo := database.NewUserRepository(db, logger)
	sessionRepo := database.NewSessionRepository(db, logger)
	failedRepo := database.NewFailedLoginRepository(db, logger)
	roleRepo := database.NewRoleRepository(db, logger)
	apiKeyRepo := database.NewApiKeyRepository(db, logger)

	smtpNotifier := smtp.NewSMTPNotifier(logger)
	wsBroadcaster := ws.NewNotificationBroadcaster(logger)
	notificationService := service.NewNotificationService(notificationRepo, wsBroadcaster, logger)
	notificationService.SetEmailNotifier(smtpNotifier)

	sensorService := service.NewSensorService(sensorRepo, readingsRepo, mtRepo, alertRepo, notificationService, logger)
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

	readingsService := service.NewReadingsService(readingsRepo, logger)
	propertiesService := service.NewPropertiesService(logger)
	_ = service.NewCleanupService(sensorRepo, readingsRepo, failedRepo, notificationRepo, logger)

	userService := service.NewUserService(userRepo, notificationService, logger)
	authService := service.NewAuthService(userRepo, sessionRepo, failedRepo, roleRepo, logger)
	roleService := service.NewRoleService(roleRepo, logger)
	alertManagementService := service.NewAlertManagementService(alertRepo, logger)
	apiKeyService := service.NewApiKeyService(apiKeyRepo, userRepo, roleRepo, logger)

	// Init API modules (same order as serve.go)
	api.InitReadingsAPI(readingsService)
	api.InitSensorAPI(sensorService)
	api.InitPropertiesAPI(propertiesService)
	api.InitAuthAPI(authService)
	api.InitUsersAPI(userService)
	api.InitRolesAPI(roleService)
	api.InitAlertAPI(alertManagementService)
	api.InitNotificationsAPI(notificationService)
	api.InitApiKeyAPI(apiKeyService)
	api.InitOAuthAPI(nil)

	dashboardRepo := database.NewDashboardRepository(db, logger)
	dashboardService := service.NewDashboardService(dashboardRepo, logger)
	api.InitDashboardAPI(dashboardService)

	// Init middleware
	middleware.InitAuthMiddleware(authService)
	middleware.InitPermissionMiddleware(roleRepo)
	middleware.InitApiKeyMiddleware(apiKeyService)

	// Build Gin router (mirrors api.go without TLS/OTEL/CORS/SPA)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	apiGroup := router.Group("/api")
	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	apiGroup.Use(middleware.CSRFMiddleware())

	api.RegisterAuthRoutes(apiGroup)
	api.RegisterUserRoutes(apiGroup)
	api.RegisterRoleRoutes(apiGroup)
	api.RegisterReadingsRoutes(apiGroup)
	api.RegisterSensorRoutes(apiGroup)
	api.RegisterPropertiesRoutes(apiGroup)
	api.RegisterAlertRoutes(apiGroup)
	api.RegisterOAuthRoutes(apiGroup)
	api.RegisterNotificationRoutes(apiGroup)
	api.RegisterApiKeyRoutes(apiGroup)
	api.RegisterDashboardRoutes(apiGroup)
	api.RegisterDriverRoutes(apiGroup)

	// Start HTTP server on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		db.Close()
		cleanupDir()
		return nil, func() {}, fmt.Errorf("failed to listen: %w", err)
	}
	serverURL := fmt.Sprintf("http://%s", listener.Addr().String())

	srv := &http.Server{Handler: router}
	go srv.Serve(listener)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		db.Close()
		cleanupDir()
	}

	// Create admin user
	if err := authService.CreateInitialAdminIfNone(context.Background(), DefaultAdminUser, DefaultAdminPass); err != nil {
		cleanup()
		return nil, func() {}, fmt.Errorf("failed to create admin user: %w", err)
	}

	return &Env{
		ServerURL: serverURL,
		AdminUser: DefaultAdminUser,
		AdminPass: DefaultAdminPass,
		DB:        db,
	}, cleanup, nil
}

func writeFileOrErr(path, content string) {
	os.WriteFile(path, []byte(content), 0644)
}

package main

import (
	"database/sql"
	"example/sensorHub/api"
	"example/sensorHub/api/middleware"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/oauth"
	"example/sensorHub/service"
	"log"
	"os"
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

	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	failedRepo := database.NewFailedLoginRepository(db)
	roleRepo := database.NewRoleRepository(db)

	sensorService := service.NewSensorService(sensorRepo, tempRepo)
	tempService := service.NewTemperatureService(tempRepo)
	propertiesService := service.NewPropertiesService()
	cleanupService := service.NewCleanupService(sensorRepo, tempRepo, failedRepo)

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, sessionRepo, failedRepo)
	roleService := service.NewRoleService(roleRepo)

	api.InitTemperatureAPI(tempService)
	api.InitSensorAPI(sensorService)
	api.InitPropertiesAPI(propertiesService)
	api.InitAuthAPI(authService)
	api.InitUsersAPI(userService)
	api.InitRolesAPI(roleService)

	// initialize middleware
	middleware.InitAuthMiddleware(authService)
	middleware.InitPermissionMiddleware(roleRepo)

	// bootstrap admin if env provided and no users exist
	initialAdmin := os.Getenv("SENSOR_HUB_INITIAL_ADMIN") // format username:password
	if initialAdmin != "" {
		// split username:password
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
				log.Fatalf("Failed to create initial admin user: %v", err)
			}
			log.Printf("Initial admin user '%s' created (or already exists)", username)
		}
	}

	err = sensorService.ServiceDiscoverSensors()

	if err != nil {
		log.Fatalf("Failed to discover sensors: %v", err)
	}

	err = oauth.InitialiseOauth()
	if err != nil {
		log.Printf("Failed to initialise OAuth: %v", err)
	}
	sensorService.ServiceStartPeriodicSensorCollection()

	cleanupService.StartPeriodicCleanup()

	err = api.InitialiseAndListen()
	if err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

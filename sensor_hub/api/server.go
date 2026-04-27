package api

import "example/sensorHub/service"

// Server holds all service dependencies for the API layer.
type Server struct {
	sensorService       service.SensorServiceInterface
	readingsService     service.ReadingsServiceInterface
	authService         service.AuthServiceInterface
	userService         service.UserServiceInterface
	roleService         service.RoleServiceInterface
	alertService        service.AlertManagementServiceInterface
	notificationService service.NotificationServiceInterface
	apiKeyService       service.ApiKeyServiceInterface
	dashboardService    service.DashboardServiceInterface
	propertiesService   service.PropertiesServiceInterface
	mqttService         service.MQTTServiceInterface
	oauthService        OAuthAPIServiceInterface
	mqttStatsProvider   MQTTStatsProvider
}

// NewServer constructs a Server with all service dependencies.
func NewServer(
	sensorService service.SensorServiceInterface,
	readingsService service.ReadingsServiceInterface,
	authService service.AuthServiceInterface,
	userService service.UserServiceInterface,
	roleService service.RoleServiceInterface,
	alertService service.AlertManagementServiceInterface,
	notificationService service.NotificationServiceInterface,
	apiKeyService service.ApiKeyServiceInterface,
	dashboardService service.DashboardServiceInterface,
	propertiesService service.PropertiesServiceInterface,
	mqttService service.MQTTServiceInterface,
	oauthService OAuthAPIServiceInterface,
	mqttStatsProvider MQTTStatsProvider,
) *Server {
	return &Server{
		sensorService:       sensorService,
		readingsService:     readingsService,
		authService:         authService,
		userService:         userService,
		roleService:         roleService,
		alertService:        alertService,
		notificationService: notificationService,
		apiKeyService:       apiKeyService,
		dashboardService:    dashboardService,
		propertiesService:   propertiesService,
		mqttService:         mqttService,
		oauthService:        oauthService,
		mqttStatsProvider:   mqttStatsProvider,
	}
}

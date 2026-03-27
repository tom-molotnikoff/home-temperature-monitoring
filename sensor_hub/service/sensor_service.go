package service

import (
	"context"
	"encoding/json"
	"errors"
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/alerting"
	database "example/sensorHub/db"
	"example/sensorHub/notifications"
	"example/sensorHub/periodic"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"example/sensorHub/ws"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"gopkg.in/yaml.v3"
)

type AlreadyExistsError struct {
	Message string
}

func NewAlreadyExistsError(message string) *AlreadyExistsError {
	return &AlreadyExistsError{Message: message}
}

func (e *AlreadyExistsError) Error() string {
	return e.Message
}

type SensorService struct {
	sensorRepo   database.SensorRepositoryInterface[types.Sensor]
	tempRepo     database.ReadingsRepository[types.TemperatureReading]
	alertService *alerting.AlertService
	notifSvc     NotificationServiceInterface
	httpClient   *http.Client
	logger       *slog.Logger
}

func NewSensorService(sensorRepo database.SensorRepositoryInterface[types.Sensor], tempRepo database.ReadingsRepository[types.TemperatureReading], alertRepo database.AlertRepository, notifSvc NotificationServiceInterface, logger *slog.Logger) *SensorService {
	alertService := alerting.NewAlertService(alertRepo, logger)
	return &SensorService{
		sensorRepo:   sensorRepo,
		tempRepo:     tempRepo,
		alertService: alertService,
		notifSvc:     notifSvc,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
		logger: logger.With("component", "sensor_service"),
	}
}

func (s *SensorService) notifyConfigEvent(action, sensorName string, metadata map[string]interface{}) {
	if s.notifSvc == nil {
		return
	}
	notif := notifications.Notification{
		Category: notifications.CategoryConfigChange,
		Severity: notifications.SeverityInfo,
		Title:    fmt.Sprintf("Sensor %s", action),
		Message:  fmt.Sprintf("Sensor '%s' was %s", sensorName, action),
		Metadata: metadata,
	}
	go s.notifSvc.CreateNotification(context.Background(), notif, "view_notifications_config")
}

func (s *SensorService) GetAlertService() *alerting.AlertService {
	return s.alertService
}

func (s *SensorService) ServiceAddSensor(ctx context.Context, sensor types.Sensor) error {
	err := s.ServiceValidateSensorConfig(ctx, sensor)
	if err != nil {
		return fmt.Errorf("sensor validation failed: %w", err)
	}

	exists, err := s.sensorRepo.SensorExists(ctx, sensor.Name)
	if err != nil {
		return fmt.Errorf("error checking if sensor exists: %w", err)
	}
	if exists {
		return NewAlreadyExistsError(fmt.Sprintf("sensor with name %s already exists", sensor.Name))
	}
	err = s.sensorRepo.AddSensor(ctx, sensor)
	if err != nil {
		return fmt.Errorf("error adding sensor: %w", err)
	}
	s.logger.Info("sensor added", "name", sensor.Name)
	go s.broadcastSensors(context.Background())
	s.notifyConfigEvent("added", sensor.Name, map[string]interface{}{"sensor_name": sensor.Name})
	return nil
}

func (s *SensorService) ServiceUpdateSensorById(ctx context.Context, sensor types.Sensor) error {
	err := s.ServiceValidateSensorConfig(ctx, sensor)
	if err != nil {
		return fmt.Errorf("sensor validation failed: %w", err)
	}
	err = s.sensorRepo.UpdateSensorById(ctx, sensor)
	if err != nil {
		return fmt.Errorf("error updating sensor: %w", err)
	}
	s.logger.Info("sensor updated", "id", sensor.Id, "name", sensor.Name)
	go s.broadcastSensors(context.Background())
	s.notifyConfigEvent("updated", sensor.Name, map[string]interface{}{"sensor_name": sensor.Name})
	return nil
}

func (s *SensorService) ServiceDeleteSensorByName(ctx context.Context, name string) error {
	exists, err := s.sensorRepo.SensorExists(ctx, name)
	if err != nil {
		return fmt.Errorf("error checking if sensor exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("sensor with name %s does not exist", name)
	}
	err = s.sensorRepo.DeleteSensorByName(ctx, name)
	if err != nil {
		return fmt.Errorf("error deleting sensor: %w", err)
	}
	s.logger.Info("sensor deleted", "name", name)
	go s.broadcastSensors(context.Background())
	s.notifyConfigEvent("removed", name, map[string]interface{}{"sensor_name": name})
	return nil
}

func (s *SensorService) ServiceGetSensorByName(ctx context.Context, name string) (*types.Sensor, error) {
	if name == "" {
		return nil, fmt.Errorf("sensor name cannot be empty")
	}
	sensor, err := s.sensorRepo.GetSensorByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *SensorService) ServiceGetAllSensors(ctx context.Context) ([]types.Sensor, error) {
	sensors, err := s.sensorRepo.GetAllSensors(ctx)
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

func (s *SensorService) ServiceGetSensorsByType(ctx context.Context, sensorType string) ([]types.Sensor, error) {
	sensors, err := s.sensorRepo.GetSensorsByType(ctx, sensorType)
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

func (s *SensorService) ServiceGetSensorIdByName(ctx context.Context, name string) (int, error) {
	return s.sensorRepo.GetSensorIdByName(ctx, name)
}

func (s *SensorService) ServiceSensorExists(ctx context.Context, name string) (bool, error) {
	return s.sensorRepo.SensorExists(ctx, name)
}

func (s *SensorService) ServiceCollectAndStoreAllSensorReadings(ctx context.Context) error {
	return s.ServiceCollectAndStoreTemperatureReadings(ctx)
}

func (s *SensorService) ServiceCollectFromSensorByName(ctx context.Context, sensorName string) error {
	sensor, err := s.ServiceGetSensorByName(ctx, sensorName)
	if err != nil {
		return fmt.Errorf("error retrieving sensor %s: %w", sensorName, err)
	}
	if sensor == nil {
		return fmt.Errorf("sensor %s not found", sensorName)
	}

	if !sensor.Enabled {
		return fmt.Errorf("sensor %s is disabled", sensorName)
	}

	switch sensor.Type {
	case "Temperature":
		reading, err := s.ServiceFetchTemperatureReadingFromSensor(ctx, *sensor)
		if err != nil {
			s.ServiceUpdateSensorHealthById(ctx, sensor.Id, types.SensorBadHealth, fmt.Sprintf("error fetching temperature from sensor: %v", err))
			return fmt.Errorf("error fetching temperature from sensor %s: %w", sensorName, err)
		}
		err = s.tempRepo.Add(ctx, []types.TemperatureReading{reading})
		if err != nil {
			s.ServiceUpdateSensorHealthById(ctx, sensor.Id, types.SensorBadHealth, fmt.Sprintf("error storing temperature reading from sensor: %v", err))
			return fmt.Errorf("error storing temperature reading from sensor %s: %w", sensorName, err)
		}
		s.logger.Debug("collected temperature reading", "sensor", sensorName, "temperature", reading.Temperature)
		ws.BroadcastToTopic("current-temperatures", []types.TemperatureReading{reading})

		// Process alert for this reading
		go func(sensorID int, sensorName string, temp float64) {
			err := s.alertService.ProcessReadingAlert(context.Background(), sensorID, sensorName, "temperature", temp, "")
			if err != nil {
				s.logger.Error("failed to process alert", "sensor", sensorName, "error", err)
			}
		}(sensor.Id, sensorName, reading.Temperature)
	default:
		return fmt.Errorf("unsupported sensor type %s for sensor %s", sensor.Type, sensorName)
	}
	return nil
}

func (s *SensorService) ServiceUpdateSensorHealthById(ctx context.Context, sensorId int, healthStatus types.SensorHealthStatus, healthReason string) {
	err := s.sensorRepo.UpdateSensorHealthById(ctx, sensorId, healthStatus, healthReason)
	if err != nil {
		s.logger.Error("error updating sensor health", "error", err)
		return
	}
	go s.broadcastSensors(context.Background())
}

func (s *SensorService) ServiceCollectReadingToValidateSensor(ctx context.Context, sensor types.Sensor) error {
	switch sensor.Type {
	case "Temperature":
		_, err := s.ServiceFetchTemperatureReadingFromSensor(ctx, sensor)
		if err != nil {
			return fmt.Errorf("error fetching temperature from sensor %s: %w", sensor.Name, err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported sensor type %s for sensor %s", sensor.Type, sensor.Name)
	}
}

func (s *SensorService) ServiceCollectAndStoreTemperatureReadings(ctx context.Context) error {
	readings, err := s.ServiceFetchAllTemperatureReadings(ctx)
	if err != nil {
		return fmt.Errorf("error fetching temperature readings: %w", err)
	}

	for _, reading := range readings {
		err = s.tempRepo.Add(ctx, []types.TemperatureReading{reading})
		if err != nil {
			s.logger.Error("error storing temperature reading", "sensor", reading.SensorName, "error", err)
			continue
		}
		s.logger.Debug("collected temperature reading", "sensor", reading.SensorName, "temperature", reading.Temperature)

		sensor, err := s.sensorRepo.GetSensorByName(ctx, reading.SensorName)
		if err != nil {
			s.logger.Error("failed to get sensor for alert processing", "error", err)
		} else if sensor != nil {
			go func(sensorID int, sensorName string, temp float64) {
				err := s.alertService.ProcessReadingAlert(context.Background(), sensorID, sensorName, "temperature", temp, "")
				if err != nil {
					s.logger.Error("failed to process alert", "sensor", sensorName, "error", err)
				}
			}(sensor.Id, reading.SensorName, reading.Temperature)
		}
	}
	ws.BroadcastToTopic("current-temperatures", readings)

	return nil
}

func (s *SensorService) ServiceFetchAllTemperatureReadings(ctx context.Context) ([]types.TemperatureReading, error) {
	sensors, err := s.ServiceGetSensorsByType(ctx, "temperature")
	if err != nil {
		return nil, fmt.Errorf("error fetching sensors of type 'temperature': %w", err)
	}

	var allReadings []types.TemperatureReading
	for _, sensor := range sensors {
		if !sensor.Enabled {
			s.logger.Debug("skipping disabled sensor", "name", sensor.Name, "url", sensor.URL)
			continue
		}
		reading, err := s.ServiceFetchTemperatureReadingFromSensor(ctx, sensor)
		if err != nil {
			s.ServiceUpdateSensorHealthById(ctx, sensor.Id, types.SensorBadHealth, fmt.Sprintf("error fetching temperature from sensor: %v", err))
			s.logger.Error("error fetching temperature from sensor", "name", sensor.Name, "url", sensor.URL, "error", err)
			continue
		}
		allReadings = append(allReadings, reading)
	}
	return allReadings, nil
}

func (s *SensorService) ServiceFetchTemperatureReadingFromSensor(ctx context.Context, sensor types.Sensor) (types.TemperatureReading, error) {
	rawTempReading := types.RawTempReading{}
	tempReading := types.TemperatureReading{}
	req, err := http.NewRequestWithContext(ctx, "GET", sensor.URL+"/temperature", nil)
	if err != nil {
		return tempReading, fmt.Errorf("error creating request to sensor at %s: %w", sensor.URL, err)
	}
	response, err := s.httpClient.Do(req)
	if err != nil {
		return tempReading, fmt.Errorf("error making GET request to sensor at %s: %w", sensor.URL, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return tempReading, fmt.Errorf("received non-200 response from sensor at %s: %d", sensor.URL, response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(&rawTempReading)
	if err != nil {
		return tempReading, fmt.Errorf("error decoding JSON response from sensor at %s: %w", sensor.URL, err)
	}
	rawTempReading.Time = utils.NormalizeTimeToSpaceFormat(rawTempReading.Time)
	tempReading = types.TemperatureReading{
		SensorName:  sensor.Name,
		Time:        rawTempReading.Time,
		Temperature: rawTempReading.Temperature,
	}
	s.ServiceUpdateSensorHealthById(ctx, sensor.Id, types.SensorGoodHealth, "successful reading")
	return tempReading, nil
}

func (s *SensorService) ServiceDiscoverSensors(ctx context.Context) error {
	shouldSkipDiscovery := appProps.AppConfig.SensorDiscoverySkip

	if shouldSkipDiscovery {
		s.logger.Info("skipping sensor discovery as per configuration")
		return nil
	}

	fileData, err := os.ReadFile(appProps.AppConfig.OpenAPILocation)
	if err != nil {
		return fmt.Errorf("cannot find the openapi.yaml file for the temperature sensors: %w", err)
	}
	var servers types.SensorServers

	err = yaml.Unmarshal(fileData, &servers)
	if err != nil {
		return fmt.Errorf("cannot unmarshal the yaml into a map: %w", err)
	}

	for _, value := range servers.Servers {
		sensorName := value.Variables["sensor_name"].Default
		url := value.Url
		sensorType := value.Variables["sensor_type"].Default

		sensor := types.Sensor{
			Name: sensorName,
			Type: sensorType,
			URL:  url,
		}
		err = s.ServiceAddSensor(ctx, sensor)
		if err != nil {
			s.logger.Warn("error adding sensor during discovery", "sensor", sensorName, "error", err)
			var alreadyExistsErr *AlreadyExistsError
			if errors.As(err, &alreadyExistsErr) {
				s.logger.Info("sensor already exists, updating", "sensor", sensorName)
				err = s.ServiceUpdateSensorById(ctx, sensor)
				if err != nil {
					s.logger.Error("error updating sensor during discovery", "sensor", sensorName, "error", err)
				} else {
					s.logger.Info("sensor updated during discovery", "sensor", sensor.Name)
				}
			}
			continue
		}
		s.logger.Info("sensor discovered and added", "sensor", sensor.Name)
	}
	return nil
}

func (s *SensorService) ServiceStartPeriodicSensorCollection(ctx context.Context) {
	intervalSec := appProps.AppConfig.SensorCollectionInterval

	periodic.RunTask(ctx, periodic.TaskConfig{
		Name:           "sensor_collection",
		Interval:       time.Duration(intervalSec) * time.Second,
		Logger:         s.logger,
		RunImmediately: true,
	}, func(ctx context.Context) error {
		return s.ServiceCollectAndStoreAllSensorReadings(ctx)
	})
}

func (s *SensorService) ServiceSetEnabledSensorByName(ctx context.Context, name string, enabled bool) error {
	exists, err := s.sensorRepo.SensorExists(ctx, name)
	if err != nil {
		return fmt.Errorf("error checking if sensor exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("sensor with name %s does not exist", name)
	}
	err = s.sensorRepo.SetEnabledSensorByName(ctx, name, enabled)
	if err != nil {
		return fmt.Errorf("error setting enabled status for sensor: %w", err)
	}
	s.logger.Info("sensor enabled status changed", "name", name, "enabled", enabled)
	go s.broadcastSensors(context.Background())
	if enabled {
		go func() {
			err := s.ServiceCollectFromSensorByName(context.Background(), name)
			if err != nil {
				s.logger.Error("error collecting initial reading from enabled sensor", "name", name, "error", err)
			}
		}()
	}
	return nil
}

func (s *SensorService) ServiceGetTotalReadingsForEachSensor(ctx context.Context) (map[string]int, error) {
	sensors, err := s.sensorRepo.GetAllSensors(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all sensors: %w", err)
	}

	totalReadings := make(map[string]int)
	for _, sensor := range sensors {
		count := 0
		switch sensor.Type {
		case "Temperature":
			count, err = s.tempRepo.GetTotalReadingsBySensorId(ctx, sensor.Id)
			if err != nil {
				return nil, fmt.Errorf("error retrieving total readings for sensor %s: %w", sensor.Name, err)
			}
		default:
			continue
		}

		totalReadings[sensor.Name] = count
	}
	return totalReadings, nil
}

func (s *SensorService) ServiceGetSensorHealthHistoryByName(ctx context.Context, name string, limit int) ([]types.SensorHealthHistory, error) {
	sensorId, err := s.ServiceGetSensorIdByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving sensor ID for sensor %s: %w", name, err)
	}

	history, err := s.sensorRepo.GetSensorHealthHistoryById(ctx, sensorId, limit)
	if err != nil {
		return nil, fmt.Errorf("error retrieving health history for sensor %s: %w", name, err)
	}
	return history, nil
}

func (s *SensorService) ServiceValidateSensorConfig(ctx context.Context, sensor types.Sensor) error {
	if sensor.Name == "" || sensor.Type == "" || sensor.URL == "" {
		return fmt.Errorf("sensor name, type, and URL cannot be empty")
	}
	err := s.ServiceCollectReadingToValidateSensor(ctx, sensor)
	if err != nil {
		return fmt.Errorf("invalid sensor, failed to collect a reading: %w", err)
	}
	return nil
}

func (s *SensorService) broadcastSensors(ctx context.Context) {
	sensors, err := s.sensorRepo.GetAllSensors(ctx)
	if err != nil {
		s.logger.Error("failed to fetch sensors for broadcast", "error", err)
		return
	}

	byType := make(map[string][]types.Sensor)
	for _, sensor := range sensors {
		byType[sensor.Type] = append(byType[sensor.Type], sensor)
	}
	for t, list := range byType {
		topic := "sensors:" + t
		ws.BroadcastToTopic(topic, list)
	}
}

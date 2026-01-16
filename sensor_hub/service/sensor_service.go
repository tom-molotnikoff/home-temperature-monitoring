package service

import (
	"encoding/json"
	"errors"
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/alerting"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"example/sensorHub/ws"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
}

func NewSensorService(sensorRepo database.SensorRepositoryInterface[types.Sensor], tempRepo database.ReadingsRepository[types.TemperatureReading], alertRepo database.AlertRepository, notifier alerting.Notifier) *SensorService {
	alertService := alerting.NewAlertService(alertRepo, notifier)
	return &SensorService{
		sensorRepo:   sensorRepo,
		tempRepo:     tempRepo,
		alertService: alertService,
	}
}

func (s *SensorService) ServiceAddSensor(sensor types.Sensor) error {
	err := s.ServiceValidateSensorConfig(sensor)
	if err != nil {
		return fmt.Errorf("sensor validation failed: %w", err)
	}

	exists, err := s.sensorRepo.SensorExists(sensor.Name)
	if err != nil {
		return fmt.Errorf("error checking if sensor exists: %w", err)
	}
	if exists {
		return NewAlreadyExistsError(fmt.Sprintf("sensor with name %s already exists", sensor.Name))
	}
	err = s.sensorRepo.AddSensor(sensor)
	if err != nil {
		return fmt.Errorf("error adding sensor: %w", err)
	}
	log.Printf("Added sensor: %v", sensor)
	go s.broadcastSensors()
	return nil
}

func (s *SensorService) ServiceUpdateSensorById(sensor types.Sensor) error {
	err := s.ServiceValidateSensorConfig(sensor)
	if err != nil {
		return fmt.Errorf("sensor validation failed: %w", err)
	}
	err = s.sensorRepo.UpdateSensorById(sensor)
	if err != nil {
		return fmt.Errorf("error updating sensor: %w", err)
	}
	log.Printf("Updated sensor with id %v to: %v", sensor.Id, sensor)
	go s.broadcastSensors()
	return nil
}

func (s *SensorService) ServiceDeleteSensorByName(name string) error {
	exists, err := s.sensorRepo.SensorExists(name)
	if err != nil {
		return fmt.Errorf("error checking if sensor exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("sensor with name %s does not exist", name)
	}
	err = s.sensorRepo.DeleteSensorByName(name)
	if err != nil {
		return fmt.Errorf("error deleting sensor: %w", err)
	}
	log.Printf("Deleted sensor with name %s", name)
	go s.broadcastSensors()

	return nil
}

func (s *SensorService) ServiceGetSensorByName(name string) (*types.Sensor, error) {
	if name == "" {
		return nil, fmt.Errorf("sensor name cannot be empty")
	}
	sensor, err := s.sensorRepo.GetSensorByName(name)
	if err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *SensorService) ServiceGetAllSensors() ([]types.Sensor, error) {
	sensors, err := s.sensorRepo.GetAllSensors()
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

func (s *SensorService) ServiceGetSensorsByType(sensorType string) ([]types.Sensor, error) {
	sensors, err := s.sensorRepo.GetSensorsByType(sensorType)
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

func (s *SensorService) ServiceGetSensorIdByName(name string) (int, error) {
	return s.sensorRepo.GetSensorIdByName(name)
}

func (s *SensorService) ServiceSensorExists(name string) (bool, error) {
	return s.sensorRepo.SensorExists(name)
}

func (s *SensorService) ServiceCollectAndStoreAllSensorReadings() error {
	return s.ServiceCollectAndStoreTemperatureReadings()
}

func (s *SensorService) ServiceCollectFromSensorByName(sensorName string) error {
	sensor, err := s.ServiceGetSensorByName(sensorName)
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
		reading, err := s.ServiceFetchTemperatureReadingFromSensor(*sensor)
		if err != nil {
			s.ServiceUpdateSensorHealthById(sensor.Id, types.SensorBadHealth, fmt.Sprintf("error fetching temperature from sensor: %v", err))
			return fmt.Errorf("error fetching temperature from sensor %s: %w", sensorName, err)
		}
		err = s.tempRepo.Add([]types.TemperatureReading{reading})
		if err != nil {
			s.ServiceUpdateSensorHealthById(sensor.Id, types.SensorBadHealth, fmt.Sprintf("error storing temperature reading from sensor: %v", err))
			return fmt.Errorf("error storing temperature reading from sensor %s: %w", sensorName, err)
		}
		log.Printf("Collected temperature reading from sensor %s: %v", sensorName, reading)
		ws.BroadcastToTopic("current-temperatures", []types.TemperatureReading{reading})

		// Process alert for this reading
		go func(sensorID int, sensorName string, temp float64) {
			err := s.alertService.ProcessReadingAlert(sensorID, sensorName, "temperature", temp, "")
			if err != nil {
				log.Printf("Failed to process alert for sensor %s: %v", sensorName, err)
			}
		}(sensor.Id, sensorName, reading.Temperature)
	default:
		return fmt.Errorf("unsupported sensor type %s for sensor %s", sensor.Type, sensorName)
	}
	return nil
}

func (s *SensorService) ServiceUpdateSensorHealthById(sensorId int, healthStatus types.SensorHealthStatus, healthReason string) {
	err := s.sensorRepo.UpdateSensorHealthById(sensorId, healthStatus, healthReason)
	if err != nil {
		log.Printf("error updating sensor health: %v", err)
		return
	}
	go s.broadcastSensors()
}

func (s *SensorService) ServiceCollectReadingToValidateSensor(sensor types.Sensor) error {
	switch sensor.Type {
	case "Temperature":
		_, err := s.ServiceFetchTemperatureReadingFromSensor(sensor)
		if err != nil {
			return fmt.Errorf("error fetching temperature from sensor %s: %w", sensor.Name, err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported sensor type %s for sensor %s", sensor.Type, sensor.Name)
	}
}

func (s *SensorService) ServiceCollectAndStoreTemperatureReadings() error {
	readings, err := s.ServiceFetchAllTemperatureReadings()
	if err != nil {
		return fmt.Errorf("error fetching temperature readings: %w", err)
	}

	for _, reading := range readings {
		err = s.tempRepo.Add([]types.TemperatureReading{reading})
		if err != nil {
			log.Printf("Error storing temperature reading %v: %v", reading, err)
			continue
		}
		log.Printf("Collected temperature reading: %v", reading)

		sensor, err := s.sensorRepo.GetSensorByName(reading.SensorName)
		if err != nil {
			log.Printf("Failed to get sensor for alert processing: %v", err)
		} else if sensor != nil {
			go func(sensorID int, sensorName string, temp float64) {
				err := s.alertService.ProcessReadingAlert(sensorID, sensorName, "temperature", temp, "")
				if err != nil {
					log.Printf("Failed to process alert for sensor %s: %v", sensorName, err)
				}
			}(sensor.Id, reading.SensorName, reading.Temperature)
		}
	}
	ws.BroadcastToTopic("current-temperatures", readings)

	return nil
}

func (s *SensorService) ServiceFetchAllTemperatureReadings() ([]types.TemperatureReading, error) {
	sensors, err := s.ServiceGetSensorsByType("temperature")
	if err != nil {
		return nil, fmt.Errorf("error fetching sensors of type 'temperature': %w", err)
	}

	var allReadings []types.TemperatureReading
	for _, sensor := range sensors {
		if !sensor.Enabled {
			log.Printf("Skipping disabled sensor %s at %s", sensor.Name, sensor.URL)
			continue
		}
		reading, err := s.ServiceFetchTemperatureReadingFromSensor(sensor)
		if err != nil {
			s.ServiceUpdateSensorHealthById(sensor.Id, types.SensorBadHealth, fmt.Sprintf("error fetching temperature from sensor: %v", err))
			log.Printf("Error fetching temperature from sensor %s at %s: %v", sensor.Name, sensor.URL, err)
			continue
		}
		allReadings = append(allReadings, reading)
	}
	return allReadings, nil
}

func (s *SensorService) ServiceFetchTemperatureReadingFromSensor(sensor types.Sensor) (types.TemperatureReading, error) {
	rawTempReading := types.RawTempReading{}
	tempReading := types.TemperatureReading{}
	response, err := http.Get(sensor.URL + "/temperature")
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
	s.ServiceUpdateSensorHealthById(sensor.Id, types.SensorGoodHealth, "successful reading")
	return tempReading, nil
}

func (s *SensorService) ServiceDiscoverSensors() error {
	shouldSkipDiscovery := appProps.AppConfig.SensorDiscoverySkip

	if shouldSkipDiscovery {
		log.Printf("Skipping sensor discovery as per configuration")
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
		err = s.ServiceAddSensor(sensor)
		if err != nil {
			log.Printf("Error adding sensor %s: %v", sensorName, err)
			var alreadyExistsErr *AlreadyExistsError
			if errors.As(err, &alreadyExistsErr) {
				log.Printf("Sensor %s already exists, updating instead", sensorName)
				err = s.ServiceUpdateSensorById(sensor)
				if err != nil {
					log.Printf("Error updating sensor %s: %v", sensorName, err)
				} else {
					log.Printf("Updated sensor: %v", sensor)
				}
			}
			continue
		}
		log.Printf("Discovered and added sensor: %v", sensor)
	}
	return nil
}

func (s *SensorService) ServiceStartPeriodicSensorCollection() {
	intervalSec := appProps.AppConfig.SensorCollectionInterval

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()
		for {
			log.Printf("Starting periodic sensor readings collection")
			err := s.ServiceCollectAndStoreAllSensorReadings()
			if err != nil {
				log.Printf("Error taking periodic readings from sensors, skipping: %v", err)
			}
			<-ticker.C
		}
	}()
}

func (s *SensorService) ServiceSetEnabledSensorByName(name string, enabled bool) error {
	exists, err := s.sensorRepo.SensorExists(name)
	if err != nil {
		return fmt.Errorf("error checking if sensor exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("sensor with name %s does not exist", name)
	}
	err = s.sensorRepo.SetEnabledSensorByName(name, enabled)
	if err != nil {
		return fmt.Errorf("error setting enabled status for sensor: %w", err)
	}
	log.Printf("Set enabled status for sensor %s to %v", name, enabled)
	go s.broadcastSensors()
	if enabled {
		go func() {
			err := s.ServiceCollectFromSensorByName(name)
			if err != nil {
				log.Printf("Error collecting initial reading from enabled sensor %s: %v", name, err)
			}
		}()
	}
	return nil
}

func (s *SensorService) ServiceGetTotalReadingsForEachSensor() (map[string]int, error) {
	sensors, err := s.sensorRepo.GetAllSensors()
	if err != nil {
		return nil, fmt.Errorf("error retrieving all sensors: %w", err)
	}

	totalReadings := make(map[string]int)
	for _, sensor := range sensors {
		count := 0
		switch sensor.Type {
		case "Temperature":
			count, err = s.tempRepo.GetTotalReadingsBySensorId(sensor.Id)
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

func (s *SensorService) ServiceGetSensorHealthHistoryByName(name string, limit int) ([]types.SensorHealthHistory, error) {
	sensorId, err := s.ServiceGetSensorIdByName(name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving sensor ID for sensor %s: %w", name, err)
	}

	history, err := s.sensorRepo.GetSensorHealthHistoryById(sensorId, limit)
	if err != nil {
		return nil, fmt.Errorf("error retrieving health history for sensor %s: %w", name, err)
	}
	return history, nil
}

func (s *SensorService) ServiceValidateSensorConfig(sensor types.Sensor) error {
	if sensor.Name == "" || sensor.Type == "" || sensor.URL == "" {
		return fmt.Errorf("sensor name, type, and URL cannot be empty")
	}
	err := s.ServiceCollectReadingToValidateSensor(sensor)
	if err != nil {
		return fmt.Errorf("invalid sensor, failed to collect a reading: %w", err)
	}
	return nil
}

func (s *SensorService) broadcastSensors() {
	sensors, err := s.sensorRepo.GetAllSensors()
	if err != nil {
		log.Printf("ws: failed to fetch sensors for broadcast: %v", err)
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

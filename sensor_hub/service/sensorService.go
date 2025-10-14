package service

import (
	"encoding/json"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/smtp"
	"example/sensorHub/types"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type SensorService struct {
	sensorRepo database.SensorRepositoryInterface[types.Sensor]
	tempRepo   database.ReadingsRepository[types.TemperatureReading]
}

func NewSensorService(sensorRepo database.SensorRepositoryInterface[types.Sensor], tempRepo database.ReadingsRepository[types.TemperatureReading]) *SensorService {
	return &SensorService{
		sensorRepo: sensorRepo,
		tempRepo:   tempRepo,
	}
}
func (s *SensorService) ServiceAddSensor(sensor types.Sensor) error {
	return s.sensorRepo.AddSensor(sensor)
}

func (s *SensorService) ServiceUpdateSensorByName(sensor types.Sensor) error {
	return s.sensorRepo.UpdateSensorByName(sensor)
}

func (s *SensorService) ServiceDeleteSensorByName(name string) error {
	return s.sensorRepo.DeleteSensorByName(name)
}

func (s *SensorService) ServiceGetSensorByName(name string) (*types.Sensor, error) {
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
	return s.CollectAndStoreTemperatureReadings()
}

func (s *SensorService) ServiceCollectFromSensorByName(sensorName string) error {
	sensor, err := s.ServiceGetSensorByName(sensorName)
	if err != nil {
		return fmt.Errorf("error retrieving sensor %s: %w", sensorName, err)
	}
	if sensor == nil {
		return fmt.Errorf("sensor %s not found", sensorName)
	}
	switch sensor.Type {
	case "Temperature":
		reading, err := s.FetchTemperatureReadingFromSensor(*sensor)
		if err != nil {
			return fmt.Errorf("error fetching temperature from sensor %s: %w", sensorName, err)
		}
		err = s.tempRepo.Add([]types.TemperatureReading{reading})
		if err != nil {
			return fmt.Errorf("error storing temperature reading from sensor %s: %w", sensorName, err)
		}
		log.Printf("Collected temperature reading from sensor %s: %v", sensorName, reading)
	default:
		return fmt.Errorf("unsupported sensor type %s for sensor %s", sensor.Type, sensorName)
	}
	return nil
}

func (s *SensorService) CollectAndStoreTemperatureReadings() error {
	readings, err := s.FetchAllTemperatureReadings()
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

	}

	err = smtp.SendAlertEmailIfNeeded(readings)
	if err != nil {
		log.Printf("Failed to send alerts: %v", err)
	}

	return nil
}

func (s *SensorService) FetchAllTemperatureReadings() ([]types.TemperatureReading, error) {
	sensors, err := s.ServiceGetSensorsByType("temperature")
	if err != nil {
		return nil, fmt.Errorf("error fetching sensors of type 'temperature': %w", err)
	}

	var allReadings []types.TemperatureReading
	for _, sensor := range sensors {
		reading, err := s.FetchTemperatureReadingFromSensor(sensor)
		if err != nil {
			log.Printf("Error fetching temperature from sensor %s at %s: %v", sensor.Name, sensor.URL, err)
			continue
		}
		allReadings = append(allReadings, reading)
	}
	return allReadings, nil
}

func (s *SensorService) FetchTemperatureReadingFromSensor(sensor types.Sensor) (types.TemperatureReading, error) {
	rawTempReading := types.RawTempReading{}
	tempReading := types.TemperatureReading{}
	response, err := http.Get(sensor.URL + "/temperature")
	if err != nil {
		return tempReading, fmt.Errorf("error making GET request to sensor at %s: %w", sensor.URL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return tempReading, fmt.Errorf("received non-200 response from sensor at %s: %d", sensor.URL, response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(&rawTempReading)
	if err != nil {
		return tempReading, fmt.Errorf("error decoding JSON response from sensor at %s: %w", sensor.URL, err)
	}
	tempReading = types.TemperatureReading{
		SensorName:  sensor.Name,
		Time:        rawTempReading.Time,
		Temperature: rawTempReading.Temperature,
	}

	return tempReading, nil
}

func (s *SensorService) DiscoverSensors() error {
	fileData, err := os.ReadFile(appProps.APPLICATION_PROPERTIES["openapi.yaml.location"])
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
			continue
		}
		log.Printf("Discovered and added sensor: %v", sensor)
	}
	return nil
}

func (s *SensorService) StartPeriodicSensorCollection() {
	intervalStr := appProps.APPLICATION_PROPERTIES["sensor.collection.interval"]
	intervalSec, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("Invalid sensor.collection.interval value: %s, defaulting to 60 seconds", intervalStr)
		intervalSec = 60
	}
	go func() {
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()
		for {
			err := s.ServiceCollectAndStoreAllSensorReadings()
			if err != nil {
				log.Printf("Error taking periodic readings from sensors, skipping: %v", err)
			}
			<-ticker.C
		}
	}()
}

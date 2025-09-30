package sensors

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/types"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

var sensors []ISensor

// Exported function to get reading from a specific temperature sensor by name
// It returns the APIReading for the specified sensor if found and successfully read.
// If the sensor name is empty or not found, or if there is an error taking the reading,
// it returns an error.
var GetReadingFromTemperatureSensor = func(sensorName string) (*types.APIReading, error) {
	if sensorName == "" {
		return nil, fmt.Errorf("sensor name cannot be empty")
	}

	var foundSensor ISensor
	for _, sensor := range sensors {
		if sensor.GetName() == sensorName {
			foundSensor = sensor
			break
		}
	}
	if foundSensor == nil {
		return nil, fmt.Errorf("sensor name %s not found", sensorName)
	}

	tempSensor, ok := foundSensor.(*TemperatureSensor)
	if !ok {
		return nil, fmt.Errorf("sensor %s is not a TemperatureSensor", sensorName)
	}

	if err := tempSensor.TakeReading(true); err != nil {
		return nil, fmt.Errorf("error taking reading from sensor %s: %w", tempSensor.GetName(), err)
	}
	return tempSensor.GetLatestReading(), nil
}

// Exported function to get readings from all temperature sensors
// It returns a slice of APIReading for all sensors that successfully provided a reading.
// If no sensors are discovered or no readings are collected, it returns an error.
var GetReadingFromAllTemperatureSensors = func() ([]types.APIReading, error) {
	if len(sensors) == 0 {
		return nil, fmt.Errorf("no sensors discovered")
	}
	var readings []types.APIReading
	for _, sensor := range sensors {
		tempSensor, ok := sensor.(*TemperatureSensor)
		if !ok {
			continue
		}
		if err := tempSensor.TakeReading(true); err != nil {
			log.Printf("Error taking reading from sensor %s: %v", tempSensor.GetName(), err)
			continue
		}
		reading := tempSensor.GetLatestReading()
		if reading != nil {
			readings = append(readings, *reading)
		}
	}
	if len(readings) == 0 {
		return nil, fmt.Errorf("no readings collected from any sensors")
	}
	return readings, nil
}

// This function reads the OpenAPI specification file (openapi.yaml) to build an object of sensors.
// It expects the file to be in the same directory as the executable or at the specified relative path.
// It will log a fatal error if it cannot read the file or parse it correctly.
func DiscoverSensors() error {
	fileData, err := os.ReadFile(appProps.APPLICATION_PROPERTIES["openapi.yaml.location"])
	if err != nil {
		return fmt.Errorf("cannot find the openapi.yaml file for the temperature sensors: %w", err)
	}
	var servers types.SensorServers

	err = yaml.Unmarshal(fileData, &servers)
	if err != nil {
		return fmt.Errorf("cannot unmarshal the yaml into a map: %w", err)
	}
	sensors = make([]ISensor, 0)

	for _, value := range servers.Servers {
		sensorName := value.Variables["sensor_name"].Default
		url := value.Url
		sensorType := value.Variables["sensor_type"].Default
		switch sensorType {
		case "Temperature":
			sensors = append(sensors, NewTemperatureSensor(sensorName, url))
		default:
			log.Printf("Unknown sensor type %s for sensor %s, skipping...", sensorType, sensorName)
			continue
		}
	}
	log.Printf("Discovered sensors:")
	for _, sensor := range sensors {
		log.Printf(" - %s", sensor.ToString())
	}
	return nil
}

// This function fetches the readings from all known sensors.
// If at least one sensor provides readings successfully, it returns nil.
// If no readings are successfully collected, it returns an error.
var takeReadingsFromAllSensors = func() error {
	if len(sensors) == 0 {
		log.Println("No sensors discovered, cannot take readings.")
		return fmt.Errorf("no sensors discovered")
	}
	count := 0
	for _, sensor := range sensors {
		log.Printf("Taking reading from sensor: %s", sensor.GetName())
		err := sensor.TakeReading(true)
		if err != nil {
			log.Printf("Error taking reading from sensor %s: %v", sensor.GetName(), err)
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("no readings collected from any sensors")
	}
	return nil
}

// This function starts a goroutine that periodically collects readings from all sensors - regardless of type -
// at an interval defined in the application properties (sensor.collection.interval).
// It uses a ticker to trigger the collection at the specified interval.
// Any errors encountered during the collection process are logged but do not stop the periodic collection.
func StartPeriodicSensorCollection() {
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
			err := takeReadingsFromAllSensors()
			if err != nil {
				log.Printf("Error taking periodic readings from sensors, skipping: %v", err)
			}
			<-ticker.C
		}
	}()
}

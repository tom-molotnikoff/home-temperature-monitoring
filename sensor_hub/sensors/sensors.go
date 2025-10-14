package sensors

import (
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

var sensors []ISensor

var GetReadingFromTemperatureSensor = func(sensorName string) (*types.APITempReading, error) {
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

var GetReadingFromAllTemperatureSensors = func() ([]types.APITempReading, error) {
	if len(sensors) == 0 {
		return nil, fmt.Errorf("no sensors discovered")
	}
	var readings []types.APITempReading
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

func DiscoverSensors(repos map[string]interface{}) error {
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

		repo, ok := repos[sensorType]
		if !ok {
			log.Printf("No repository found for sensor type %s, skipping...", sensorType)
			continue
		}

		switch sensorType {
		case "Temperature":
			tempRepo, ok := repo.(database.Repository[types.DbTempReading])
			if !ok {
				log.Printf("Repository for Temperature sensor is of wrong type, skipping...")
				continue
			}
			sensors = append(sensors, NewTemperatureSensor(sensorName, url, tempRepo))
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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Servers struct {
	Servers []ServerItem `yaml:"servers"`
}

type ServerItem struct {
	Url       string                      `yaml:"url"`
	Variables map[string]VariableProperty `yaml:"variables"`
}

type VariableProperty struct {
	Default string `yaml:"default"`
}

type SensorReading struct {
	SensorName string             `json:"sensor_name"`
	Reading    TemperatureReading `json:"reading"`
}

type TemperatureReading struct {
	Temperature float64 `json:"temperature"`
	Time        string  `json:"time"`
}

var sensorUrls map[string]string

// This function reads the OpenAPI specification file (openapi.yaml) to discover the URLs of temperature sensors.
// It expects the file to be in the same directory as the executable or at the specified relative path.
// It will log a fatal error if it cannot read the file or parse it correctly.
func discover_sensor_urls() error {
	fileData, err := os.ReadFile(APPLICATION_PROPERTIES["openapi.yaml.location"])
	if err != nil {
		log.Printf("Cannot find the openapi.yaml file for the temperature sensors: %s\n", err)
		return err
	}
	var servers Servers

	err = yaml.Unmarshal(fileData, &servers)
	if err != nil {
		log.Printf("Cannot unmarshal the yaml into a map: %s\n", err)
		return err
	}
	urls := make(map[string]string, 0)

	for _, value := range servers.Servers {
		for _, variable := range value.Variables {
			urls[variable.Default] = value.Url
		}
	}
	sensorUrls = urls
	log.Printf("Discovered sensor URLs: %v\n", sensorUrls)
	return nil
}

var take_reading_from_named_sensor = func(sensorName string) (*SensorReading, error) {
	if sensorName == "" {
		log.Println("Sensor name is empty, cannot fetch reading.")
		return nil, fmt.Errorf("sensor name cannot be empty")
	}
	if _, exists := sensorUrls[sensorName]; !exists {
		log.Printf("Sensor name %s not found in discovered sensor URLs.\n", sensorName)
		return nil, fmt.Errorf("sensor name %s not found", sensorName)
	}

	readingUrl := sensorUrls[sensorName] + "/temperature"

	resp, err := http.Get(readingUrl)
	if err != nil {
		log.Printf("Issue fetching temperature from sensor %s: %s\n", sensorName, err)
		return nil, err
	}
	defer resp.Body.Close()

	response := new(SensorReading)
	err = json.NewDecoder(resp.Body).Decode(response)

	if err != nil {
		log.Printf("Issue reading request body from sensor %s: %s\n", sensorName, err)
		return nil, err
	}
	// insert into database
	readings := make([]*SensorReading, 0)
	readings = append(readings, response)
	err = add_list_of_readings(readings)
	if err != nil {
		log.Printf("Issue persisting readings to database: %s\n", err)
		return nil, err
	}
	err = sendAlertEmailIfNeeded(readings)
	if err != nil {
		log.Printf("Failed to send alerts: %s", err)
	}

	return response, nil
}

// This function takes a list of sensor URLs, fetches the temperature readings from each sensor,
// and returns a slice of SensorReading objects containing the sensor name and its reading.
// It logs any errors encountered during the fetching or decoding process.
// It assumes that the sensor API returns a JSON object with the structure:
//
//	{
//	  "temperature": <float>,
//	  "time": <string>
//	}
var take_readings = func() ([]*SensorReading, error) {
	responses := make([]*SensorReading, 0)
	for _, url := range sensorUrls {
		readingUrl := url + "temperature"
		resp, err := http.Get(readingUrl)
		if err != nil {
			log.Printf("Issue fetching temperature from a sensor: %s\n", err)
			continue
		}
		defer resp.Body.Close()
		response := new(SensorReading)
		err = json.NewDecoder(resp.Body).Decode(response)

		if err != nil {
			log.Printf("Issue reading request body: %s\n", err)
			continue
		}

		responses = append(responses, response)
	}
	err := add_list_of_readings(responses)
	if err != nil {
		log.Printf("Issue persisting readings to database: %s\n", err)
		return nil, err
	}
	err = sendAlertEmailIfNeeded(responses)
	if err != nil {
		log.Printf("Failed to send alerts: %s", err)
	}
	return responses, err
}

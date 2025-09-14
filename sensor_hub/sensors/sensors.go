package sensors

import (
	"encoding/json"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/smtp"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

var sensorUrls map[string]string

// This function reads the OpenAPI specification file (openapi.yaml) to discover the URLs of temperature sensors.
// It expects the file to be in the same directory as the executable or at the specified relative path.
// It will log a fatal error if it cannot read the file or parse it correctly.
func DiscoverSensorUrls() error {
	fileData, err := os.ReadFile(appProps.APPLICATION_PROPERTIES["openapi.yaml.location"])
	if err != nil {
		log.Printf("Cannot find the openapi.yaml file for the temperature sensors: %s\n", err)
		return err
	}
	var servers types.SensorServers

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

var TakeReadingFromNamedSensor = func(sensorName string) (*types.RawSensorReading, error) {
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

	response := new(types.RawSensorReading)
	err = json.NewDecoder(resp.Body).Decode(response)

	if err != nil {
		log.Printf("Issue reading request body from sensor %s: %s\n", sensorName, err)
		return nil, err
	}
	dereferencedResponse := utils.DereferenceRawSensorReading(response)

	// insert into database
	readings := make([]types.RawSensorReading, 0)
	readings = append(readings, dereferencedResponse)
	err = database.AddListOfRawReadings(readings)
	if err != nil {
		log.Printf("Issue persisting readings to database: %s\n", err)
		return nil, err
	}
	err = smtp.SendAlertEmailIfNeeded(readings)
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
var TakeReadingsFromAllSensors = func() ([]types.RawSensorReading, error) {
	responses := make([]types.RawSensorReading, 0)
	for _, url := range sensorUrls {
		readingUrl := url + "temperature"
		resp, err := http.Get(readingUrl)
		if err != nil {
			log.Printf("Issue fetching temperature from a sensor: %s\n", err)
			continue
		}
		defer resp.Body.Close()
		response := new(types.RawSensorReading)
		err = json.NewDecoder(resp.Body).Decode(response)

		if err != nil {
			log.Printf("Issue reading request body: %s\n", err)
			continue
		}
		dereferencedResponse := utils.DereferenceRawSensorReading(response)
		responses = append(responses, dereferencedResponse)
	}
	err := database.AddListOfRawReadings(responses)
	if err != nil {
		log.Printf("Issue persisting readings to database: %s\n", err)
		return nil, err
	}
	err = smtp.SendAlertEmailIfNeeded(responses)
	if err != nil {
		log.Printf("Failed to send alerts: %s", err)
	}
	return responses, err
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
			_, err := TakeReadingsFromAllSensors()
			if err != nil {
				log.Printf("Error taking readings: %s", err)
			}

			<-ticker.C
		}
	}()
}

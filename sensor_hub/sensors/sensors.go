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

// This function takes a sensor name, fetches the temperature reading from the corresponding sensor URL,
// and returns a SensorReading object containing the sensor name and its reading.
// It logs any errors encountered during the fetching or decoding process.
// The persist parameter determines whether to save the reading to the database and send alerts.
var TakeReadingFromNamedSensor = func(sensorName string, persist bool) (*types.APIReading, error) {
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

	response := new(types.RawTemperatureReading)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Printf("Issue reading request body from sensor %s: %s\n", sensorName, err)
		return nil, err
	}
	nameTaggedResponse := utils.ConvertRawSensorReadingToAPIReading(sensorName, *response)

	// insert into database
	if !persist {
		return &nameTaggedResponse, nil
	}

	readings := make([]types.APIReading, 0)
	readings = append(readings, nameTaggedResponse)
	err = database.AddListOfRawReadings(readings)
	if err != nil {
		log.Printf("Issue persisting readings to database: %s\n", err)
		return nil, err
	}
	err = smtp.SendAlertEmailIfNeeded(readings)
	if err != nil {
		log.Printf("Failed to send alerts: %s", err)
	}

	return &nameTaggedResponse, nil
}

// This function takes a list of sensor URLs, fetches the temperature readings from each sensor,
// and returns a slice of SensorReading objects containing the sensor name and its reading.
// It logs any errors encountered during the fetching or decoding process, but continues to
// attempt to fetch readings from all sensors. If no readings are successfully collected, it returns an error.
var TakeReadingsFromAllSensors = func() ([]types.APIReading, error) {
	responses := make([]types.APIReading, 0)
	for sensorName := range sensorUrls {
		reading, err := TakeReadingFromNamedSensor(sensorName, true)
		if err != nil {
			log.Printf("Error taking reading from sensor %s: %s", sensorName, err)
			continue
		}
		responses = append(responses, *reading)
	}
	if len(responses) == 0 {
		return nil, fmt.Errorf("no readings collected from sensors")
	}

	return responses, nil
}

// This function starts a goroutine that periodically collects temperature readings from all sensors
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
			_, err := TakeReadingsFromAllSensors()
			if err != nil {
				log.Printf("Error taking periodic readings from sensors: %s", err)
			}
			<-ticker.C
		}
	}()
}

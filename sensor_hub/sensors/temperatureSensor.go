package sensors

import (
	"encoding/json"
	database "example/sensorHub/db"
	"example/sensorHub/smtp"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"fmt"
	"log"
	"net/http"
)

// TemperatureSensor implements the ISensor interface for temperature sensors.
type TemperatureSensor struct {
	name          string
	url           string
	latestReading types.APIReading
}

// ToString returns a string representation of the TemperatureSensor.
func (ts *TemperatureSensor) ToString() string {
	return fmt.Sprintf("TemperatureSensor(Name: %s, URL: %s)", ts.name, ts.url)
}

// NewTemperatureSensor creates a new TemperatureSensor instance.
func NewTemperatureSensor(name string, url string) *TemperatureSensor {
	return &TemperatureSensor{
		name: name,
		url:  url,
	}
}

// GetLatestReading returns the most recent reading taken by the sensor.
func (ts *TemperatureSensor) GetLatestReading() *types.APIReading {
	if (ts.latestReading == types.APIReading{}) {
		log.Printf("No reading taken yet for sensor %s\n", ts.name)
		return &types.APIReading{}
	}
	return &ts.latestReading
}

// GetName returns the name of the sensor.
func (ts *TemperatureSensor) GetName() string {
	return ts.name
}

// GetURL returns the URL of the sensor.
func (ts *TemperatureSensor) GetURL() string {
	return ts.url
}

// TakeReading fetches the current temperature from the sensor's API,
// updates the latestReading field, and optionally persists the reading to the database.
// If persist is true, it also checks if an alert email needs to be sent based on the reading.
func (ts *TemperatureSensor) TakeReading(persist bool) error {
	if ts.name == "" || ts.url == "" {
		return fmt.Errorf("sensor name or URL cannot be empty")
	}
	readingUrl := ts.url + "/temperature"
	resp, err := http.Get(readingUrl)
	if err != nil {
		log.Printf("Issue fetching temperature from sensor %s: %s\n", ts.name, err)
		return err
	}
	defer resp.Body.Close()
	response := new(types.RawTemperatureReading)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Printf("Issue reading request body from sensor %s: %s\n", ts.name, err)
		return err
	}
	nameTaggedResponse := utils.ConvertRawSensorReadingToAPIReading(ts.name, *response)
	log.Printf("Sensor %s reading: %v\n", ts.name, nameTaggedResponse)

	ts.latestReading = nameTaggedResponse

	// insert into database
	if !persist {
		return nil
	}

	readings := make([]types.APIReading, 0)
	readings = append(readings, nameTaggedResponse)
	err = database.AddListOfRawReadings(readings)
	if err != nil {
		log.Printf("Issue persisting readings to database: %s\n", err)
		return nil
	}
	err = smtp.SendAlertEmailIfNeeded(readings)
	if err != nil {
		log.Printf("Failed to send alerts: %s", err)
	}

	return nil
}

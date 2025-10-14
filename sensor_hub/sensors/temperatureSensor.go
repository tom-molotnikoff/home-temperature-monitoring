package sensors

import (
	"encoding/json"
	database "example/sensorHub/db"
	"example/sensorHub/smtp"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"fmt"
	"io"
	"log"
	"net/http"
)

type TemperatureSensor struct {
	name          string
	url           string
	latestReading types.APITempReading
	repo          database.Repository[types.DbTempReading]
}

func (ts *TemperatureSensor) ToString() string {
	return fmt.Sprintf("TemperatureSensor(Name: %s, URL: %s)", ts.name, ts.url)
}

func NewTemperatureSensor(name string, url string, repo database.Repository[types.DbTempReading]) *TemperatureSensor {
	return &TemperatureSensor{
		name: name,
		url:  url,
		repo: repo,
	}
}

func (ts *TemperatureSensor) GetLatestReading() *types.APITempReading {
	if (ts.latestReading == types.APITempReading{}) {
		log.Printf("No reading taken yet for sensor %s\n", ts.name)
		return &types.APITempReading{}
	}
	return &ts.latestReading
}

func (ts *TemperatureSensor) GetName() string {
	return ts.name
}

func (ts *TemperatureSensor) GetURL() string {
	return ts.url
}

func (ts *TemperatureSensor) TakeReading(persist bool) error {
	if ts.name == "" || ts.url == "" {
		return fmt.Errorf("sensor name or URL cannot be empty")
	}
	readingUrl := ts.url + "/temperature"
	resp, err := http.Get(readingUrl)
	if err != nil {
		return fmt.Errorf("issue fetching temperature from sensor %s: %w", ts.name, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Issue closing response body: %v", err)
		}
	}(resp.Body)
	response := new(types.RawTempReading)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("issue reading request body from sensor %s: %w", ts.name, err)
	}
	nameTaggedResponse := utils.ConvertRawSensorReadingToDbReading(ts.name, *response)
	log.Printf("Sensor %s reading: %v\n", ts.name, nameTaggedResponse)

	ts.latestReading = utils.ConvertDbReadingToApiReading(nameTaggedResponse)
	readings := make([]types.DbTempReading, 0)
	readings = append(readings, nameTaggedResponse)

	if persist {
		err = ts.repo.Add(readings)
		if err != nil {
			log.Printf("Issue persisting readings to database: %v", err)
		}
	}

	err = smtp.SendAlertEmailIfNeeded(utils.ConvertDbReadingsToApiReadings(readings))
	if err != nil {
		log.Printf("Failed to send alerts: %v", err)
	}

	return nil
}

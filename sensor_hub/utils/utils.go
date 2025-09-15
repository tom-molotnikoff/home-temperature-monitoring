package utils

import (
	"bufio"
	"example/sensorHub/types"
	"log"
	"os"
	"strings"
)

// This function reads a properties file and returns a map of key-value pairs
// It expects the properties file to be in the format: key=value
// It will log a fatal error if it cannot read the file or parse it correctly.
var ReadPropertiesFile = func(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to read %s: %s", path, err)
		return nil, err
	}
	defer file.Close()

	props := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			props[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Failed to read %s: %s", path, err)
		return nil, err
	}
	return props, nil
}

// This function converts a slice of DbReading objects to a slice of APIReading objects.
// It rounds the temperature to one decimal place for consistency in the API response.
var ConvertDbReadingsToApiReadings = func(dbReadings []types.DbReading) []types.APIReading {
	var apiReadings []types.APIReading
	for _, r := range dbReadings {
		apiReadings = append(apiReadings, types.APIReading{
			SensorName: r.SensorName,
			Reading: struct {
				Temperature float64 `json:"temperature"`
				Time        string  `json:"time"`
			}{
				Temperature: r.Temperature,
				Time:        r.Time,
			},
		})
	}
	return apiReadings
}

// This function converts a slice of APIReading objects to a slice of DbReading objects.
// It extracts the sensor name, temperature, and time from each APIReading.
var ConvertAPIReadingsToDbReadings = func(raw []types.APIReading) []types.DbReading {
	var readings []types.DbReading
	for _, r := range raw {
		var reading types.DbReading
		reading.SensorName = r.SensorName
		reading.Temperature = r.Reading.Temperature
		reading.Time = r.Reading.Time
		readings = append(readings, reading)
	}
	return readings
}

// This function converts a RawSensorReading and a sensor name into an APIReading.
// It combines the sensor name with the reading from the RawSensorReading.
var ConvertRawSensorReadingToAPIReading = func(name string, reading types.RawTemperatureReading) types.APIReading {
	return types.APIReading{
		SensorName: name,
		Reading:    reading,
	}
}

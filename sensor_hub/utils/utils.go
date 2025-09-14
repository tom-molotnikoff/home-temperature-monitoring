package utils

import (
	"bufio"
	"example/sensorHub/types"
	"log"
	"math"
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

var ConvertDbReadingsToApiReadings = func(dbReadings []types.DbReading) []types.APIReading {
	var apiReadings []types.APIReading
	for _, r := range dbReadings {
		apiReadings = append(apiReadings, types.APIReading{
			SensorName: r.SensorName,
			Reading: struct {
				Temperature float64 `json:"temperature"`
				Time        string  `json:"time"`
			}{
				Temperature: math.Round(r.Temperature*10) / 10,
				Time:        r.Time,
			},
		})
	}
	return apiReadings
}

var ConvertRawSensorReadingsToDbReadings = func(raw []types.RawSensorReading) []types.DbReading {
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

var DereferenceRawSensorReading = func(r *types.RawSensorReading) types.RawSensorReading {
	if r == nil {
		return types.RawSensorReading{}
	}
	return types.RawSensorReading{
		SensorName: r.SensorName,
		Reading: struct {
			Temperature float64 `json:"temperature"`
			Time        string  `json:"time"`
		}{
			Temperature: r.Reading.Temperature,
			Time:        r.Reading.Time,
		},
	}
}

package utils

import (
	"bufio"
	"example/sensorHub/types"
	"fmt"
	"os"
	"strings"
)

var ReadPropertiesFile = func(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open properties file: %w", err)
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
		return nil, fmt.Errorf("failed to read properties file: %w", err)
	}
	return props, nil
}

var ConvertDbReadingsToApiReadings = func(dbReadings []types.DbTempReading) []types.APITempReading {
	var apiReadings []types.APITempReading
	for _, r := range dbReadings {
		apiReadings = append(apiReadings, types.APITempReading{
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

var ConvertRawSensorReadingToDbReading = func(name string, raw types.RawTempReading) types.DbTempReading {
	return types.DbTempReading{
		SensorName:  name,
		Temperature: raw.Temperature,
		Time:        raw.Time,
	}
}

var ConvertDbReadingToApiReading = func(dbReading types.DbTempReading) types.APITempReading {
	return types.APITempReading{
		SensorName: dbReading.SensorName,
		Reading: types.RawTempReading{
			Temperature: dbReading.Temperature,
			Time:        dbReading.Time,
		},
	}
}

var ConvertAPIReadingsToDbReadings = func(raw []types.APITempReading) []types.DbTempReading {
	var readings []types.DbTempReading
	for _, r := range raw {
		var reading types.DbTempReading
		reading.SensorName = r.SensorName
		reading.Temperature = r.Reading.Temperature
		reading.Time = r.Reading.Time
		readings = append(readings, reading)
	}
	return readings
}

var ConvertRawSensorReadingToAPIReading = func(name string, reading types.RawTempReading) types.APITempReading {
	return types.APITempReading{
		SensorName: name,
		Reading:    reading,
	}
}

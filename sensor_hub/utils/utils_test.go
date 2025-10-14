package utils

import (
	"example/sensorHub/types"
	"testing"
)

func TestReadPropertiesFile(t *testing.T) {
	// Test with a valid properties file
	props, err := ReadPropertiesFile("testdata/valid.properties")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if props["key1"] != "value1" || props["key2"] != "value2" {
		t.Errorf("unexpected properties: %v", props)
	}

	// Test with a non-existent file
	_, err = ReadPropertiesFile("testdata/nonexistent.properties")
	if err == nil {
		t.Fatalf("expected error for non-existent file, got nil")
	}

	// Test with an invalid format file
	props, err = ReadPropertiesFile("testdata/invalid.properties")
	if err != nil {
		t.Fatalf("expected no error for invalid format, got %v", err)
	}
	if len(props) != 0 {
		t.Errorf("expected empty properties for invalid format, got %v", props)
	}
}

func TestConvertDbReadingsToApiReadings(t *testing.T) {
	dbReadings := []types.DbTempReading{
		{SensorName: "sensor1", Temperature: 23.456, Time: "2024-01-01T12:00:00Z"},
		{SensorName: "sensor2", Temperature: 24.567, Time: "2024-01-01T12:05:00Z"},
	}
	apiReadings := ConvertDbReadingsToApiReadings(dbReadings)

	if len(apiReadings) != len(dbReadings) {
		t.Fatalf("expected %d api readings, got %d", len(dbReadings), len(apiReadings))
	}

	for i, r := range apiReadings {
		if r.SensorName != dbReadings[i].SensorName {
			t.Errorf("expected sensor name %s, got %s", dbReadings[i].SensorName, r.SensorName)
		}
		if r.Reading.Temperature != dbReadings[i].Temperature {
			t.Errorf("expected temperature %f, got %f", dbReadings[i].Temperature, r.Reading.Temperature)
		}
		if r.Reading.Time != dbReadings[i].Time {
			t.Errorf("expected time %s, got %s", dbReadings[i].Time, r.Reading.Time)
		}
	}
}

func TestConvertAPIReadingsToDbReadings(t *testing.T) {
	apiReadings := []types.APITempReading{
		{SensorName: "sensor1", Reading: struct {
			Temperature float64 `json:"temperature"`
			Time        string  `json:"time"`
		}{Temperature: 23.456, Time: "2024-01-01T12:00:00Z"}},
		{SensorName: "sensor2", Reading: struct {
			Temperature float64 `json:"temperature"`
			Time        string  `json:"time"`
		}{Temperature: 24.567, Time: "2024-01-01T12:05:00Z"}},
	}
	dbReadings := ConvertAPIReadingsToDbReadings(apiReadings)

	if len(dbReadings) != len(apiReadings) {
		t.Fatalf("expected %d db readings, got %d", len(apiReadings), len(dbReadings))
	}
	for i, r := range dbReadings {
		if r.SensorName != apiReadings[i].SensorName {
			t.Errorf("expected sensor name %s, got %s", apiReadings[i].SensorName, r.SensorName)
		}
		if r.Temperature != apiReadings[i].Reading.Temperature {
			t.Errorf("expected temperature %f, got %f", apiReadings[i].Reading.Temperature, r.Temperature)
		}
		if r.Time != apiReadings[i].Reading.Time {
			t.Errorf("expected time %s, got %s", apiReadings[i].Reading.Time, r.Time)
		}
	}
}

func TestConvertRawSensorReadingToAPIReading(t *testing.T) {
	name := "sensor1"
	rawReading := types.RawTempReading{
		Temperature: 23.456,
		Time:        "2024-01-01T12:00:00Z",
	}
	apiReading := ConvertRawSensorReadingToAPIReading(name, rawReading)

	if apiReading.SensorName != name {
		t.Errorf("expected sensor name %s, got %s", name, apiReading.SensorName)
	}
	if apiReading.Reading.Temperature != rawReading.Temperature {
		t.Errorf("expected temperature %f, got %f", rawReading.Temperature, apiReading.Reading.Temperature)
	}
	if apiReading.Reading.Time != rawReading.Time {
		t.Errorf("expected time %s, got %s", rawReading.Time, apiReading.Reading.Time)
	}
}

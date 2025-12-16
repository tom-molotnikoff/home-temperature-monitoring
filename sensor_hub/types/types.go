package types

type SensorServers struct {
	Servers []SensorServerItem `yaml:"servers"`
}

type SensorServerItem struct {
	Url       string                                  `yaml:"url"`
	Variables map[string]SensorServerVariableProperty `yaml:"variables"`
}

type SensorServerVariableProperty struct {
	Default string `yaml:"default"`
}

type Sensor struct {
	Id           int                `json:"id"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	URL          string             `json:"url"`
	HealthStatus SensorHealthStatus `json:"health_status"`
	HealthReason string             `json:"health_reason"`
	Enabled      bool               `json:"enabled"`
}

const (
	TableTemperatureReadings      = "temperature_readings"
	TableHourlyAverageTemperature = "hourly_avg_temperature"
	TableSensorHealthHistory      = "sensor_health_history"
)

type SensorHealthStatus string

const (
	SensorBadHealth     SensorHealthStatus = "bad"
	SensorGoodHealth    SensorHealthStatus = "good"
	SensorUnknownHealth SensorHealthStatus = "unknown"
)

type RawTempReading struct {
	Temperature float64 `json:"temperature"`
	Time        string  `json:"time"`
}

type TemperatureReading struct {
	Id          int     `json:"id"`
	SensorName  string  `json:"sensor_name"`
	Time        string  `json:"time"`
	Temperature float64 `json:"temperature"`
}

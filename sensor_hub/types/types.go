package types

import "time"

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

type SensorHealthHistory struct {
	Id           int                `json:"id"`
	SensorId     string             `json:"sensor_id"`
	HealthStatus SensorHealthStatus `json:"health_status"`
	RecordedAt   time.Time          `json:"recorded_at"`
}

type TemperatureReading struct {
	Id          int     `json:"id"`
	SensorName  string  `json:"sensor_name"`
	Time        string  `json:"time"`
	Temperature float64 `json:"temperature"`
}

type User struct {
	Id                 int       `json:"id"`
	Username           string    `json:"username"`
	Email              string    `json:"email"`
	Disabled           bool      `json:"disabled"`
	MustChangePassword bool      `json:"must_change_password"`
	Roles              []string  `json:"roles"`
	Permissions        []string  `json:"permissions"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Predefined role names - need to try and remove hardcoding these, we don't really need them
const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleViewer = "viewer"
)

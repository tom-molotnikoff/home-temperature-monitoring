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
	Id           int               `json:"id"`
	Name         string            `json:"name"`
	SensorDriver string            `json:"sensor_driver"`
	Config       map[string]string `json:"config"`
	HealthStatus SensorHealthStatus `json:"health_status"`
	HealthReason string            `json:"health_reason"`
	Enabled      bool              `json:"enabled"`
}

const (
	TableReadings             = "readings"
	TableHourlyAverages       = "hourly_averages"
	TableHourlyEvents         = "hourly_events"
	TableSensorHealthHistory  = "sensor_health_history"
	TableMeasurementTypes     = "measurement_types"
	TableSensorMeasurementTypes = "sensor_measurement_types"
)

type SensorHealthStatus string

const (
	SensorBadHealth     SensorHealthStatus = "bad"
	SensorGoodHealth    SensorHealthStatus = "good"
	SensorUnknownHealth SensorHealthStatus = "unknown"
)

type SensorHealthHistory struct {
	Id           int                `json:"id"`
	SensorId     string             `json:"sensor_id"`
	HealthStatus SensorHealthStatus `json:"health_status"`
	RecordedAt   time.Time          `json:"recorded_at"`
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

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
	Status       SensorStatus      `json:"status"`
}

type SensorStatus string

const (
	SensorStatusActive    SensorStatus = "active"
	SensorStatusPending   SensorStatus = "pending"
	SensorStatusDismissed SensorStatus = "dismissed"
)

const (
	TableReadings               = "readings"
	TableHourlyAverages         = "hourly_averages"
	TableHourlyEvents           = "hourly_events"
	TableSensorHealthHistory    = "sensor_health_history"
	TableMeasurementTypes       = "measurement_types"
	TableSensorMeasurementTypes = "sensor_measurement_types"
	TableMQTTBrokers            = "mqtt_brokers"
	TableMQTTSubscriptions      = "mqtt_subscriptions"
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

// MQTTBroker represents an MQTT broker connection configuration.
type MQTTBroker struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	Username       string    `json:"username,omitempty"`
	Password       string    `json:"password,omitempty"`
	ClientId       string    `json:"client_id,omitempty"`
	CACertPath     string    `json:"ca_cert_path,omitempty"`
	ClientCertPath string    `json:"client_cert_path,omitempty"`
	ClientKeyPath  string    `json:"client_key_path,omitempty"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MQTTSubscription represents an MQTT topic subscription bound to a driver.
type MQTTSubscription struct {
	Id           int       `json:"id"`
	BrokerId     int       `json:"broker_id"`
	TopicPattern string    `json:"topic_pattern"`
	DriverType   string    `json:"driver_type"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
